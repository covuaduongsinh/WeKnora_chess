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
	repo         interfaces.ChessCourseRepository
	chessRefRepo interfaces.WikiChessRefRepository
}

// NewChessCourseService tạo service khóa học cờ vua.
func NewChessCourseService(
	repo interfaces.ChessCourseRepository,
	chessRefRepo interfaces.WikiChessRefRepository,
) interfaces.ChessCourseService {
	return &chessCourseService{repo: repo, chessRefRepo: chessRefRepo}
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

func (s *chessCourseService) GetCourseBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessCourse, error) {
	course, err := s.repo.GetCourseBySlug(ctx, tenantID, slug)
	if err != nil {
		return nil, err
	}
	if n, err := s.repo.CountLessons(ctx, tenantID, course.ID); err == nil {
		course.LessonCount = n
	}
	return course, nil
}

func (s *chessCourseService) GetCourseBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error) {
	if s.chessRefRepo == nil {
		return nil, nil
	}
	return s.chessRefRepo.ListBacklinks(ctx, tenantID, types.ChessRefTypeCourse, slug)
}

func (s *chessCourseService) CreateCourse(ctx context.Context, course *types.ChessCourse) (*types.ChessCourse, error) {
	if strings.TrimSpace(course.Title) == "" {
		return nil, fmt.Errorf("tên khóa học không được để trống")
	}
	course.ID = uuid.New().String()
	slug, err := ensureUniqueChessSlug(ctx, course.TenantID, courseSlugBase(course), course.ID, s.repo.CourseSlugExists)
	if err != nil {
		return nil, err
	}
	course.Slug = slug
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
	// Lấy trước slug khóa + danh sách bài (để dọn tham chiếu sau khi xóa).
	course, _ := s.repo.GetCourse(ctx, tenantID, id)
	lessons, _ := s.repo.ListLessons(ctx, tenantID, id)
	// Xóa khóa học kéo theo xóa toàn bộ bài học của nó.
	if err := s.repo.DeleteLessonsByCourse(ctx, tenantID, id); err != nil {
		return err
	}
	if err := s.repo.DeleteCourse(ctx, tenantID, id); err != nil {
		return err
	}
	if s.chessRefRepo != nil {
		if course != nil && course.Slug != "" {
			_ = s.chessRefRepo.DeleteForChess(ctx, tenantID, types.ChessRefTypeCourse, course.Slug)
		}
		for _, l := range lessons {
			if l.Slug == "" {
				continue
			}
			_ = s.chessRefRepo.DeleteForChess(ctx, tenantID, types.ChessRefTypeLesson, l.Slug) // ref TRỎ TỚI bài
			_ = s.chessRefRepo.DeleteForLesson(ctx, tenantID, l.Slug)                          // ref TỪ nội dung bài
		}
	}
	return nil
}

// ---- Bài học ----

func (s *chessCourseService) ListLessons(ctx context.Context, tenantID uint64, courseID string) ([]*types.ChessLesson, error) {
	return s.repo.ListLessons(ctx, tenantID, courseID)
}

func (s *chessCourseService) GetLesson(ctx context.Context, tenantID uint64, id string) (*types.ChessLesson, error) {
	return s.repo.GetLesson(ctx, tenantID, id)
}

func (s *chessCourseService) GetLessonBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessLesson, error) {
	return s.repo.GetLessonBySlug(ctx, tenantID, slug)
}

func (s *chessCourseService) GetLessonBacklinks(ctx context.Context, tenantID uint64, slug string) ([]types.ChessBacklink, error) {
	if s.chessRefRepo == nil {
		return nil, nil
	}
	return s.chessRefRepo.ListBacklinks(ctx, tenantID, types.ChessRefTypeLesson, slug)
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
	slug, err := ensureUniqueChessSlug(ctx, lesson.TenantID, lessonSlugBase(lesson), lesson.ID, s.repo.LessonSlugExists)
	if err != nil {
		return nil, err
	}
	lesson.Slug = slug
	if err := s.repo.CreateLesson(ctx, lesson); err != nil {
		return nil, err
	}
	s.syncLessonChessRefs(ctx, lesson)
	return lesson, nil
}

func (s *chessCourseService) UpdateLesson(ctx context.Context, lesson *types.ChessLesson) (*types.ChessLesson, error) {
	if _, err := s.repo.GetLesson(ctx, lesson.TenantID, lesson.ID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateLesson(ctx, lesson); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetLesson(ctx, lesson.TenantID, lesson.ID)
	if err != nil {
		return nil, err
	}
	s.syncLessonChessRefs(ctx, updated)
	return updated, nil
}

func (s *chessCourseService) DeleteLesson(ctx context.Context, tenantID uint64, id string) error {
	l, _ := s.repo.GetLesson(ctx, tenantID, id)
	if err := s.repo.DeleteLesson(ctx, tenantID, id); err != nil {
		return err
	}
	if l != nil && s.chessRefRepo != nil && l.Slug != "" {
		_ = s.chessRefRepo.DeleteForChess(ctx, tenantID, types.ChessRefTypeLesson, l.Slug) // ref TRỎ TỚI bài
		_ = s.chessRefRepo.DeleteForLesson(ctx, tenantID, l.Slug)                          // ref TỪ nội dung bài
	}
	return nil
}

// parseChessRefSlugs bóc tham chiếu cờ ([[game/x]], ![[puzzle/y|nhãn]], …) từ nội
// dung bài giảng — TÁI DÙNG wikiLinkRegex + splitChessRef + normalizeSlug ở
// wiki_page.go (cùng package). Trả WikiChessRef chỉ với ChessType/ChessSlug.
func parseChessRefSlugs(content string) []types.WikiChessRef {
	matches := wikiLinkRegex.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	var out []types.WikiChessRef
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		slug := strings.TrimSpace(m[1])
		if parts := strings.SplitN(slug, "|", 2); len(parts) == 2 {
			slug = strings.TrimSpace(parts[0])
		}
		slug = normalizeSlug(slug)
		t, sl, ok := splitChessRef(slug)
		if !ok {
			continue
		}
		key := t + "/" + sl
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, types.WikiChessRef{ChessType: t, ChessSlug: sl})
	}
	return out
}

// syncLessonChessRefs đồng bộ wiki_chess_refs theo nội dung bài giảng (nguồn lesson).
func (s *chessCourseService) syncLessonChessRefs(ctx context.Context, lesson *types.ChessLesson) {
	if s.chessRefRepo == nil || lesson == nil || lesson.Slug == "" {
		return
	}
	refs := parseChessRefSlugs(lesson.Content)
	for i := range refs {
		refs[i].ID = uuid.New().String()
		refs[i].TenantID = lesson.TenantID
		refs[i].SourceType = types.ChessRefSourceLesson
		refs[i].PageSlug = lesson.Slug
	}
	_ = s.chessRefRepo.ReplaceForLesson(ctx, lesson.TenantID, lesson.Slug, refs)
}
