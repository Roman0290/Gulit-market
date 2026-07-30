package payouts

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/auth"
	"github.com/romina/pocket-market-api/internal/users"
)

type Handler struct {
	repo    *Repository
	service *Service
}

func NewHandler(repo *Repository, service *Service) *Handler {
	return &Handler{repo: repo, service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authService *auth.Service) {
	admin := rg.Group("/admin", authService.RequireAuth(), auth.RequireRole(users.RoleAdmin))
	admin.GET("/vendors/:id/balance", h.AdminBalance)
	admin.GET("/vendors/:id/payouts", h.AdminList)
	admin.POST("/payouts", h.AdminCreate)

	vendorOnly := rg.Group("/vendor", authService.RequireAuth(), auth.RequireRole(users.RoleVendor))
	vendorOnly.GET("/balance", h.MyBalance)
	vendorOnly.GET("/payouts", h.MyPayouts)
}

func (h *Handler) AdminBalance(c *gin.Context) {
	b, err := h.repo.Balance(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute balance"})
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Handler) AdminList(c *gin.Context) {
	list, err := h.repo.ListByVendor(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list payouts"})
		return
	}
	c.JSON(http.StatusOK, list)
}

type createPayoutRequest struct {
	VendorID string  `json:"vendor_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Note     string  `json:"note"`
}

func (h *Handler) AdminCreate(c *gin.Context) {
	var req createPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	balance, err := h.repo.Balance(c.Request.Context(), req.VendorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check vendor balance"})
		return
	}
	if req.Amount > balance.BalanceDue {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payout amount exceeds vendor's outstanding balance"})
		return
	}

	p, err := h.repo.Create(c.Request.Context(), &Payout{VendorID: req.VendorID, Amount: req.Amount, Note: req.Note})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payout"})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) MyBalance(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	b, err := h.service.BalanceForUser(c.Request.Context(), userID)
	if err != nil {
		h.handleVendorError(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Handler) MyPayouts(c *gin.Context) {
	userID := c.GetString(auth.ContextUserIDKey)
	list, err := h.service.ListForUser(c.Request.Context(), userID)
	if err != nil {
		h.handleVendorError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) handleVendorError(c *gin.Context, err error) {
	if errors.Is(err, ErrNoVendorProfile) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch payout information"})
}
