package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// --- Stub gọn: embed interface để chỉ override method cần, phần còn lại không gọi ---

type stubKBService struct {
	interfaces.KnowledgeBaseService
	kbs []*types.KnowledgeBase
}

func (s stubKBService) ListKnowledgeBases(ctx context.Context) ([]*types.KnowledgeBase, error) {
	return s.kbs, nil
}

type stubKnowService struct {
	interfaces.KnowledgeService
	byKB map[string][]*types.Knowledge
}

func (s stubKnowService) ListKnowledgeByKnowledgeBaseID(ctx context.Context, kbID string) ([]*types.Knowledge, error) {
	return s.byKB[kbID], nil
}

type stubIdxRepo struct {
	interfaces.ChessKBIndexRepository
}

func newTestIndexer(kbs []*types.KnowledgeBase, byKB map[string][]*types.Knowledge) *ChessKnowledgeIndexer {
	return NewChessKnowledgeIndexer(
		stubKBService{kbs: kbs},
		stubKnowService{byKB: byKB},
		stubIdxRepo{},
	)
}

func TestChessIndexStatus_CountsByParseStatus(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "true")
	kb := &types.KnowledgeBase{
		ID: "kb-chess", Name: chessKBName, EmbeddingModelID: "emb-1",
		IndexingStrategy: types.DefaultIndexingStrategy(), // vector+keyword bật
	}
	ks := []*types.Knowledge{
		{ParseStatus: types.ParseStatusCompleted, EnableStatus: "enabled"},
		{ParseStatus: types.ParseStatusCompleted, EnableStatus: "enabled"},
		{ParseStatus: types.ParseStatusPending, EnableStatus: "disabled"},
		{ParseStatus: types.ParseStatusFailed, ErrorMessage: "embed lỗi mẫu", EnableStatus: "disabled"},
	}
	ix := newTestIndexer([]*types.KnowledgeBase{kb}, map[string][]*types.Knowledge{"kb-chess": ks})

	st, err := ix.IndexStatus(context.Background())
	if err != nil {
		t.Fatalf("IndexStatus lỗi: %v", err)
	}
	if !st.Enabled || !st.KBExists || !st.EmbeddingConfigured {
		t.Fatalf("kỳ vọng enabled/kb_exists/embedding_configured = true, nhận %+v", st)
	}
	if !st.VectorEnabled || !st.KeywordEnabled || !st.Searchable {
		t.Errorf("kỳ vọng vector/keyword/searchable = true, nhận %+v", st)
	}
	if st.KBID != "kb-chess" || st.EmbeddingModelID != "emb-1" {
		t.Errorf("KBID/EmbeddingModelID sai: %+v", st)
	}
	if st.Total != 4 || st.Completed != 2 || st.Pending != 1 || st.Failed != 1 {
		t.Errorf("đếm sai: total=%d completed=%d pending=%d failed=%d", st.Total, st.Completed, st.Pending, st.Failed)
	}
	if st.EnabledDocs != 2 || st.DisabledDocs != 2 {
		t.Errorf("đếm enable_status sai: enabled=%d disabled=%d", st.EnabledDocs, st.DisabledDocs)
	}
	if st.SampleError != "embed lỗi mẫu" {
		t.Errorf("SampleError sai: %q", st.SampleError)
	}
}

func TestChessIndexStatus_NotSearchableWhenIndexingOff(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "true")
	// KB cờ có embedding model nhưng TẮT vector+keyword → bị loại khỏi search.
	kb := &types.KnowledgeBase{ID: "kb-chess", Name: chessKBName, EmbeddingModelID: "emb-1"}
	ix := newTestIndexer([]*types.KnowledgeBase{kb}, map[string][]*types.Knowledge{})

	st, err := ix.IndexStatus(context.Background())
	if err != nil {
		t.Fatalf("IndexStatus lỗi: %v", err)
	}
	if st.Searchable || st.VectorEnabled || st.KeywordEnabled {
		t.Errorf("kỳ vọng searchable=false khi indexing off, nhận %+v", st)
	}
}

func TestChessIndexStatus_NoChessKB(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "true")
	other := &types.KnowledgeBase{ID: "kb-x", Name: "KB khác", EmbeddingModelID: "emb-1"}
	ix := newTestIndexer([]*types.KnowledgeBase{other}, nil)

	st, err := ix.IndexStatus(context.Background())
	if err != nil {
		t.Fatalf("IndexStatus lỗi: %v", err)
	}
	if st.KBExists {
		t.Errorf("không có KB cờ → kb_exists phải false, nhận %+v", st)
	}
}

func TestChessIndexStatus_EmbeddingNotConfigured(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "true")
	// KB cờ tồn tại nhưng KHÔNG có embedding model → nguyên nhân gốc RAG rỗng.
	kb := &types.KnowledgeBase{ID: "kb-chess", Name: chessKBName, EmbeddingModelID: ""}
	ix := newTestIndexer([]*types.KnowledgeBase{kb}, map[string][]*types.Knowledge{})

	st, err := ix.IndexStatus(context.Background())
	if err != nil {
		t.Fatalf("IndexStatus lỗi: %v", err)
	}
	if !st.KBExists || st.EmbeddingConfigured {
		t.Errorf("kỳ vọng kb_exists=true, embedding_configured=false, nhận %+v", st)
	}
}

