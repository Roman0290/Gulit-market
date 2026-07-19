package addresses

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/auth"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authService *auth.Service) {
	g := rg.Group("/addresses", authService.RequireAuth())
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

func (h *Handler) List(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	list, err := h.repo.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list addresses"})
		return
	}
	c.JSON(http.StatusOK, list)
}

type addressRequest struct {
	Label     string   `json:"label"`
	Line1     string   `json:"line1" binding:"required"`
	City      string   `json:"city" binding:"required"`
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
	IsDefault bool     `json:"is_default"`
}

func (h *Handler) Create(c *gin.Context) {
	var req addressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	a, err := h.repo.Create(c.Request.Context(), &Address{
		UserID: userID, Label: req.Label, Line1: req.Line1, City: req.City,
		Lat: req.Lat, Lng: req.Lng, IsDefault: req.IsDefault,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create address"})
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *Handler) Update(c *gin.Context) {
	var req addressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	a, err := h.repo.Update(c.Request.Context(), c.Param("id"), userID, &Address{
		Label: req.Label, Line1: req.Line1, City: req.City, Lat: req.Lat, Lng: req.Lng, IsDefault: req.IsDefault,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update address"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	err := h.repo.Delete(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete address"})
		return
	}
	c.Status(http.StatusNoContent)
}
