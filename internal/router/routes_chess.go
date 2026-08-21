package router

// Route của lớp cờ vua (fork Dương Sinh).
//
// Upstream 0.7.2 tách router.go 2636 dòng thành các file routes_*.go theo miền
// (routes_agent / routes_chat / routes_knowledge / routes_infra /
// routes_auth_tenant). Nhân dịp đó tách luôn 4 nhóm route cờ ra file RIÊNG
// này thay vì nhét lại vào router.go: từ nay upstream sửa router.go gần như
// không còn đụng code cờ — đúng nguyên tắc "code cờ để riêng" của AGENTS.md.
//
// Còn phải chạm router.go đúng 2 chỗ: 4 field Chess*Handler trong RouterParams
// và 4 dòng gọi Register trong NewRouter.

import (
	"github.com/gin-gonic/gin"

	"github.com/Tencent/WeKnora/internal/handler"
)

// RegisterChessCourseRoutes đăng ký API quản lý khóa học & bài học cờ vua.
// Đây là tài nguyên cấp tenant (không gắn KB): đọc cần Viewer, ghi cần Contributor.
func RegisterChessCourseRoutes(r *gin.RouterGroup, h *handler.ChessCourseHandler, g *rbacGuards) {
	if h == nil {
		return
	}
	courses := r.Group("/chess/courses")
	{
		courses.GET("", g.Viewer(), h.ListCourses)
		courses.POST("", g.Contributor(), h.CreateCourse)
		// Export/Import (route tĩnh đặt trước param ":id").
		courses.GET("/export", g.Viewer(), h.ExportCourses)
		courses.POST("/import", g.Contributor(), h.ImportCourses)
		// Route tĩnh "by-slug" đặt trước param ":id" (giải mã wikilink [[course/<slug>]]).
		courses.GET("/by-slug/:slug", g.Viewer(), h.GetCourseBySlug)
		courses.GET("/by-slug/:slug/backlinks", g.Viewer(), h.GetCourseBacklinks)
		courses.GET("/:id", g.Viewer(), h.GetCourse)
		courses.PUT("/:id", g.Contributor(), h.UpdateCourse)
		courses.DELETE("/:id", g.Contributor(), h.DeleteCourse)
		courses.GET("/:id/lessons", g.Viewer(), h.ListLessons)
		courses.POST("/:id/lessons", g.Contributor(), h.CreateLesson)
	}
	lessons := r.Group("/chess/lessons")
	{
		// Route tĩnh "by-slug" đặt trước param ":lesson_id" (giải mã wikilink).
		lessons.GET("/by-slug/:slug", g.Viewer(), h.GetLessonBySlug)
		lessons.GET("/by-slug/:slug/backlinks", g.Viewer(), h.GetLessonBacklinks)
		lessons.GET("/:lesson_id", g.Viewer(), h.GetLesson)
		lessons.PUT("/:lesson_id", g.Contributor(), h.UpdateLesson)
		lessons.DELETE("/:lesson_id", g.Contributor(), h.DeleteLesson)
	}
}

// RegisterChessRefRoutes đăng ký API tìm kiếm hợp nhất tham chiếu cờ
// (autocomplete wikilink khi gõ "[["). Chỉ đọc → cần Viewer.
func RegisterChessRefRoutes(r *gin.RouterGroup, h *handler.ChessRefHandler, g *rbacGuards) {
	if h == nil {
		return
	}
	refs := r.Group("/chess/refs")
	{
		refs.GET("/search", g.Viewer(), h.SearchRefs)
	}
}

// RegisterChessEngineRoutes đăng ký API trạng thái engine cờ (vận hành/monitor).
// Chỉ đọc → cần Viewer.
func RegisterChessEngineRoutes(r *gin.RouterGroup, h *handler.ChessEngineHandler, g *rbacGuards) {
	if h == nil {
		return
	}
	engine := r.Group("/chess/engine")
	{
		engine.GET("/health", g.Viewer(), h.Health)
	}
}

