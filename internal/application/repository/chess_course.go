package repository

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"gorm.io/gorm"
)

// chessCourseRepository lưu trữ khóa học & bài học cờ vua trên GORM.
type chessCourseRepository struct {
	db *gorm.DB
}

// NewChessCourseRepository tạo repository khóa học cờ vua.
func NewChessCourseRepository(db *gorm.DB) interfaces.ChessCourseRepository {
	return &chessCourseRepository{db: db}
}

// ---- Khóa học ----

func (r *chessCourseRepository) ListCourses(ctx context.Context, tenantID uint64) ([]*types.ChessCourse, error) {
	var courses []*types.ChessCourse
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("sort_order ASC, created_at DESC").
		Find(&courses).Error
	return courses, err
}

func (r *chessCourseRepository) GetCourse(ctx context.Context, tenantID uint64, id string) (*types.ChessCourse, error) {
	var course types.ChessCourse
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *chessCourseRepository) GetCourseBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessCourse, error) {
	var course types.ChessCourse
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).
		First(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *chessCourseRepository) CourseSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessCourse{}).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).Limit(1).Count(&count).Error
	return count > 0, err
}

func (r *chessCourseRepository) CreateCourse(ctx context.Context, course *types.ChessCourse) error {
	return r.db.WithContext(ctx).Create(course).Error
}

func (r *chessCourseRepository) UpdateCourse(ctx context.Context, course *types.ChessCourse) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessCourse{}).
		Where("tenant_id = ? AND id = ?", course.TenantID, course.ID).
		Updates(map[string]interface{}{
			"title":       course.Title,
			"description": course.Description,
			"level":       course.Level,
			"cover_url":   course.CoverURL,
			"sort_order":  course.SortOrder,
		}).Error
}

func (r *chessCourseRepository) DeleteCourse(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&types.ChessCourse{}).Error
}

func (r *chessCourseRepository) CountLessons(ctx context.Context, tenantID uint64, courseID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&types.ChessLesson{}).
		Where("tenant_id = ? AND course_id = ?", tenantID, courseID).
		Count(&count).Error
	return count, err
}

// ---- Bài học ----

func (r *chessCourseRepository) ListLessons(ctx context.Context, tenantID uint64, courseID string) ([]*types.ChessLesson, error) {
	var lessons []*types.ChessLesson
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND course_id = ?", tenantID, courseID).
		Order("sort_order ASC, created_at ASC").
		Find(&lessons).Error
	return lessons, err
}

func (r *chessCourseRepository) GetLesson(ctx context.Context, tenantID uint64, id string) (*types.ChessLesson, error) {
	var lesson types.ChessLesson
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&lesson).Error; err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (r *chessCourseRepository) GetLessonBySlug(ctx context.Context, tenantID uint64, slug string) (*types.ChessLesson, error) {
	var lesson types.ChessLesson
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).
		First(&lesson).Error; err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (r *chessCourseRepository) LessonSlugExists(ctx context.Context, tenantID uint64, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.ChessLesson{}).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).Limit(1).Count(&count).Error
	return count > 0, err
}

func (r *chessCourseRepository) CreateLesson(ctx context.Context, lesson *types.ChessLesson) error {
	return r.db.WithContext(ctx).Create(lesson).Error
}

func (r *chessCourseRepository) UpdateLesson(ctx context.Context, lesson *types.ChessLesson) error {
	return r.db.WithContext(ctx).
		Model(&types.ChessLesson{}).
		Where("tenant_id = ? AND id = ?", lesson.TenantID, lesson.ID).
		Updates(map[string]interface{}{
			"title":      lesson.Title,
			"content":    lesson.Content,
			"fen":        lesson.FEN,
			"pgn":        lesson.PGN,
			"sort_order": lesson.SortOrder,
		}).Error
}

func (r *chessCourseRepository) DeleteLesson(ctx context.Context, tenantID uint64, id string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&types.ChessLesson{}).Error
}

func (r *chessCourseRepository) DeleteLessonsByCourse(ctx context.Context, tenantID uint64, courseID string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND course_id = ?", tenantID, courseID).
		Delete(&types.ChessLesson{}).Error
}
