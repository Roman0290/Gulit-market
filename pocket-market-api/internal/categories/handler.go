package categories

import (
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
	rg.GET("/categories", h.List)
	rg.POST("/admin/categories", authService.RequireAuth(), auth.RequireRole(users.RoleAdmin), h.Create)
}

func (h *Handler) List(c *gin.Context) {
	cats, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list categories"})
		return
	}
	c.JSON(http.StatusOK, cats)
}

type createCategoryRequest struct {
	Name    string `json:"name" binding:"required"`
	IconURL string `json:"icon_url"`
}

func (h *Handler) Create(c *gin.Context) {
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cat, err := h.repo.Create(c.Request.Context(), &Category{Name: req.Name, IconURL: req.IconURL})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
		return
	}
	c.JSON(http.StatusCreated, cat)
}
