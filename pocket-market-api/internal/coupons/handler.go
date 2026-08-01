package coupons

import (
	"errors"
	"net/http"
	"time"

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
	g := rg.Group("/admin/coupons", authService.RequireAuth(), auth.RequireRole(users.RoleAdmin))
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list coupons"})
		return
	}
	c.JSON(http.StatusOK, list)
}

type couponRequest struct {
	Code          string       `json:"code" binding:"required"`
	DiscountType  DiscountType `json:"discount_type" binding:"required"`
	DiscountValue float64      `json:"discount_value" binding:"required,gt=0"`
	IsActive      *bool        `json:"is_active"`
	ExpiresAt     *time.Time   `json:"expires_at"`
	UsageLimit    *int         `json:"usage_limit"`
}

func (r couponRequest) toCoupon() *Coupon {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return &Coupon{
		Code: r.Code, DiscountType: r.DiscountType, DiscountValue: r.DiscountValue,
		IsActive: isActive, ExpiresAt: r.ExpiresAt, UsageLimit: r.UsageLimit,
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req couponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.DiscountType != DiscountPercent && req.DiscountType != DiscountFixed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discount_type must be one of: percent, fixed"})
		return
	}
	if req.DiscountType == DiscountPercent && req.DiscountValue > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a percent discount cannot exceed 100"})
		return
	}

	created, err := h.repo.Create(c.Request.Context(), req.toCoupon())
	if err != nil {
		if errors.Is(err, ErrDuplicateCode) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create coupon"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) Update(c *gin.Context) {
	var req couponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.DiscountType != DiscountPercent && req.DiscountType != DiscountFixed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discount_type must be one of: percent, fixed"})
		return
	}
	if req.DiscountType == DiscountPercent && req.DiscountValue > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a percent discount cannot exceed 100"})
		return
	}

	updated, err := h.repo.Update(c.Request.Context(), c.Param("id"), req.toCoupon())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "coupon not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update coupon"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) Delete(c *gin.Context) {
	err := h.repo.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "coupon not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete coupon"})
		return
	}
	c.Status(http.StatusNoContent)
}
