package payments

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authService *auth.Service) {
	rg.POST("/payments/intent", authService.RequireAuth(), h.CreateIntent)
	rg.POST("/payments/webhook", h.Webhook)
}

type createIntentRequest struct {
	OrderID string `json:"order_id" binding:"required"`
}

func (h *Handler) CreateIntent(c *gin.Context) {
	var req createIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	result, err := h.service.CreateIntent(c.Request.Context(), userID, req.OrderID)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrOrderAlreadyPaid):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment intent"})
		}
		return
	}
	c.JSON(http.StatusCreated, result)
}

// Webhook must read the raw request body (not ShouldBindJSON) because
// Stripe's signature is computed over the exact bytes sent.
func (h *Handler) Webhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	if err := h.service.HandleWebhook(c.Request.Context(), payload, sigHeader); err != nil {
		log.Printf("webhook error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook signature verification failed"})
		return
	}

	c.Status(http.StatusOK)
}
