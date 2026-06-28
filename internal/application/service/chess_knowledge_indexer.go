package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// chessKBName là tên KB chuyên dụng chứa tri thức cờ (tự tạo mỗi tenant).
const chessKBName = "Tri thức cờ vua"

// ChessKnowledgeIndexer đồng bộ đối tượng cờ (ván/thế cờ/bài giảng) thành bản ghi
// Knowledge trong "KB tri thức cờ vua" để trợ lý/HLV truy hồi qua RAG.
//
// AN TOÀN:
//   - Gate sau env CHESS_KB_INDEX (mặc định TẮT). Tắt → mọi hàm là no-op.
//   - Mọi thao tác BEST-EFFORT: lỗi chỉ log, KHÔNG bao giờ chặn CRUD đối tượng cờ.
//   - KB cờ được tạo bằng cách SAO CHÉP cấu hình embedding của một KB sẵn có của
//     tenant (cần ≥1 KB đã cấu hình model). Chưa có → bỏ qua index (log).
//
// LƯU Ý: tính năng này cần verify trên full stack (model embedding + vector store
// + worker async) trước khi bật thật.
type ChessKnowledgeIndexer struct {
	kbService        interfaces.KnowledgeBaseService
	knowledgeService interfaces.KnowledgeService
	idxRepo          interfaces.ChessKBIndexRepository
}

// NewChessKnowledgeIndexer tạo indexer tri thức cờ.
func NewChessKnowledgeIndexer(
	kbService interfaces.KnowledgeBaseService,
	knowledgeService interfaces.KnowledgeService,
	idxRepo interfaces.ChessKBIndexRepository,
) *ChessKnowledgeIndexer {
	return &ChessKnowledgeIndexer{kbService: kbService, knowledgeService: knowledgeService, idxRepo: idxRepo}
}