func TestReindexAll_FailsWhenNoEmbeddingTemplate(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "true")
	// Không có KB cờ và không KB nào có embedding model → không thể tạo KB cờ.
	noEmbed := &types.KnowledgeBase{ID: "kb-x", Name: "KB rỗng", EmbeddingModelID: ""}
	ix := newTestIndexer([]*types.KnowledgeBase{noEmbed}, nil)
	svc := &chessLibraryService{indexer: ix}

	_, err := svc.ReindexAll(context.Background(), 1)
	if err == nil {
		t.Fatal("kỳ vọng ReindexAll trả lỗi khi không có embedding template, nhận nil")
	}
	if !strings.Contains(err.Error(), "index") && !strings.Contains(err.Error(), "embedding") {
		t.Errorf("thông báo lỗi nên nhắc embedding/index, nhận: %v", err)
	}
}

func TestReindexAll_FailsWhenChessKBHasNoEmbedding(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "true")
	// KB cờ đã tồn tại nhưng thiếu embedding model → chặn fail-loud, không success giả.
	kb := &types.KnowledgeBase{ID: "kb-chess", Name: chessKBName, EmbeddingModelID: ""}
	ix := newTestIndexer([]*types.KnowledgeBase{kb}, nil)
	svc := &chessLibraryService{indexer: ix}

	_, err := svc.ReindexAll(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "embedding") {
		t.Fatalf("kỳ vọng lỗi nhắc 'embedding', nhận: %v", err)
	}
}

// --- Self-heal: mapping mồ côi (knowledge bị xóa cùng KB) → tạo mới, không kẹt ---

type healStubKB struct {
	interfaces.KnowledgeBaseService
	kb *types.KnowledgeBase
}

func (s healStubKB) ListKnowledgeBases(ctx context.Context) ([]*types.KnowledgeBase, error) {
	return []*types.KnowledgeBase{s.kb}, nil
}

type healStubKnow struct {
	interfaces.KnowledgeService
	getErr  error
	created int
}

func (s *healStubKnow) GetKnowledgeByID(ctx context.Context, id string) (*types.Knowledge, error) {
	return nil, s.getErr // mô phỏng "knowledge not found" sau khi KB bị xóa
}

func (s *healStubKnow) CreateKnowledgeFromManual(ctx context.Context, kbID string, p *types.ManualKnowledgePayload, ch string) (*types.Knowledge, error) {
	s.created++
	return &types.Knowledge{ID: "new-k"}, nil
}

type healStubIdx struct {
	interfaces.ChessKBIndexRepository
	mapping  *types.ChessKBIndex
	deleted  int
	upserted int
}

func (s *healStubIdx) Get(ctx context.Context, tenantID uint64, ct, slug string) (*types.ChessKBIndex, error) {
	return s.mapping, nil
}
func (s *healStubIdx) Delete(ctx context.Context, tenantID uint64, ct, slug string) error {
	s.deleted++
	return nil
}
func (s *healStubIdx) Upsert(ctx context.Context, m *types.ChessKBIndex) error {
	s.upserted++
	return nil
}

func TestUpsert_OrphanMappingSelfHeals(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "true")
	kb := &types.KnowledgeBase{
		ID: "kb-chess", Name: chessKBName, EmbeddingModelID: "emb-1",
		IndexingStrategy: types.DefaultIndexingStrategy(),
	}
	knu := &healStubKnow{getErr: fmt.Errorf("knowledge not found")}
	idx := &healStubIdx{mapping: &types.ChessKBIndex{KnowledgeID: "dead-k"}}
	ix := NewChessKnowledgeIndexer(healStubKB{kb: kb}, knu, idx)

	err := ix.IndexGame(context.Background(), &types.ChessGame{Slug: "van-mau", TenantID: 1})
	if err != nil {
		t.Fatalf("IndexGame nên tự lành mapping mồ côi, nhận lỗi: %v", err)
	}
	if idx.deleted != 1 {
		t.Errorf("kỳ vọng gỡ 1 mapping mồ côi, nhận %d", idx.deleted)
	}
	if knu.created != 1 {
		t.Errorf("kỳ vọng tạo 1 knowledge mới, nhận %d", knu.created)
	}
	if idx.upserted != 1 {
		t.Errorf("kỳ vọng ghi 1 mapping mới, nhận %d", idx.upserted)
	}
}

func TestReindexAll_DisabledGate(t *testing.T) {
	t.Setenv("CHESS_KB_INDEX", "false")
	ix := newTestIndexer(nil, nil)
	svc := &chessLibraryService{indexer: ix}
	if _, err := svc.ReindexAll(context.Background(), 1); err == nil {
		t.Fatal("gate tắt → ReindexAll phải trả lỗi 'chưa bật'")
	}
}