// RegisterChessLibraryRoutes đăng ký API kho ván đấu & ngân hàng bài tập cờ vua.
// Tài nguyên cấp tenant: đọc cần Viewer, ghi cần Contributor.
func RegisterChessLibraryRoutes(r *gin.RouterGroup, h *handler.ChessLibraryHandler, g *rbacGuards) {
	if h == nil {
		return
	}
	games := r.Group("/chess/games")
	{
		games.GET("", g.Viewer(), h.ListGames)
		games.POST("", g.Contributor(), h.CreateGame)
		games.POST("/import", g.Contributor(), h.ImportGames)
		games.GET("/export", g.Viewer(), h.ExportGames)
		// Route tĩnh "by-slug" đặt trước param ":id" (giải mã wikilink [[game/<slug>]]).
		games.GET("/by-slug/:slug", g.Viewer(), h.GetGameBySlug)
		games.GET("/by-slug/:slug/backlinks", g.Viewer(), h.GetGameBacklinks)
		games.GET("/:id", g.Viewer(), h.GetGame)
		games.PUT("/:id", g.Contributor(), h.UpdateGame)
		games.PUT("/:id/slug", g.Contributor(), h.RenameGameSlug)
		games.DELETE("/:id", g.Contributor(), h.DeleteGame)
	}
	puzzles := r.Group("/chess/puzzles")
	{
		puzzles.GET("", g.Viewer(), h.ListPuzzles)
		puzzles.POST("", g.Contributor(), h.CreatePuzzle)
		puzzles.GET("/export", g.Viewer(), h.ExportPuzzles)
		puzzles.POST("/import", g.Contributor(), h.ImportPuzzles)
		puzzles.GET("/random", g.Viewer(), h.RandomPuzzle)
		// Route tĩnh "by-slug" đặt trước param ":id" (giải mã wikilink [[puzzle/<slug>]]).
		puzzles.GET("/by-slug/:slug", g.Viewer(), h.GetPuzzleBySlug)
		puzzles.GET("/by-slug/:slug/backlinks", g.Viewer(), h.GetPuzzleBacklinks)
		puzzles.GET("/:id", g.Viewer(), h.GetPuzzle)
		puzzles.PUT("/:id", g.Contributor(), h.UpdatePuzzle)
		puzzles.PUT("/:id/slug", g.Contributor(), h.RenamePuzzleSlug)
		puzzles.DELETE("/:id", g.Contributor(), h.DeletePuzzle)
	}
	// Ngân hàng thế cờ — thực thể cờ thứ 5 (game/puzzle/lesson/course/position).
	// Khác puzzle (bài TẬP có lời giải): position là thế cờ THAM CHIẾU để dạy/
	// trích dẫn/phân tích, CỐ Ý cho phép FEN không có quân Vua (dạy trẻ mới học).
	positions := r.Group("/chess/positions")
	{
		positions.GET("", g.Viewer(), h.ListPositions)
		positions.POST("", g.Contributor(), h.CreatePosition)
		positions.GET("/export", g.Viewer(), h.ExportPositions)
		positions.POST("/import", g.Contributor(), h.ImportPositions)
		// Route tĩnh "by-slug" đặt trước param ":id" (giải mã wikilink [[position/<slug>]]).
		positions.GET("/by-slug/:slug", g.Viewer(), h.GetPositionBySlug)
		positions.GET("/by-slug/:slug/backlinks", g.Viewer(), h.GetPositionBacklinks)
		positions.GET("/:id", g.Viewer(), h.GetPosition)
		positions.PUT("/:id", g.Contributor(), h.UpdatePosition)
		positions.PUT("/:id/slug", g.Contributor(), h.RenamePositionSlug)
		positions.DELETE("/:id", g.Contributor(), h.DeletePosition)
	}
	// Ngân hàng bài viết — thực thể cờ thứ 7 (bên cạnh game/puzzle/lesson/
	// course/position/book+chapter). Trang tri thức ĐỘC LẬP (khái niệm/thuật
	// ngữ/kinh nghiệm), KHÔNG thuộc sách/khóa học nào — vừa là ĐÍCH
	// [[article/<slug>]] vừa là NGUỒN wikilink. Chỉ status="published" được
	// index vào KB tri thức cờ (cùng quy tắc sách).
	articles := r.Group("/chess/articles")
	{
		articles.GET("", g.Viewer(), h.ListArticles)
		articles.POST("", g.Contributor(), h.CreateArticle)
		articles.GET("/export", g.Viewer(), h.ExportArticles)
		articles.POST("/import", g.Contributor(), h.ImportArticles)
		// Route tĩnh "images/:imageId" đặt trước param ":id" — phục vụ ảnh chèn
		// trong bài qua URL ổn định (KHÔNG phải presigned URL có hạn dùng).
		articles.GET("/images/:imageId", g.Viewer(), h.GetArticleImage)
		// Route tĩnh "by-slug" đặt trước param ":id" (giải mã wikilink [[article/<slug>]]).
		articles.GET("/by-slug/:slug", g.Viewer(), h.GetArticleBySlug)
		articles.GET("/by-slug/:slug/backlinks", g.Viewer(), h.GetArticleBacklinks)
		articles.GET("/:id", g.Viewer(), h.GetArticle)
		articles.PUT("/:id", g.Contributor(), h.UpdateArticle)
		articles.PUT("/:id/slug", g.Contributor(), h.RenameArticleSlug)
		articles.POST("/:id/images", g.Contributor(), h.UploadArticleImage)
		articles.GET("/:id/revisions", g.Viewer(), h.ListArticleRevisions)
		articles.GET("/:id/revisions/:rev_id", g.Viewer(), h.GetArticleRevision)
		articles.POST("/:id/revisions/:rev_id/restore", g.Contributor(), h.RestoreArticleRevision)
		articles.DELETE("/:id", g.Contributor(), h.DeleteArticle)
	}
	// Chuyên mục bài viết — cây tối đa 2 tầng, KHÔNG phải đích wikilink (chỉ
	// điều hướng UI, giống kệ sách).
	articleTopics := r.Group("/chess/article-topics")
	{
		articleTopics.GET("", g.Viewer(), h.ListArticleTopics)
		articleTopics.POST("", g.Contributor(), h.CreateArticleTopic)
		// Route tĩnh "by-slug" đặt trước param ":id".
		articleTopics.GET("/by-slug/:slug", g.Viewer(), h.GetArticleTopicBySlug)
		articleTopics.GET("/:id", g.Viewer(), h.GetArticleTopic)
		articleTopics.PUT("/:id", g.Contributor(), h.UpdateArticleTopic)
		articleTopics.PUT("/:id/slug", g.Contributor(), h.RenameArticleTopicSlug)
		// SetTopicArticles ghi đè toàn bộ bài viết trong chuyên mục (nhiều-nhiều).
		articleTopics.PUT("/:id/articles", g.Contributor(), h.SetTopicArticles)
		articleTopics.DELETE("/:id", g.Contributor(), h.DeleteArticleTopic)
	}
	// Thư viện sách cờ vua — Kệ → Sách → Chương. Biên soạn NỘI BỘ (markdown +
	// FEN + ván cờ + ảnh), KHÔNG phải kho ebook PDF/EPUB để đọc. Chỉ sách
	// status="published" được index vào KB tri thức cờ.
	shelves := r.Group("/chess/shelves")
	{
		shelves.GET("", g.Viewer(), h.ListShelves)
		shelves.POST("", g.Contributor(), h.CreateShelf)
		// Route tĩnh "by-slug" đặt trước param ":id".
		shelves.GET("/by-slug/:slug", g.Viewer(), h.GetShelfBySlug)
		shelves.GET("/:id", g.Viewer(), h.GetShelf)
		shelves.PUT("/:id", g.Contributor(), h.UpdateShelf)
		shelves.PUT("/:id/slug", g.Contributor(), h.RenameShelfSlug)
		// SetShelfBooks ghi đè toàn bộ sách trên kệ (nhiều-nhiều) theo thứ tự truyền vào.
		shelves.PUT("/:id/books", g.Contributor(), h.SetShelfBooks)
		shelves.DELETE("/:id", g.Contributor(), h.DeleteShelf)
	}
	books := r.Group("/chess/books")
	{
		books.GET("", g.Viewer(), h.ListBooks)
		books.POST("", g.Contributor(), h.CreateBook)
		books.GET("/export", g.Viewer(), h.ExportBooks)
		books.POST("/import", g.Contributor(), h.ImportBooks)
		// Route tĩnh "images/:imageId" đặt trước param ":id" — phục vụ ảnh chèn
		// trong chương qua URL ổn định (KHÔNG phải presigned URL có hạn dùng).
		books.GET("/images/:imageId", g.Viewer(), h.GetBookImage)
		// Route tĩnh "by-slug" đặt trước param ":id" (giải mã wikilink [[book/<slug>]]).
		books.GET("/by-slug/:slug", g.Viewer(), h.GetBookBySlug)
		books.GET("/by-slug/:slug/backlinks", g.Viewer(), h.GetBookBacklinks)
		books.GET("/:id", g.Viewer(), h.GetBook)
		books.PUT("/:id", g.Contributor(), h.UpdateBook)
		books.PUT("/:id/slug", g.Contributor(), h.RenameBookSlug)
		books.DELETE("/:id", g.Contributor(), h.DeleteBook)
		books.GET("/:id/shelves", g.Viewer(), h.ListShelvesOfBook)
		books.GET("/:id/chapters", g.Viewer(), h.ListChapters)
		books.POST("/:id/chapters", g.Contributor(), h.CreateChapter)
		books.PUT("/:id/chapters/reorder", g.Contributor(), h.ReorderChapters)
		books.POST("/:id/images", g.Contributor(), h.UploadBookImage)
	}
	chapters := r.Group("/chess/chapters")
	{
		// Route tĩnh "by-slug" đặt trước param ":chapter_id" (giải mã wikilink [[chapter/<slug>]]).
		chapters.GET("/by-slug/:slug", g.Viewer(), h.GetChapterBySlug)
		chapters.GET("/by-slug/:slug/backlinks", g.Viewer(), h.GetChapterBacklinks)
		chapters.GET("/:chapter_id", g.Viewer(), h.GetChapter)
		chapters.PUT("/:chapter_id", g.Contributor(), h.UpdateChapter)
		chapters.PUT("/:chapter_id/slug", g.Contributor(), h.RenameChapterSlug)
		chapters.DELETE("/:chapter_id", g.Contributor(), h.DeleteChapter)
		chapters.GET("/:chapter_id/revisions", g.Viewer(), h.ListChapterRevisions)
		chapters.GET("/:chapter_id/revisions/:rev_id", g.Viewer(), h.GetChapterRevision)
		chapters.POST("/:chapter_id/revisions/:rev_id/restore", g.Contributor(), h.RestoreChapterRevision)
	}
	// TÌM KIẾM HỢP NHẤT — một từ khóa, kết quả của cả 8 loại nội dung, xếp
	// hạng chung. Khác /chess/refs/search (autocomplete wikilink, không chấm
	// điểm): đây là ô tìm để TRA CỨU.
	r.GET("/chess/search", g.Viewer(), h.SearchChess)

	// HỆ THẺ THỐNG NHẤT — trục phân loại NGANG phủ cả 8 loại nội dung cờ.
	// Gồm 2 nhóm từ vựng: thẻ "group" (8 nhóm nội dung dựng sẵn, không xóa
	// được) và thẻ "free" (tự do). Đích chính là "/by-slug/:slug/items" — bấm
	// một thẻ ra MỌI loại nội dung mang thẻ đó, có phân trang thật.
	tags := r.Group("/chess/tags")
	{
		tags.GET("", g.Viewer(), h.ListChessTags)
		tags.POST("", g.Contributor(), h.CreateChessTag)
		// Route TĨNH đặt trước param ":id" (Gin không cho hai wildcard khác tên
		// cùng cấp, và "assign"/"backfill" sẽ bị nuốt thành :id nếu đặt sau).
		tags.PUT("/assign", g.Contributor(), h.AssignChessTags)
		tags.POST("/backfill", g.Contributor(), h.BackfillChessTags)
		tags.POST("/recount", g.Contributor(), h.RecountChessTags)
		tags.GET("/of/:type/:id", g.Viewer(), h.GetChessTagsOf)
		// POST (không phải GET) vì danh sách id có thể dài quá giới hạn URL.
		tags.POST("/of", g.Viewer(), h.ListChessTagsOfMany)
		tags.GET("/by-slug/:slug", g.Viewer(), h.GetChessTagBySlug)
		tags.GET("/by-slug/:slug/items", g.Viewer(), h.ListChessTagItems)
		tags.PUT("/:id", g.Contributor(), h.UpdateChessTag)
		tags.PUT("/:id/merge", g.Contributor(), h.MergeChessTags)
		tags.DELETE("/:id", g.Contributor(), h.DeleteChessTag)
	}
	// Bảo trì KB tri thức cờ (đẩy lại index sau khi bật CHESS_KB_INDEX). Nặng →
	// cần Contributor. No-op khi RAG cờ chưa bật.
	library := r.Group("/chess/library")
	{
		library.POST("/reindex", g.Contributor(), h.ReindexKB)
		// Chẩn đoán RAG cờ: KB tồn tại?, có embedding model?, completed/pending/failed.
		library.GET("/index-status", g.Contributor(), h.IndexStatus)
	}
}
