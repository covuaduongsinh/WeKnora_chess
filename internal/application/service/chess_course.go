package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/Tencent/WeKnora/internal/types/interfaces"

	"github.com/Tencent/WeKnora/internal/types"
)

// chessCourseService triển khai nghiệp vụ khóa học & bài học cờ vua.
type chessCourseService struct {
	repo interfaces.ChessCourseRepository
}

// NewChessCourseService tạo service khóa học cờ vua.
func NewChessCourseService(repo interfaces.ChessCourseRepository) interfaces.ChessCourseService {
	return &chessCourseService{repo: repo}
}

// ---- Khóa học ----

func (s *chessCourseService) ListCourses(ctx context.Context, tenantID uint64) ([]*types.ChessCourse, error) {
	courses, err := s.repo.ListCourses(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// Đính kèm số bài học cho mỗi khóa.
	for _, c := range courses {
		if n, err := s.repo.CountLessons(ctx, tenantID, c.ID); err == nil {
			c.LessonCount = n
		}
	}
	return courses, nil
}

func (s *chessCourseService) GetCourse(ctx context.Context, tenantID uint64, id string) (*types.ChessCourse, error) {
	course, err := s.repo.GetCourse(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if n, err := s.repo.CountLessons(ctx, tenantID, id); err == nil {
		course.LessonCount = n
	}
	return course, nil
}

func (s *chessCourseService) CreateCourse(ctx context.Context, course *types.ChessCourse) (*types.ChessCourse, error) {
	if strings.TrimSpace(course.Title) == "" {
		return nil, fmt.Errorf("tên khóa học không được để trống")
	}
	course.ID = uuid.New().String()
	if err := s.repo.CreateCourse(ctx, course); err != nil {
		return nil, err
	}
	return course, nil
}

func (s *chessCourseService) UpdateCourse(ctx context.Context, course *types.ChessCourse) (*types.ChessCourse, error) {
	if _, err := s.repo.GetCourse(ctx, course.TenantID, course.ID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateCourse(ctx, course); err != nil {
		return nil, err
	}
	return s.GetCourse(ctx, course.TenantID, course.ID)
}

func (s *chessCourseService) DeleteCourse(ctx context.Context, tenantID uint64, id string) error {
	// Xóa khóa học kéo theo xóa toàn bộ bài học của nó.
	if err := s.repo.DeleteLessonsByCourse(ctx, tenantID, id); err != nil {
		return err
	}
	return s.repo.DeleteCourse(ctx, tenantID, id)
}

// ---- Bài học ----

func (s *chessCourseService) ListLessons(ctx context.Context, tenantID uint64, courseID string) ([]*types.ChessLesson, error) {
	return s.repo.ListLessons(ctx, tenantID, courseID)
}

func (s *chessCourseService) GetLesson(ctx context.Context, tenantID uint64, id string) (*types.ChessLesson, error) {
	return s.repo.GetLesson(ctx, tenantID, id)
}

func (s *chessCourseService) CreateLesson(ctx context.Context, lesson *types.ChessLesson) (*types.ChessLesson, error) {
	if strings.TrimSpace(lesson.Title) == "" {
		return nil, fmt.Errorf("tên bài học không được để trống")
	}
	if strings.TrimSpace(lesson.CourseID) == "" {
		return nil, fmt.Errorf("thiếu course_id")
	}
	// Đảm bảo khóa học tồn tại và thuộc tenant.
	if _, err := s.repo.GetCourse(ctx, lesson.TenantID, lesson.CourseID); err != nil {
		return nil, fmt.Errorf("khóa học không tồn tại")
	}
	lesson.ID = uuid.New().String()
	if err := s.repo.CreateLesson(ctx, lesson); err != nil {
		return nil, err
	}
	return lesson, nil
}

func (s *chessCourseService) UpdateLesson(ctx context.Context, lesson *types.ChessLesson) (*types.ChessLesson, error) {
	if _, err := s.repo.GetLesson(ctx, lesson.TenantID, lesson.ID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateLesson(ctx, lesson); err != nil {
		return nil, err
	}
	return s.repo.GetLesson(ctx, lesson.TenantID, lesson.ID)
}

func (s *chessCourseService) DeleteLesson(ctx context.Context, tenantID uint64, id string) error {
	return s.repo.DeleteLesson(ctx, tenantID, id)
}
