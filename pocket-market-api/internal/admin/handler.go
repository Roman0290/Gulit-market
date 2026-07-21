package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/auth"
	"github.com/romina/pocket-market-api/internal/orders"
	"github.com/romina/pocket-market-api/internal/users"
	"github.com/romina/pocket-market-api/internal/vendors"
)

type Handler struct {
	repo       *Repository
	orderRepo  *orders.Repository
	vendorRepo *vendors.Repository
}

func NewHandler(repo *Repository, orderRepo *orders.Repository, vendorRepo *vendors.Repository) *Handler {
	return &Handler{repo: repo, orderRepo: orderRepo, vendorRepo: vendorRepo}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authService *auth.Service) {
	g := rg.Group("/admin", authService.RequireAuth(), auth.RequireRole(users.RoleAdmin))
	g.GET("/orders", h.ListOrders)
	g.GET("/analytics", h.Analytics)
	g.GET("/vendors/pending", h.ListPendingVendors)
}

func (h *Handler) ListOrders(c *gin.Context) {
	list, err := h.orderRepo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) Analytics(c *gin.Context) {
	summary, err := h.repo.Analytics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute analytics"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) ListPendingVendors(c *gin.Context) {
	list, err := h.vendorRepo.ListPending(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list pending vendors"})
		return
	}
	c.JSON(http.StatusOK, list)
}
