package interfaces

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
)

// ChessCourseService định nghĩa nghiệp vụ quản lý khóa học & bài học cờ vua.
type ChessCourseService interface {
	// ---- Khóa học ----
	ListCourses(ctx context.Context, tenantID uint64) ([]*types.ChessCourse, error)
	GetCourse(ctx context.Context, tenantID uint64, id string) (*types.ChessCourse, error)
	// GetCourseBySlug giải mã wikilink [[course/<slug>]] về khóa học.
	GetCourseBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessCourse, error)
	// GetCourseBacklinks liệt kê trang wiki/bài giảng trỏ tới khóa học này.
	GetCourseBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreateCourse(ctx context.Context, course *types.ChessCourse) (*types.ChessCourse, error)
	UpdateCourse(ctx context.Context, course *types.ChessCourse) (*types.ChessCourse, error)
	DeleteCourse(ctx context.Context, tenantID uint64, id string) error

	// ---- Bài học ----
	ListLessons(ctx context.Context, tenantID uint64, courseID string) ([]*types.ChessLesson, error)
	GetLesson(ctx context.Context, tenantID uint64, id string) (*types.ChessLesson, error)
	// GetLessonBySlug giải mã wikilink [[lesson/<slug>]] về bài giảng.
	GetLessonBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessLesson, error)
	// GetLessonBacklinks liệt kê trang wiki/bài giảng trỏ tới bài giảng này.
	GetLessonBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error)
	CreateLesson(ctx context.Context, lesson *types.ChessLesson) (*types.ChessLesson, error)
	UpdateLesson(ctx context.Context, lesson *types.ChessLesson) (*types.ChessLesson, error)
	DeleteLesson(ctx context.Context, tenantID uint64, id string) error

	// ---- Export / Import ----
	// ExportCourses xuất toàn bộ khóa học (kèm bài học) của tenant để sao lưu/chia sẻ.
	ExportCourses(ctx context.Context, tenantID uint64) ([]types.ChessCourseBundle, error)
	// ImportCourses nhập danh sách khóa học (kèm bài học), luôn tạo mới; trả số khóa đã thêm.
	ImportCourses(ctx context.Context, tenantID uint64, bundles []types.ChessCourseBundle) (int, error)
}

// ChessCourseRepository định nghĩa thao tác lưu trữ khóa học & bài học.
type ChessCourseRepository interface {
	// ---- Khóa học ----
	ListCourses(ctx context.Context, tenantID uint64) ([]*types.ChessCourse, error)
	GetCourse(ctx context.Context, tenantID uint64, id string) (*types.ChessCourse, error)
	GetCourseBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessCourse, error)
	CourseSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreateCourse(ctx context.Context, course *types.ChessCourse) error
	UpdateCourse(ctx context.Context, course *types.ChessCourse) error
	DeleteCourse(ctx context.Context, tenantID uint64, id string) error
	CountLessons(ctx context.Context, tenantID uint64, courseID string) (int64, error)

	// ---- Bài học ----
	ListLessons(ctx context.Context, tenantID uint64, courseID string) ([]*types.ChessLesson, error)
	GetLesson(ctx context.Context, tenantID uint64, id string) (*types.ChessLesson, error)
	GetLessonBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessLesson, error)
	LessonSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error)
	CreateLesson(ctx context.Context, lesson *types.ChessLesson) error
	UpdateLesson(ctx context.Context, lesson *types.ChessLesson) error
	DeleteLesson(ctx context.Context, tenantID uint64, id string) error
	DeleteLessonsByCourse(ctx context.Context, tenantID uint64, courseID string) error
}
