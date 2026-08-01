package orders

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/auth"
	"github.com/romina/pocket-market-api/internal/users"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authService *auth.Service) {
	rg.POST("/orders/checkout", authService.RequireAuth(), auth.RequireRole(users.RoleCustomer), h.Checkout)
	rg.GET("/orders", authService.RequireAuth(), auth.RequireRole(users.RoleCustomer), h.List)
	rg.GET("/orders/:id", authService.RequireAuth(), h.Get)

	vendorOnly := rg.Group("/vendor/orders", authService.RequireAuth(), auth.RequireRole(users.RoleVendor))
	vendorOnly.GET("", h.ListForVendor)
	vendorOnly.PATCH("/:id/status", h.UpdateStatus)
}

type checkoutRequest struct {
	AddressID     string        `json:"address_id" binding:"required"`
	PaymentMethod PaymentMethod `json:"payment_method" binding:"required"`
	CouponCode    string        `json:"coupon_code"`
}

func (h *Handler) Checkout(c *gin.Context) {
	var req checkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	createdOrders, err := h.service.Checkout(c.Request.Context(), userID, req.AddressID, req.PaymentMethod, req.CouponCode)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidPaymentMethod), errors.Is(err, ErrProductUnavailable), errors.Is(err, ErrInsufficientStock),
			errors.Is(err, ErrCouponInactive), errors.Is(err, ErrCouponExpired), errors.Is(err, ErrCouponExhausted):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrAddressNotOwned), errors.Is(err, ErrCouponNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrEmptyCart):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "checkout failed"})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"orders": createdOrders})
}

func (h *Handler) List(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	list, err := h.service.ListByCustomer(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)

	var o *Order
	var err error
	if role, _ := c.Get(auth.ContextRoleKey); role == users.RoleAdmin {
		o, err = h.service.GetAny(c.Request.Context(), c.Param("id"))
	} else {
		o, err = h.service.GetForParticipant(c.Request.Context(), c.Param("id"), userID)
	}
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *Handler) ListForVendor(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	list, err := h.service.ListForVendor(c.Request.Context(), userID)
	if err != nil {
		h.handleVendorError(c, err, "failed to list orders")
		return
	}
	c.JSON(http.StatusOK, list)
}

type updateStatusRequest struct {
	Status Status `json:"status" binding:"required"`
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	o, err := h.service.UpdateStatus(c.Request.Context(), userID, c.Param("id"), req.Status)
	if err != nil {
		h.handleVendorError(c, err, "failed to update order status")
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *Handler) handleVendorError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, ErrNoVendorProfile):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidStatusValue), errors.Is(err, ErrInvalidStatusTransition):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}
