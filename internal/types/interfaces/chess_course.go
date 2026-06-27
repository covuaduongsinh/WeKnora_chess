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
	CreateCourse(ctx context.Context, course *types.ChessCourse) (*types.ChessCourse, error)
	UpdateCourse(ctx context.Context, course *types.ChessCourse) (*types.ChessCourse, error)
	DeleteCourse(ctx context.Context, tenantID uint64, id string) error

	// ---- Bài học ----
	ListLessons(ctx context.Context, tenantID uint64, courseID string) ([]*types.ChessLesson, error)
	GetLesson(ctx context.Context, tenantID uint64, id string) (*types.ChessLesson, error)
	CreateLesson(ctx context.Context, lesson *types.ChessLesson) (*types.ChessLesson, error)
	UpdateLesson(ctx context.Context, lesson *types.ChessLesson) (*types.ChessLesson, error)
	DeleteLesson(ctx context.Context, tenantID uint64, id string) error
}

// ChessCourseRepository định nghĩa thao tác lưu trữ khóa học & bài học.
type ChessCourseRepository interface {
	// ---- Khóa học ----
	ListCourses(ctx context.Context, tenantID uint64) ([]*types.ChessCourse, error)
	GetCourse(ctx context.Context, tenantID uint64, id string) (*types.ChessCourse, error)
	CreateCourse(ctx context.Context, course *types.ChessCourse) error
	UpdateCourse(ctx context.Context, course *types.ChessCourse) error
	DeleteCourse(ctx context.Context, tenantID uint64, id string) error
	CountLessons(ctx context.Context, tenantID uint64, courseID string) (int64, error)

	// ---- Bài học ----
	ListLessons(ctx context.Context, tenantID uint64, courseID string) ([]*types.ChessLesson, error)
	GetLesson(ctx context.Context, tenantID uint64, id string) (*types.ChessLesson, error)
	CreateLesson(ctx context.Context, lesson *types.ChessLesson) error
	UpdateLesson(ctx context.Context, lesson *types.ChessLesson) error
	DeleteLesson(ctx context.Context, tenantID uint64, id string) error
	DeleteLessonsByCourse(ctx context.Context, tenantID uint64, courseID string) error
}
