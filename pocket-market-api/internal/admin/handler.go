package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/auth"
	"github.com/romina/pocket-market-api/internal/orders"
	"github.com/romina/pocket-market-api/internal/products"
	"github.com/romina/pocket-market-api/internal/users"
	"github.com/romina/pocket-market-api/internal/vendors"
)

type Handler struct {
	repo        *Repository
	orderRepo   *orders.Repository
	vendorRepo  *vendors.Repository
	userRepo    *users.Repository
	productRepo *products.Repository
}

func NewHandler(repo *Repository, orderRepo *orders.Repository, vendorRepo *vendors.Repository, userRepo *users.Repository, productRepo *products.Repository) *Handler {
	return &Handler{repo: repo, orderRepo: orderRepo, vendorRepo: vendorRepo, userRepo: userRepo, productRepo: productRepo}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authService *auth.Service) {
	g := rg.Group("/admin", authService.RequireAuth(), auth.RequireRole(users.RoleAdmin))
	g.GET("/orders", h.ListOrders)
	g.GET("/analytics", h.Analytics)
	g.GET("/vendors/pending", h.ListPendingVendors)
	g.GET("/users", h.ListUsers)
	g.PATCH("/users/:id/status", h.UpdateUserStatus)
	g.PATCH("/products/:id/status", h.UpdateProductStatus)
	g.DELETE("/products/:id", h.DeleteProduct)
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

func (h *Handler) ListUsers(c *gin.Context) {
	role := users.Role(c.Query("role"))
	list, err := h.userRepo.List(c.Request.Context(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, list)
}

type updateUserStatusRequest struct {
	Status users.Status `json:"status" binding:"required"`
}

func (h *Handler) UpdateUserStatus(c *gin.Context) {
	var req updateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status != users.StatusActive && req.Status != users.StatusSuspended {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be one of: active, suspended"})
		return
	}

	targetID := c.Param("id")
	if targetID == c.GetString(auth.ContextUserIDKey) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot change the status of your own account"})
		return
	}

	u, err := h.userRepo.UpdateStatus(c.Request.Context(), targetID, req.Status)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user status"})
		return
	}
	c.JSON(http.StatusOK, u)
}

type updateProductStatusRequest struct {
	IsActive *bool `json:"is_active" binding:"required"`
}

func (h *Handler) UpdateProductStatus(c *gin.Context) {
	var req updateProductStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.productRepo.SetActiveByAdmin(c.Request.Context(), c.Param("id"), *req.IsActive)
	if err != nil {
		if errors.Is(err, products.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product status"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	err := h.productRepo.DeleteByAdmin(c.Request.Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, products.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		case errors.Is(err, products.ErrHasOrderHistory):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
