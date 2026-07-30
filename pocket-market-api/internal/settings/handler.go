package settings

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
	g := rg.Group("/admin/settings", authService.RequireAuth(), auth.RequireRole(users.RoleAdmin))
	g.GET("", h.Get)
	g.PUT("", h.Update)
}

func (h *Handler) Get(c *gin.Context) {
	s, err := h.repo.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch settings"})
		return
	}
	c.JSON(http.StatusOK, s)
}

type updateRequest struct {
	CommissionRate     float64 `json:"commission_rate" binding:"gte=0,lte=100"`
	TaxRate            float64 `json:"tax_rate" binding:"gte=0,lte=100"`
	DefaultDeliveryFee float64 `json:"default_delivery_fee" binding:"gte=0"`
}

func (h *Handler) Update(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s, err := h.repo.Update(c.Request.Context(), &Settings{
		CommissionRate:     req.CommissionRate,
		TaxRate:            req.TaxRate,
		DefaultDeliveryFee: req.DefaultDeliveryFee,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}
	c.JSON(http.StatusOK, s)
}
