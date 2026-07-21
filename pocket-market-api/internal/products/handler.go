package products

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
	rg.GET("/products", h.List)
	rg.GET("/products/:id", h.Get)

	vendorOnly := rg.Group("/vendor/products", authService.RequireAuth(), auth.RequireRole(users.RoleVendor))
	vendorOnly.POST("", h.Create)
	vendorOnly.PUT("/:id", h.Update)
	vendorOnly.DELETE("/:id", h.Delete)
	vendorOnly.PATCH("/:id/stock", h.UpdateStock)
}

func (h *Handler) List(c *gin.Context) {
	f := ListFilter{
		CategoryID: c.Query("category"),
		VendorID:   c.Query("vendor_id"),
		Query:      c.Query("q"),
	}

	list, err := h.service.List(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) Get(c *gin.Context) {
	p, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch product"})
		return
	}
	c.JSON(http.StatusOK, p)
}

type productRequest struct {
	CategoryID    string  `json:"category_id"`
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	Unit          string  `json:"unit" binding:"required"`
	StockQuantity int     `json:"stock_quantity"`
	ImageURL      string  `json:"image_url"`
	IsActive      *bool   `json:"is_active"`
}

func (r productRequest) toProduct() *Product {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return &Product{
		CategoryID:    r.CategoryID,
		Name:          r.Name,
		Description:   r.Description,
		Price:         r.Price,
		Unit:          r.Unit,
		StockQuantity: r.StockQuantity,
		ImageURL:      r.ImageURL,
		IsActive:      isActive,
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	p, err := h.service.Create(c.Request.Context(), userID, req.toProduct())
	if err != nil {
		h.handleVendorError(c, err, "failed to create product")
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) Update(c *gin.Context) {
	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	p, err := h.service.Update(c.Request.Context(), userID, c.Param("id"), req.toProduct())
	if err != nil {
		h.handleVendorError(c, err, "failed to update product")
		return
	}
	c.JSON(http.StatusOK, p)
}

type stockRequest struct {
	StockQuantity int `json:"stock_quantity" binding:"gte=0"`
}

func (h *Handler) UpdateStock(c *gin.Context) {
	var req stockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	p, err := h.service.UpdateStock(c.Request.Context(), userID, c.Param("id"), req.StockQuantity)
	if err != nil {
		h.handleVendorError(c, err, "failed to update stock")
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	err := h.service.Delete(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		h.handleVendorError(c, err, "failed to delete product")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) handleVendorError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, ErrNoVendorProfile):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}
