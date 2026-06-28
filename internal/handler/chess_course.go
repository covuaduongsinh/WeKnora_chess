package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

// ChessCourseHandler xử lý API quản lý khóa học & bài học cờ vua.
type ChessCourseHandler struct {
	service interfaces.ChessCourseService
}

// NewChessCourseHandler tạo handler khóa học cờ vua.
func NewChessCourseHandler(service interfaces.ChessCourseService) *ChessCourseHandler {
	return &ChessCourseHandler{service: service}
}

func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func fail(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{"success": false, "error": err.Error()})
}

// ---- Khóa học ----

// ListCourses GET /chess/courses
func (h *ChessCourseHandler) ListCourses(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	courses, err := h.service.ListCourses(ctx, tenantID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, courses)
}

// GetCourse GET /chess/courses/:id
func (h *ChessCourseHandler) GetCourse(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	course, err := h.service.GetCourse(ctx, tenantID, c.Param("id"))
	if err != nil {
		fail(c, http.StatusNotFound, err)
		return
	}
	ok(c, course)
}

// GetCourseBySlug GET /chess/courses/by-slug/:slug — giải mã wikilink [[course/<slug>]].
func (h *ChessCourseHandler) GetCourseBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	course, err := h.service.GetCourseBySlug(ctx, tenantID, c.Param("slug"))
	if err != nil {
		fail(c, http.StatusNotFound, err)
		return
	}
	ok(c, course)
}

// GetCourseBacklinks GET /chess/courses/by-slug/:slug/backlinks — trang/bài giảng trỏ tới khóa học.
func (h *ChessCourseHandler) GetCourseBacklinks(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	links, err := h.service.GetCourseBacklinks(ctx, tenantID, c.Param("slug"))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, links)
}

type courseBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Level       string `json:"level"`
	CoverURL    string `json:"cover_url"`
	SortOrder   int    `json:"sort_order"`
}

// CreateCourse POST /chess/courses
func (h *ChessCourseHandler) CreateCourse(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var body courseBody
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	course, err := h.service.CreateCourse(ctx, &types.ChessCourse{
		TenantID:    tenantID,
		Title:       body.Title,
		Description: body.Description,
		Level:       body.Level,
		CoverURL:    body.CoverURL,
		SortOrder:   body.SortOrder,
	})
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	ok(c, course)
}

// UpdateCourse PUT /chess/courses/:id
func (h *ChessCourseHandler) UpdateCourse(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var body courseBody
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	course, err := h.service.UpdateCourse(ctx, &types.ChessCourse{
		ID:          c.Param("id"),
		TenantID:    tenantID,
		Title:       body.Title,
		Description: body.Description,
		Level:       body.Level,
		CoverURL:    body.CoverURL,
		SortOrder:   body.SortOrder,
	})
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	ok(c, course)
}

// ExportCourses GET /chess/courses/export — xuất toàn bộ khóa học kèm bài học (JSON).
func (h *ChessCourseHandler) ExportCourses(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	bundles, err := h.service.ExportCourses(ctx, tenantID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, bundles)
}

// ImportCourses POST /chess/courses/import {courses:[...]} — tạo mới; trả số đã thêm.
func (h *ChessCourseHandler) ImportCourses(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var b struct {
		Courses []types.ChessCourseBundle `json:"courses"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	count, err := h.service.ImportCourses(ctx, tenantID, b.Courses)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	ok(c, gin.H{"imported": count})
}

// DeleteCourse DELETE /chess/courses/:id
func (h *ChessCourseHandler) DeleteCourse(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	if err := h.service.DeleteCourse(ctx, tenantID, c.Param("id")); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, gin.H{"deleted": true})
}

// ---- Bài học ----

// ListLessons GET /chess/courses/:id/lessons
func (h *ChessCourseHandler) ListLessons(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	lessons, err := h.service.ListLessons(ctx, tenantID, c.Param("id"))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, lessons)
}

type lessonBody struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	FEN       string `json:"fen"`
	PGN       string `json:"pgn"`
	SortOrder int    `json:"sort_order"`
}

// CreateLesson POST /chess/courses/:id/lessons
func (h *ChessCourseHandler) CreateLesson(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var body lessonBody
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	lesson, err := h.service.CreateLesson(ctx, &types.ChessLesson{
		TenantID:  tenantID,
		CourseID:  c.Param("id"),
		Title:     body.Title,
		Content:   body.Content,
		FEN:       body.FEN,
		PGN:       body.PGN,
		SortOrder: body.SortOrder,
	})
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	ok(c, lesson)
}

// GetLesson GET /chess/lessons/:lesson_id
func (h *ChessCourseHandler) GetLesson(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	lesson, err := h.service.GetLesson(ctx, tenantID, c.Param("lesson_id"))
	if err != nil {
		fail(c, http.StatusNotFound, err)
		return
	}
	ok(c, lesson)
}

// GetLessonBySlug GET /chess/lessons/by-slug/:slug — giải mã wikilink [[lesson/<slug>]].
func (h *ChessCourseHandler) GetLessonBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	lesson, err := h.service.GetLessonBySlug(ctx, tenantID, c.Param("slug"))
	if err != nil {
		fail(c, http.StatusNotFound, err)
		return
	}
	ok(c, lesson)
}

// GetLessonBacklinks GET /chess/lessons/by-slug/:slug/backlinks — trang wiki trỏ tới bài giảng.
func (h *ChessCourseHandler) GetLessonBacklinks(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	links, err := h.service.GetLessonBacklinks(ctx, tenantID, c.Param("slug"))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, links)
}

// UpdateLesson PUT /chess/lessons/:lesson_id
func (h *ChessCourseHandler) UpdateLesson(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	var body lessonBody
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	lesson, err := h.service.UpdateLesson(ctx, &types.ChessLesson{
		ID:        c.Param("lesson_id"),
		TenantID:  tenantID,
		Title:     body.Title,
		Content:   body.Content,
		FEN:       body.FEN,
		PGN:       body.PGN,
		SortOrder: body.SortOrder,
	})
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	ok(c, lesson)
}

// DeleteLesson DELETE /chess/lessons/:lesson_id
func (h *ChessCourseHandler) DeleteLesson(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := types.MustTenantIDFromContext(ctx)
	if err := h.service.DeleteLesson(ctx, tenantID, c.Param("lesson_id")); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, gin.H{"deleted": true})
}
