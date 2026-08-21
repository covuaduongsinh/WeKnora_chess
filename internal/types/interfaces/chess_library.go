package interfaces

import (
	"context"
	"io"

	"github.com/Tencent/WeKnora/internal/types"
)

// ChessLibraryService định nghĩa nghiệp vụ kho ván đấu & ngân hàng bài tập cờ vua.
type ChessLibraryService interface {
	// ---- Ván đấu ----
	ListGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) ([]*types.ChessGame, error)
	GetGame(ctx context.Context, tenantID uint64, id string) (*types.ChessGame, error)
	// GetGameBySlug giải mã wikilink [[game/<slug>]] về ván cờ.
	GetGameBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessGame, error)
	// GetGameBacklinks liệt kê trang wiki/bài giảng trỏ tới ván cờ này.
	GetGameBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreateGame(ctx context.Context, game *types.ChessGame) (*types.ChessGame, error)
	UpdateGame(ctx context.Context, game *types.ChessGame) (*types.ChessGame, error)
	// RenameGameSlug đổi slug ván sang newSlug (chuẩn hóa + đảm bảo duy nhất) và ghi
	// alias slug-cũ→mới để wikilink cũ vẫn sống.
	RenameGameSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessGame, error)
	DeleteGame(ctx context.Context, tenantID uint64, id string) error
	// ImportGames tách một PGN nhiều ván và tạo nhiều ChessGame; trả số ván đã thêm.
	ImportGames(ctx context.Context, tenantID uint64, pgn string) (int, error)

	// ---- Bài tập ----
	ListPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]*types.ChessPuzzle, error)
	GetPuzzle(ctx context.Context, tenantID uint64, id string) (*types.ChessPuzzle, error)
	// GetPuzzleBySlug giải mã wikilink [[puzzle/<slug>]] về thế cờ/bài tập.
	GetPuzzleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessPuzzle, error)
	// GetPuzzleBacklinks liệt kê trang wiki/bài giảng trỏ tới thế cờ này.
	GetPuzzleBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) (*types.ChessPuzzle, error)
	UpdatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) (*types.ChessPuzzle, error)
	// RenamePuzzleSlug đổi slug bài tập sang newSlug + ghi alias slug-cũ→mới.
	RenamePuzzleSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessPuzzle, error)
	DeletePuzzle(ctx context.Context, tenantID uint64, id string) error
	RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error)

	// ---- Thế cờ (Ngân hàng thế cờ) ----
	ListPositions(ctx context.Context, tenantID uint64, f types.ChessPositionFilter) ([]*types.ChessPosition, error)
	GetPosition(ctx context.Context, tenantID uint64, id string) (*types.ChessPosition, error)
	// GetPositionBySlug giải mã wikilink [[position/<slug>]] về thế cờ.
	GetPositionBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessPosition, error)
	// GetPositionBacklinks liệt kê trang wiki/bài giảng/thế cờ khác trỏ tới thế cờ này.
	GetPositionBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreatePosition(ctx context.Context, position *types.ChessPosition) (*types.ChessPosition, error)
	UpdatePosition(ctx context.Context, position *types.ChessPosition) (*types.ChessPosition, error)
	// RenamePositionSlug đổi slug thế cờ sang newSlug + ghi alias slug-cũ→mới.
	RenamePositionSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessPosition, error)
	DeletePosition(ctx context.Context, tenantID uint64, id string) error
	// ListPositionsByGame liệt kê các thế cờ đã trích từ MỘT ván cụ thể.
	ListPositionsByGame(ctx context.Context, tenantID uint64, gameID string) ([]*types.ChessPosition, error)

	// ---- Bài viết (Ngân hàng bài viết) ----
	ListArticles(ctx context.Context, tenantID uint64, f types.ChessArticleFilter) ([]*types.ChessArticle, error)
	// SearchArticles tìm bài viết theo từ khóa (slug/title/aliases/summary/
	// content) toàn tenant — autocomplete wikilink.
	SearchArticles(ctx context.Context, tenantID uint64, keyword string, limit int) ([]*types.ChessArticle, error)
	GetArticle(ctx context.Context, tenantID uint64, id string) (*types.ChessArticle, error)
	// GetArticleBySlug giải mã wikilink [[article/<slug>]] về bài viết.
	GetArticleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessArticle, error)
	// GetArticleBacklinks liệt kê trang wiki/bài giảng/thế cờ/chương/bài viết
	// khác trỏ tới bài viết này.
	GetArticleBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreateArticle(ctx context.Context, article *types.ChessArticle) (*types.ChessArticle, error)
	// UpdateArticle cập nhật bài viết; nếu Title/Content thực sự đổi thì lưu
	// bản TRƯỚC khi ghi đè vào lịch sử phiên bản (revisionNote là ghi chú
	// thay đổi, tùy chọn).
	UpdateArticle(ctx context.Context, article *types.ChessArticle, revisionNote string) (*types.ChessArticle, error)
	// RenameArticleSlug đổi slug bài viết sang newSlug + ghi alias slug-cũ→mới.
	RenameArticleSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessArticle, error)
	// DeleteArticle xóa bài viết VÀ cascade: ref 2 chiều + gỡ khỏi KB tri thức cờ.
	DeleteArticle(ctx context.Context, tenantID uint64, id string) error

	// ---- Bài viết: Lịch sử phiên bản ----
	ListArticleRevisions(ctx context.Context, tenantID uint64, articleID string) ([]*types.ChessArticleRevision, error)
	GetArticleRevision(ctx context.Context, tenantID uint64, revisionID string) (*types.ChessArticleRevision, error)
	// RestoreArticleRevision ghi nội dung một bản cũ trở lại làm bản hiện tại
	// (bản thân thao tác khôi phục CŨNG tạo một bản phiên bản mới trước đó).
	RestoreArticleRevision(ctx context.Context, tenantID uint64, articleID, revisionID string) (*types.ChessArticle, error)
	// ExportArticles xuất các bài viết (theo filter) để sao lưu/chia sẻ.
	ExportArticles(ctx context.Context, tenantID uint64, f types.ChessArticleFilter) ([]types.ChessArticleBundle, error)
	// ImportArticles nhập danh sách bài viết (luôn tạo mới); trả số bài đã thêm.
	ImportArticles(ctx context.Context, tenantID uint64, items []types.ChessArticleBundle) (int, error)

	// ---- Bài viết: Chuyên mục (cây tối đa 2 tầng) ----
	ListArticleTopics(ctx context.Context, tenantID uint64, f types.ChessArticleTopicFilter) ([]*types.ChessArticleTopic, error)
	GetArticleTopic(ctx context.Context, tenantID uint64, id string) (*types.ChessArticleTopic, error)
	GetArticleTopicBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessArticleTopic, error)
	CreateArticleTopic(ctx context.Context, topic *types.ChessArticleTopic) (*types.ChessArticleTopic, error)
	UpdateArticleTopic(ctx context.Context, topic *types.ChessArticleTopic) (*types.ChessArticleTopic, error)
	// RenameArticleTopicSlug đổi slug chuyên mục sang newSlug.
	RenameArticleTopicSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessArticleTopic, error)
	// DeleteArticleTopic xóa chuyên mục — chặn nếu còn chuyên mục con.
	DeleteArticleTopic(ctx context.Context, tenantID uint64, id string) error
	// SetTopicArticles GHI ĐÈ toàn bộ danh sách bài viết trong một chuyên mục
	// theo đúng thứ tự truyền vào (xóa-rồi-chèn-lại trong transaction).
	SetTopicArticles(ctx context.Context, tenantID uint64, topicID string, articleIDs []string) error

	// ---- Bài viết: Ảnh chèn trong bài ----
	// UploadArticleImage lưu file qua FileService rồi ghi bản ghi ChessArticleImage.
	UploadArticleImage(ctx context.Context, tenantID uint64, articleID, fileName, mime string, data []byte) (*types.ChessArticleImage, error)
	// GetArticleImage trả metadata + luồng đọc nội dung ảnh (caller PHẢI Close()).
	GetArticleImage(ctx context.Context, tenantID uint64, imageID string) (*types.ChessArticleImage, io.ReadCloser, error)

	// ---- Thư viện sách: Kệ ----
	ListShelves(ctx context.Context, tenantID uint64, f types.ChessShelfFilter) ([]*types.ChessShelf, error)
	GetShelf(ctx context.Context, tenantID uint64, id string) (*types.ChessShelf, error)
	// GetShelfBySlug giải mã slug kệ (kệ KHÔNG phải đích wikilink, chỉ điều hướng UI).
	GetShelfBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessShelf, error)
	CreateShelf(ctx context.Context, shelf *types.ChessShelf) (*types.ChessShelf, error)
	UpdateShelf(ctx context.Context, shelf *types.ChessShelf) (*types.ChessShelf, error)
	// RenameShelfSlug đổi slug kệ sang newSlug + ghi alias slug-cũ→mới.
	RenameShelfSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessShelf, error)
	DeleteShelf(ctx context.Context, tenantID uint64, id string) error
	// SetShelfBooks GHI ĐÈ toàn bộ danh sách sách trên một kệ theo đúng thứ tự
	// truyền vào (xóa-rồi-chèn-lại trong transaction, mẫu ReplaceForLesson).
	SetShelfBooks(ctx context.Context, tenantID uint64, shelfID string, bookIDs []string) error
	// ListShelvesOfBook liệt kê mọi kệ đang chứa một cuốn sách (hiển thị ở form sửa sách).
	ListShelvesOfBook(ctx context.Context, tenantID uint64, bookID string) ([]*types.ChessShelf, error)

	// ---- Thư viện sách: Sách ----
	ListBooks(ctx context.Context, tenantID uint64, f types.ChessBookFilter) ([]*types.ChessBook, error)
	GetBook(ctx context.Context, tenantID uint64, id string) (*types.ChessBook, error)
	// GetBookBySlug giải mã wikilink [[book/<slug>]] về sách.
	GetBookBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessBook, error)
	// GetBookBacklinks liệt kê trang wiki/bài giảng/chương khác trỏ tới sách này.
	GetBookBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreateBook(ctx context.Context, book *types.ChessBook) (*types.ChessBook, error)
	UpdateBook(ctx context.Context, book *types.ChessBook) (*types.ChessBook, error)
	// RenameBookSlug đổi slug sách sang newSlug + ghi alias slug-cũ→mới.
	RenameBookSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessBook, error)
	// DeleteBook xóa sách VÀ cascade: chương (kèm ref 2 chiều + gỡ index từng
	// chương), pivot kệ, lịch sử phiên bản, ảnh (cả bản ghi lẫn file vật lý).
	DeleteBook(ctx context.Context, tenantID uint64, id string) error

	// ---- Thư viện sách: Chương ----
	ListChapters(ctx context.Context, tenantID uint64, bookID string) ([]*types.ChessBookChapter, error)
	// SearchChapters tìm chương theo từ khóa (slug/title/CONTENT — khác
	// SearchLessons hiện KHÔNG tìm trong content) toàn tenant — autocomplete wikilink.
	SearchChapters(ctx context.Context, tenantID uint64, keyword string, limit int) ([]*types.ChessBookChapter, error)
	GetChapter(ctx context.Context, tenantID uint64, id string) (*types.ChessBookChapter, error)
	// GetChapterBySlug giải mã wikilink [[chapter/<slug>]] về chương.
	GetChapterBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessBookChapter, error)
	// GetChapterBacklinks liệt kê trang wiki/bài giảng/chương khác trỏ tới chương này.
	GetChapterBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreateChapter(ctx context.Context, chapter *types.ChessBookChapter) (*types.ChessBookChapter, error)
	// UpdateChapter cập nhật chương; nếu Title/Content thực sự đổi thì lưu bản
	// TRƯỚC khi ghi đè vào lịch sử phiên bản (summary là ghi chú thay đổi, tùy chọn).
	UpdateChapter(ctx context.Context, chapter *types.ChessBookChapter, summary string) (*types.ChessBookChapter, error)
	// RenameChapterSlug đổi slug chương sang newSlug + ghi alias slug-cũ→mới.
	RenameChapterSlug(ctx context.Context, tenantID uint64, id, newSlug string) (*types.ChessBookChapter, error)
	DeleteChapter(ctx context.Context, tenantID uint64, id string) error
	// ReorderChapters ghi lại sort_order theo đúng thứ tự chapterIDs truyền vào.
	ReorderChapters(ctx context.Context, tenantID uint64, bookID string, chapterIDs []string) error

	// ---- Thư viện sách: Lịch sử phiên bản chương ----
	ListChapterRevisions(ctx context.Context, tenantID uint64, chapterID string) ([]*types.ChessChapterRevision, error)
	GetChapterRevision(ctx context.Context, tenantID uint64, revisionID string) (*types.ChessChapterRevision, error)
	// RestoreChapterRevision ghi nội dung một bản cũ trở lại làm bản hiện tại
	// (bản thân thao tác khôi phục CŨNG tạo một bản phiên bản mới trước đó).
	RestoreChapterRevision(ctx context.Context, tenantID uint64, chapterID, revisionID string) (*types.ChessBookChapter, error)

	// ---- Thư viện sách: Ảnh chèn trong chương ----
	// UploadBookImage lưu file qua FileService rồi ghi bản ghi ChessBookImage.
	UploadBookImage(ctx context.Context, tenantID uint64, bookID, fileName, mime string, data []byte) (*types.ChessBookImage, error)
	// GetBookImage trả metadata + luồng đọc nội dung ảnh (caller PHẢI Close()).
	GetBookImage(ctx context.Context, tenantID uint64, imageID string) (*types.ChessBookImage, io.ReadCloser, error)

	// ---- Export / Import ----
	// ExportGamesPGN xuất các ván (theo filter) thành một PGN nhiều ván.
	ExportGamesPGN(ctx context.Context, tenantID uint64, f types.ChessGameFilter) (string, error)
	// ExportPuzzles xuất các bài tập (theo filter) để sao lưu/chia sẻ.
	ExportPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]types.ChessPuzzleBundle, error)
	// ImportPuzzles nhập danh sách bài tập (luôn tạo mới); trả số bài đã thêm.
	ImportPuzzles(ctx context.Context, tenantID uint64, items []types.ChessPuzzleBundle) (int, error)
	// ExportPositions xuất các thế cờ (theo filter) để sao lưu/chia sẻ.
	ExportPositions(ctx context.Context, tenantID uint64, f types.ChessPositionFilter) ([]types.ChessPositionBundle, error)
	// ImportPositions nhập danh sách thế cờ (luôn tạo mới); trả số thế cờ đã thêm.
	ImportPositions(ctx context.Context, tenantID uint64, items []types.ChessPositionBundle) (int, error)
	// ExportBooks xuất các sách (kèm chương, theo filter) để sao lưu/chia sẻ.
	ExportBooks(ctx context.Context, tenantID uint64, f types.ChessBookFilter) ([]types.ChessBookBundle, error)
	// ImportBooks nhập danh sách sách (kèm chương), luôn tạo mới; trả số sách đã thêm.
	ImportBooks(ctx context.Context, tenantID uint64, items []types.ChessBookBundle) (int, error)

	// ReindexAll đẩy lại toàn bộ ván + bài tập của tenant vào KB tri thức cờ (chỉ
	// tác dụng khi CHESS_KB_INDEX bật). FAIL-LOUD nếu KB cờ chưa có embedding model.
	// Trả báo cáo trung thực (tổng / đã enqueue / lỗi) — "enqueued" ≠ "đã embed".
	ReindexAll(ctx context.Context, tenantID uint64) (*types.ChessReindexResult, error)
	// IndexStatus báo cáo trạng thái KB "Tri thức cờ vua" để chẩn đoán RAG cờ
	// (KB tồn tại?, có embedding model?, bao nhiêu doc completed/pending/failed).
	IndexStatus(ctx context.Context) (*types.ChessIndexStatus, error)

	// ---- Hệ thẻ thống nhất (phủ CẢ 8 loại nội dung cờ) ----
	// EnsureChessTagGroups tạo 8 thẻ nhóm nội dung dựng sẵn nếu tenant chưa có
	// (idempotent); trả số thẻ vừa tạo.
	EnsureChessTagGroups(ctx context.Context, tenantID uint64) (int, error)
	ListChessTags(ctx context.Context, tenantID uint64, f types.ChessTagFilter) ([]*types.ChessTag, error)
	GetChessTag(ctx context.Context, tenantID uint64, id string) (*types.ChessTag, error)
	GetChessTagBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessTag, error)
	// CreateChessTag tạo thẻ mới; trùng slug thì TRẢ VỀ thẻ sẵn có (không lỗi).
	CreateChessTag(ctx context.Context, tag *types.ChessTag) (*types.ChessTag, error)
	// UpdateChessTag đổi tên/mô tả/màu; nếu tên mới quy về slug của một thẻ
	// KHÁC thì tự GỘP vào thẻ đó thay vì báo trùng khóa.
	UpdateChessTag(ctx context.Context, tag *types.ChessTag) (*types.ChessTag, error)
	// DeleteChessTag xóa thẻ tự do + mọi liên kết; CHẶN xóa thẻ nhóm nội dung.
	DeleteChessTag(ctx context.Context, tenantID uint64, id string) error
	// MergeChessTags gộp thẻ nguồn vào thẻ đích rồi xóa thẻ nguồn.
	MergeChessTags(ctx context.Context, tenantID uint64, fromID, toID string) (*types.ChessTag, error)
	// RecountChessTags đếm lại cache usage_count từ bảng nối (nút chữa khi lệch).
	RecountChessTags(ctx context.Context, tenantID uint64) error
	// SetChessTags GHI ĐÈ danh sách thẻ của MỘT mục bất kỳ (tên thẻ chưa có sẽ
	// được tạo), rồi đồng bộ lại cột CSV hiển thị cho loại nào có cột đó.
	SetChessTags(ctx context.Context, tenantID uint64, chessType, chessID string, names []string) ([]*types.ChessTag, error)
	// ChessTagsFor trả thẻ của NHIỀU mục cùng loại trong một truy vấn (chống N+1).
	ChessTagsFor(ctx context.Context, tenantID uint64, chessType string, ids []string) (map[string][]*types.ChessTag, error)
	// ListChessTagItems trả một TRANG nội dung mang thẻ, gộp mọi loại (hoặc lọc
	// một loại) — "mục lục ngang" xuyên loại, có phân trang thật và tổng số.
	ListChessTagItems(ctx context.Context, tenantID uint64, tagSlug, chessType string, page, pageSize int) (*types.ChessTagItemPage, error)
	// BackfillChessTags nạp dữ liệu phân loại cũ (3 cột CSV + category/phase/
	// theme) vào hệ thẻ. Idempotent.
	BackfillChessTags(ctx context.Context, tenantID uint64) (*types.ChessTagBackfillResult, error)
}

