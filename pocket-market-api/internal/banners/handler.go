package banners

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/auth"
	"github.com/romina/pocket-market-api/internal/users"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authService *auth.Service) {
	rg.GET("/banners", h.ListPublic)

	admin := rg.Group("/admin/banners", authService.RequireAuth(), auth.RequireRole(users.RoleAdmin))
	admin.GET("", h.ListAll)
	admin.POST("", h.Create)
	admin.PUT("/:id", h.Update)
	admin.DELETE("/:id", h.Delete)
}

func (h *Handler) ListPublic(c *gin.Context) {
	list, err := h.repo.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list banners"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) ListAll(c *gin.Context) {
	list, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list banners"})
		return
	}
	c.JSON(http.StatusOK, list)
}

type bannerRequest struct {
	ImageURL  string `json:"image_url" binding:"required"`
	LinkURL   string `json:"link_url"`
	Title     string `json:"title"`
	IsActive  *bool  `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

func (r bannerRequest) toBanner() *Banner {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return &Banner{ImageURL: r.ImageURL, LinkURL: r.LinkURL, Title: r.Title, IsActive: isActive, SortOrder: r.SortOrder}
}

func (h *Handler) Create(c *gin.Context) {
	var req bannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.repo.Create(c.Request.Context(), req.toBanner())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create banner"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) Update(c *gin.Context) {
	var req bannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.repo.Update(c.Request.Context(), c.Param("id"), req.toBanner())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "banner not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update banner"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) Delete(c *gin.Context) {
	err := h.repo.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "banner not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete banner"})
		return
	}
	c.Status(http.StatusNoContent)
}
