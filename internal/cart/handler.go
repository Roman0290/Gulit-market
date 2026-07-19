package cart

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
	g := rg.Group("/cart", authService.RequireAuth())
	g.GET("", h.Get)
	g.POST("/items", h.AddItem)
	g.PUT("/items/:id", h.UpdateItem)
	g.DELETE("/items/:id", h.RemoveItem)
	g.DELETE("", h.Clear)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	cart, err := h.repo.GetCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch cart"})
		return
	}
	c.JSON(http.StatusOK, cart)
}

type addItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

func (h *Handler) AddItem(c *gin.Context) {
	var req addItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	item, err := h.repo.AddItem(c.Request.Context(), userID, req.ProductID, req.Quantity)
	if err != nil {
		h.handleItemError(c, err, "failed to add item to cart")
		return
	}
	c.JSON(http.StatusCreated, item)
}

type updateItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

func (h *Handler) UpdateItem(c *gin.Context) {
	var req updateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	item, err := h.repo.UpdateItemQuantity(c.Request.Context(), c.Param("id"), userID, req.Quantity)
	if err != nil {
		h.handleItemError(c, err, "failed to update cart item")
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) RemoveItem(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	err := h.repo.RemoveItem(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		h.handleItemError(c, err, "failed to remove cart item")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Clear(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	if err := h.repo.Clear(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear cart"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) handleItemError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrProductUnavailable):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}