// ChessLibraryRepository định nghĩa thao tác lưu trữ kho ván & bài tập.
type ChessLibraryRepository interface {
	// ---- Ván đấu ----
	ListGames(ctx context.Context, tenantID uint64, f types.ChessGameFilter) ([]*types.ChessGame, error)
	GetGame(ctx context.Context, tenantID uint64, id string) (*types.ChessGame, error)
	GetGameBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessGame, error)
	// GameSlugs trả mọi slug ván sống của tenant (pool fuzzy-resolve).
	GameSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	GameSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreateGame(ctx context.Context, game *types.ChessGame) error
	CreateGames(ctx context.Context, games []*types.ChessGame) error
	UpdateGame(ctx context.Context, game *types.ChessGame) error
	// UpdateGameSlug chỉ đổi cột slug (tách riêng vì UpdateGame cố tình không đụng slug).
	UpdateGameSlug(ctx context.Context, tenantID uint64, id, slug string) error
	DeleteGame(ctx context.Context, tenantID uint64, id string) error

	// ---- Bài tập ----
	ListPuzzles(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) ([]*types.ChessPuzzle, error)
	GetPuzzle(ctx context.Context, tenantID uint64, id string) (*types.ChessPuzzle, error)
	GetPuzzleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessPuzzle, error)
	// PuzzleSlugs trả mọi slug bài tập sống của tenant (pool fuzzy-resolve).
	PuzzleSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	PuzzleSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) error
	UpdatePuzzle(ctx context.Context, puzzle *types.ChessPuzzle) error
	// UpdatePuzzleSlug chỉ đổi cột slug (tách riêng như UpdateGameSlug).
	UpdatePuzzleSlug(ctx context.Context, tenantID uint64, id, slug string) error
	DeletePuzzle(ctx context.Context, tenantID uint64, id string) error
	RandomPuzzle(ctx context.Context, tenantID uint64, f types.ChessPuzzleFilter) (*types.ChessPuzzle, error)

	// ---- Thế cờ (Ngân hàng thế cờ) ----
	ListPositions(ctx context.Context, tenantID uint64, f types.ChessPositionFilter) ([]*types.ChessPosition, error)
	GetPosition(ctx context.Context, tenantID uint64, id string) (*types.ChessPosition, error)
	GetPositionBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessPosition, error)
	// PositionSlugs trả mọi slug thế cờ sống của tenant (pool fuzzy-resolve).
	PositionSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	PositionSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreatePosition(ctx context.Context, position *types.ChessPosition) error
	CreatePositions(ctx context.Context, positions []*types.ChessPosition) error
	UpdatePosition(ctx context.Context, position *types.ChessPosition) error
	// UpdatePositionSlug chỉ đổi cột slug (tách riêng như UpdateGameSlug).
	UpdatePositionSlug(ctx context.Context, tenantID uint64, id, slug string) error
	DeletePosition(ctx context.Context, tenantID uint64, id string) error
	ListPositionsByGame(ctx context.Context, tenantID uint64, gameID string) ([]*types.ChessPosition, error)
	// FindByFENKey tìm thế cờ trùng theo fen_key (phục vụ cảnh báo trùng khi tạo mới).
	FindByFENKey(ctx context.Context, tenantID uint64, fenKey string) ([]*types.ChessPosition, error)

	// ---- Bài viết (Ngân hàng bài viết) ----
	ListArticles(ctx context.Context, tenantID uint64, f types.ChessArticleFilter) ([]*types.ChessArticle, error)
	SearchArticles(ctx context.Context, tenantID uint64, keyword string, limit int) ([]*types.ChessArticle, error)
	GetArticle(ctx context.Context, tenantID uint64, id string) (*types.ChessArticle, error)
	GetArticleBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessArticle, error)
	// ArticleSlugs trả mọi slug bài viết sống của tenant (pool fuzzy-resolve).
	ArticleSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	ArticleSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreateArticle(ctx context.Context, article *types.ChessArticle) error
	UpdateArticle(ctx context.Context, article *types.ChessArticle) error
	// UpdateArticleSlug chỉ đổi cột slug (tách riêng như UpdateGameSlug).
	UpdateArticleSlug(ctx context.Context, tenantID uint64, id, slug string) error
	DeleteArticle(ctx context.Context, tenantID uint64, id string) error

	// ---- Bài viết: Lịch sử phiên bản ----
	CreateArticleRevision(ctx context.Context, rev *types.ChessArticleRevision) error
	ListArticleRevisions(ctx context.Context, tenantID uint64, articleID string) ([]*types.ChessArticleRevision, error)
	GetArticleRevision(ctx context.Context, tenantID uint64, revisionID string) (*types.ChessArticleRevision, error)
	// CountArticleRevisions phục vụ tính revision_number tiếp theo.
	CountArticleRevisions(ctx context.Context, tenantID uint64, articleID string) (int64, error)
	// DeleteArticleRevisionsByArticle xóa toàn bộ lịch sử của một bài viết (khi xóa bài viết).
	DeleteArticleRevisionsByArticle(ctx context.Context, tenantID uint64, articleID string) error

	// ---- Bài viết: Chuyên mục (cây tối đa 2 tầng) ----
	ListArticleTopics(ctx context.Context, tenantID uint64, f types.ChessArticleTopicFilter) ([]*types.ChessArticleTopic, error)
	GetArticleTopic(ctx context.Context, tenantID uint64, id string) (*types.ChessArticleTopic, error)
	GetArticleTopicBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessArticleTopic, error)
	ArticleTopicSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	ArticleTopicSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreateArticleTopic(ctx context.Context, topic *types.ChessArticleTopic) error
	// UpdateArticleTopic cố ý KHÔNG đụng slug/parent_id (tách riêng).
	UpdateArticleTopic(ctx context.Context, topic *types.ChessArticleTopic) error
	UpdateArticleTopicSlug(ctx context.Context, tenantID uint64, id, slug string) error
	UpdateArticleTopicParent(ctx context.Context, tenantID uint64, id, parentID string) error
	DeleteArticleTopic(ctx context.Context, tenantID uint64, id string) error
	CountArticlesOnTopic(ctx context.Context, tenantID uint64, topicID string) (int64, error)
	// CountArticleTopicChildren đếm chuyên mục CON — chặn xóa/lồng khi còn con.
	CountArticleTopicChildren(ctx context.Context, tenantID uint64, parentID string) (int64, error)
	// SetTopicArticles GHI ĐÈ toàn bộ bài viết trong một chuyên mục (nhiều-nhiều).
	SetTopicArticles(ctx context.Context, tenantID uint64, topicID string, articleIDs []string) error
	// RemoveArticleFromAllTopics gỡ một bài viết khỏi mọi chuyên mục (khi xóa bài viết).
	RemoveArticleFromAllTopics(ctx context.Context, tenantID uint64, articleID string) error
	// RemoveTopicItems xóa toàn bộ liên kết của một chuyên mục (khi xóa chuyên mục).
	RemoveTopicItems(ctx context.Context, tenantID uint64, topicID string) error

	// ---- Bài viết: Ảnh chèn trong bài ----
	CreateArticleImage(ctx context.Context, img *types.ChessArticleImage) error
	GetArticleImage(ctx context.Context, tenantID uint64, id string) (*types.ChessArticleImage, error)
	// ListArticleImagesByArticle phục vụ xóa file vật lý khi cascade delete bài viết.
	ListArticleImagesByArticle(ctx context.Context, tenantID uint64, articleID string) ([]*types.ChessArticleImage, error)
	DeleteArticleImage(ctx context.Context, tenantID uint64, id string) error

	// ---- Thư viện sách: Kệ ----
	ListShelves(ctx context.Context, tenantID uint64, f types.ChessShelfFilter) ([]*types.ChessShelf, error)
	GetShelf(ctx context.Context, tenantID uint64, id string) (*types.ChessShelf, error)
	GetShelfBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessShelf, error)
	// ShelfSlugs trả mọi slug kệ sống của tenant (pool fuzzy-resolve).
	ShelfSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	ShelfSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreateShelf(ctx context.Context, shelf *types.ChessShelf) error
	UpdateShelf(ctx context.Context, shelf *types.ChessShelf) error
	// UpdateShelfSlug chỉ đổi cột slug (tách riêng như UpdateGameSlug).
	UpdateShelfSlug(ctx context.Context, tenantID uint64, id, slug string) error
	DeleteShelf(ctx context.Context, tenantID uint64, id string) error
	CountBooksOnShelf(ctx context.Context, tenantID uint64, shelfID string) (int64, error)

	// ---- Thư viện sách: nối kệ↔sách (nhiều-nhiều) ----
	// SetShelfBooks GHI ĐÈ toàn bộ danh sách sách trên một kệ (xóa-rồi-chèn-lại
	// trong transaction, giữ nguyên thứ tự bookIDs qua cột sort_order).
	SetShelfBooks(ctx context.Context, tenantID uint64, shelfID string, bookIDs []string) error
	// ListShelvesOfBook liệt kê mọi kệ đang chứa một cuốn sách.
	ListShelvesOfBook(ctx context.Context, tenantID uint64, bookID string) ([]*types.ChessShelf, error)
	// RemoveBookFromAllShelves gỡ một cuốn sách khỏi mọi kệ (khi xóa sách).
	RemoveBookFromAllShelves(ctx context.Context, tenantID uint64, bookID string) error
	// RemoveShelfBooks xóa toàn bộ liên kết của một kệ (khi xóa kệ) — KHÔNG đụng sách.
	RemoveShelfBooks(ctx context.Context, tenantID uint64, shelfID string) error

	// ---- Thư viện sách: Sách ----
	ListBooks(ctx context.Context, tenantID uint64, f types.ChessBookFilter) ([]*types.ChessBook, error)
	GetBook(ctx context.Context, tenantID uint64, id string) (*types.ChessBook, error)
	GetBookBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessBook, error)
	// BookSlugs trả mọi slug sách sống của tenant (pool fuzzy-resolve).
	BookSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	BookSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreateBook(ctx context.Context, book *types.ChessBook) error
	UpdateBook(ctx context.Context, book *types.ChessBook) error
	// UpdateBookSlug chỉ đổi cột slug (tách riêng như UpdateGameSlug).
	UpdateBookSlug(ctx context.Context, tenantID uint64, id, slug string) error
	DeleteBook(ctx context.Context, tenantID uint64, id string) error
	CountChapters(ctx context.Context, tenantID uint64, bookID string) (int64, error)

	// ---- Thư viện sách: Chương ----
	ListChapters(ctx context.Context, tenantID uint64, bookID string) ([]*types.ChessBookChapter, error)
	// SearchChapters tìm chương theo từ khóa (slug/title/content) toàn tenant.
	SearchChapters(ctx context.Context, tenantID uint64, keyword string, limit int) ([]*types.ChessBookChapter, error)
	GetChapter(ctx context.Context, tenantID uint64, id string) (*types.ChessBookChapter, error)
	GetChapterBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessBookChapter, error)
	// ChapterSlugs trả mọi slug chương sống của tenant (pool fuzzy-resolve).
	ChapterSlugs(ctx context.Context, tenantID uint64) ([]string, error)
	ChapterSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreateChapter(ctx context.Context, chapter *types.ChessBookChapter) error
	UpdateChapter(ctx context.Context, chapter *types.ChessBookChapter) error
	// UpdateChapterSlug chỉ đổi cột slug (tách riêng như UpdateGameSlug).
	UpdateChapterSlug(ctx context.Context, tenantID uint64, id, slug string) error
	DeleteChapter(ctx context.Context, tenantID uint64, id string) error
	DeleteChaptersByBook(ctx context.Context, tenantID uint64, bookID string) error
	// ReorderChapters ghi lại sort_order theo đúng thứ tự chapterIDs (transaction).
	ReorderChapters(ctx context.Context, tenantID uint64, bookID string, chapterIDs []string) error

	// ---- Thư viện sách: Lịch sử phiên bản chương ----
	CreateChapterRevision(ctx context.Context, rev *types.ChessChapterRevision) error
	ListChapterRevisions(ctx context.Context, tenantID uint64, chapterID string) ([]*types.ChessChapterRevision, error)
	GetChapterRevision(ctx context.Context, tenantID uint64, revisionID string) (*types.ChessChapterRevision, error)
	// CountChapterRevisions phục vụ tính revision_number tiếp theo.
	CountChapterRevisions(ctx context.Context, tenantID uint64, chapterID string) (int64, error)
	// DeleteChapterRevisionsByChapter xóa toàn bộ lịch sử của một chương (khi xóa chương).
	DeleteChapterRevisionsByChapter(ctx context.Context, tenantID uint64, chapterID string) error

	// ---- Thư viện sách: Ảnh chèn trong chương ----
	CreateBookImage(ctx context.Context, img *types.ChessBookImage) error
	GetBookImage(ctx context.Context, tenantID uint64, id string) (*types.ChessBookImage, error)
	// ListBookImagesByBook phục vụ xóa file vật lý khi cascade delete sách.
	ListBookImagesByBook(ctx context.Context, tenantID uint64, bookID string) ([]*types.ChessBookImage, error)
	DeleteBookImage(ctx context.Context, tenantID uint64, id string) error

	// ---- Hệ thẻ thống nhất: từ điển thẻ ----
	ListTags(ctx context.Context, tenantID uint64, f types.ChessTagFilter) ([]*types.ChessTag, error)
	GetTag(ctx context.Context, tenantID uint64, id string) (*types.ChessTag, error)
	GetTagBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessTag, error)
	// GetTagsBySlugs tra nhiều thẻ trong MỘT truy vấn (chống N+1 khi lưu thẻ).
	GetTagsBySlugs(ctx context.Context, tenantID uint64, slugs []string) ([]*types.ChessTag, error)
	CreateTag(ctx context.Context, tag *types.ChessTag) error
	// UpdateTag cố ý KHÔNG đụng slug lẫn usage_count (tách riêng).
	UpdateTag(ctx context.Context, tag *types.ChessTag) error
	UpdateTagSlug(ctx context.Context, tenantID uint64, id, slug string) error
	DeleteTag(ctx context.Context, tenantID uint64, id string) error
	TagSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	// RecountTagUsage đồng bộ cache usage_count từ nguồn thật chess_tag_items;
	// tagIDs rỗng = đếm lại toàn bộ thẻ của tenant.
	RecountTagUsage(ctx context.Context, tenantID uint64, tagIDs []string) error

	// ---- Hệ thẻ thống nhất: nối thẻ với nội dung (đa hình) ----
	// SetTagsFor GHI ĐÈ toàn bộ thẻ của MỘT mục (xóa-rồi-chèn-lại, transaction).
	SetTagsFor(ctx context.Context, tenantID uint64, chessType, chessID string, tagIDs []string) error
	// RemoveAllTagsFor gỡ một mục khỏi mọi thẻ — BẮT BUỘC gọi khi xóa mục.
	RemoveAllTagsFor(ctx context.Context, tenantID uint64, chessType, chessID string) error
	// RemoveTagItems xóa mọi liên kết của một thẻ (khi xóa thẻ).
	RemoveTagItems(ctx context.Context, tenantID uint64, tagID string) error
	// MergeTagItems trỏ mọi liên kết của thẻ nguồn sang thẻ đích (gộp thẻ).
	MergeTagItems(ctx context.Context, tenantID uint64, fromTagID, toTagID string) error
	// TagsForMany trả thẻ của nhiều mục cùng loại, khóa theo chess_id.
	TagsForMany(ctx context.Context, tenantID uint64, chessType string, chessIDs []string) (map[string][]*types.ChessTag, error)
	// UpdateEntityTagsCSV ghi lại cột hiển thị `tags` (chỉ 3 loại có cột này;
	// loại khác là no-op). Đây là đường ghi DUY NHẤT vào cột đó.
	UpdateEntityTagsCSV(ctx context.Context, tenantID uint64, chessType, chessID, csv string) error

	// ---- Hệ thẻ thống nhất: tra nội dung theo thẻ ----
	CountTagItems(ctx context.Context, tenantID uint64, tagID, chessType string) (int64, error)
	// CountTagItemsByType đếm tách theo loại nội dung (hiện "sách 3 · bài viết 12").
	CountTagItemsByType(ctx context.Context, tenantID uint64, tagID string) (map[string]int64, error)
	// ListTagItems trả một TRANG liên kết của một thẻ (phân trang thật).
	ListTagItems(ctx context.Context, tenantID uint64, tagID, chessType string, offset, limit int) ([]*types.ChessTagItem, error)
}
