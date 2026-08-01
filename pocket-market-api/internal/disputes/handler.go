package disputes

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
	rg.POST("/orders/:id/dispute", authService.RequireAuth(), h.Create)

	admin := rg.Group("/admin/disputes", authService.RequireAuth(), auth.RequireRole(users.RoleAdmin))
	admin.GET("", h.List)
	admin.PATCH("/:id/resolve", h.Resolve)
}

type createRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	d, err := h.repo.Create(c.Request.Context(), c.Param("id"), userID, req.Reason)
	if err != nil {
		if errors.Is(err, ErrNotParticipant) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open dispute"})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *Handler) List(c *gin.Context) {
	status := Status(c.Query("status"))
	list, err := h.repo.ListAll(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list disputes"})
		return
	}
	c.JSON(http.StatusOK, list)
}

type resolveRequest struct {
	ResolutionNote string `json:"resolution_note" binding:"required"`
}

func (h *Handler) Resolve(c *gin.Context) {
	var req resolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d, err := h.repo.Resolve(c.Request.Context(), c.Param("id"), req.ResolutionNote)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "dispute not found"})
		case errors.Is(err, ErrAlreadyResolved):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve dispute"})
		}
		return
	}
	c.JSON(http.StatusOK, d)
}