// Enabled cho biết có bật index tri thức cờ không (env CHESS_KB_INDEX truthy).
func (ix *ChessKnowledgeIndexer) Enabled() bool {
	if ix == nil || ix.kbService == nil || ix.knowledgeService == nil || ix.idxRepo == nil {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(os.Getenv("CHESS_KB_INDEX")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// IndexGame / IndexPuzzle / IndexLesson đồng bộ một đối tượng cờ (tạo/cập nhật).
func (ix *ChessKnowledgeIndexer) IndexGame(ctx context.Context, g *types.ChessGame) {
	if !ix.Enabled() || g == nil || g.Slug == "" {
		return
	}
	title, content := buildGameKnowledgeText(g)
	ix.upsert(ctx, g.TenantID, types.ChessRefTypeGame, g.Slug, title, content)
}

func (ix *ChessKnowledgeIndexer) IndexPuzzle(ctx context.Context, p *types.ChessPuzzle) {
	if !ix.Enabled() || p == nil || p.Slug == "" {
		return
	}
	title, content := buildPuzzleKnowledgeText(p)
	ix.upsert(ctx, p.TenantID, types.ChessRefTypePuzzle, p.Slug, title, content)
}

func (ix *ChessKnowledgeIndexer) IndexLesson(ctx context.Context, l *types.ChessLesson) {
	if !ix.Enabled() || l == nil || l.Slug == "" {
		return
	}
	title, content := buildLessonKnowledgeText(l)
	ix.upsert(ctx, l.TenantID, types.ChessRefTypeLesson, l.Slug, title, content)
}

// Remove xóa bản ghi Knowledge tương ứng (khi đối tượng cờ bị xóa).
func (ix *ChessKnowledgeIndexer) Remove(ctx context.Context, tenantID uint64, chessType, slug string) {
	if !ix.Enabled() || slug == "" {
		return
	}
	existing, err := ix.idxRepo.Get(ctx, tenantID, chessType, slug)
	if err != nil || existing == nil {
		return
	}
	if existing.KnowledgeID != "" {
		if err := ix.knowledgeService.DeleteKnowledge(ctx, existing.KnowledgeID); err != nil {
			logger.Warnf(ctx, "chess index: xóa knowledge %s thất bại: %v", existing.KnowledgeID, err)
		}
	}
	if err := ix.idxRepo.Delete(ctx, tenantID, chessType, slug); err != nil {
		logger.Warnf(ctx, "chess index: xóa mapping %s/%s thất bại: %v", chessType, slug, err)
	}
}

// upsert tạo mới hoặc cập nhật bản ghi Knowledge cho một đối tượng cờ (best-effort).
func (ix *ChessKnowledgeIndexer) upsert(ctx context.Context, tenantID uint64, chessType, slug, title, content string) {
	payload := &types.ManualKnowledgePayload{
		Title:   title,
		Content: content,
		Status:  types.ManualKnowledgeStatusPublish,
		Channel: "chess",
	}
	existing, _ := ix.idxRepo.Get(ctx, tenantID, chessType, slug)
	if existing != nil && existing.KnowledgeID != "" {
		if _, err := ix.knowledgeService.UpdateManualKnowledge(ctx, existing.KnowledgeID, payload); err != nil {
			logger.Warnf(ctx, "chess index: cập nhật knowledge cho %s/%s thất bại: %v", chessType, slug, err)
		}
		return
	}
	kb, err := ix.ensureChessKB(ctx)
	if err != nil || kb == nil {
		logger.Warnf(ctx, "chess index: không có KB cờ để index %s/%s: %v", chessType, slug, err)
		return
	}
	k, err := ix.knowledgeService.CreateKnowledgeFromManual(ctx, kb.ID, payload, "chess")
	if err != nil || k == nil {
		logger.Warnf(ctx, "chess index: tạo knowledge cho %s/%s thất bại: %v", chessType, slug, err)
		return
	}
	if err := ix.idxRepo.Upsert(ctx, &types.ChessKBIndex{
		TenantID: tenantID, ChessType: chessType, ChessSlug: slug, KnowledgeID: k.ID, KBID: kb.ID,
	}); err != nil {
		logger.Warnf(ctx, "chess index: lưu mapping %s/%s thất bại: %v", chessType, slug, err)
	}
}

// ensureChessKB tìm KB cờ của tenant; chưa có thì tạo bằng cách SAO CHÉP cấu hình
// embedding/vector store của một KB sẵn có (cần ≥1 KB đã cấu hình model).
func (ix *ChessKnowledgeIndexer) ensureChessKB(ctx context.Context) (*types.KnowledgeBase, error) {
	kbs, err := ix.kbService.ListKnowledgeBases(ctx)
	if err != nil {
		return nil, err
	}
	var tpl *types.KnowledgeBase
	for _, kb := range kbs {
		if kb.Name == chessKBName {
			return kb, nil
		}
		if tpl == nil && kb.EmbeddingModelID != "" {
			tpl = kb
		}
	}
	if tpl == nil {
		return nil, fmt.Errorf("chưa có KB cấu hình embedding để sao chép")
	}
	// Chỉ sao chép các trường cần cho embedding/lưu trữ; bỏ wiki/faq/extract để KB
	// cờ là KB tài liệu thuần (EnsureDefaults sẽ điền phần còn lại).
	nk := &types.KnowledgeBase{
		Name:                  chessKBName,
		Type:                  "document",
		Description:           "Kho tri thức tự động từ thư viện cờ vua (ván/thế cờ/bài giảng) để trợ lý truy hồi.",
		ChunkingConfig:        tpl.ChunkingConfig,
		ImageProcessingConfig: tpl.ImageProcessingConfig,
		EmbeddingModelID:      tpl.EmbeddingModelID,
		SummaryModelID:        tpl.SummaryModelID,
		VLMConfig:             tpl.VLMConfig,
		ASRConfig:             tpl.ASRConfig,
		StorageProviderConfig: tpl.StorageProviderConfig,
		VectorStoreID:         tpl.VectorStoreID,
	}
	return ix.kbService.CreateKnowledgeBase(ctx, nk)
}
