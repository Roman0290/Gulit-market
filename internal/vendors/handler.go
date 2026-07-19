package vendors

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
	rg.GET("/vendors", h.List)
	rg.GET("/vendors/:id", h.Get)
	rg.POST("/vendors", authService.RequireAuth(), auth.RequireRole(users.RoleVendor), h.Register)
	rg.PUT("/vendors/:id", authService.RequireAuth(), auth.RequireRole(users.RoleVendor), h.Update)
	rg.PUT("/admin/vendors/:id/status", authService.RequireAuth(), auth.RequireRole(users.RoleAdmin), h.UpdateStatus)
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.service.ListApproved(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list vendors"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) Get(c *gin.Context) {
	v, err := h.service.GetApproved(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch vendor"})
		return
	}
	c.JSON(http.StatusOK, v)
}

type registerRequest struct {
	ShopName    string `json:"shop_name" binding:"required"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	v, err := h.service.Register(c.Request.Context(), userID, CreateInput{
		ShopName:    req.ShopName,
		Description: req.Description,
		Location:    req.Location,
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateUser) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register vendor"})
		return
	}
	c.JSON(http.StatusCreated, v)
}

type updateRequest struct {
	ShopName    string `json:"shop_name" binding:"required"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

func (h *Handler) Update(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString(auth.ContextUserIDKey)
	v, err := h.service.Update(c.Request.Context(), userID, c.Param("id"), UpdateInput{
		ShopName:    req.ShopName,
		Description: req.Description,
		Location:    req.Location,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update vendor"})
		return
	}
	c.JSON(http.StatusOK, v)
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

	v, err := h.service.UpdateStatus(c.Request.Context(), c.Param("id"), req.Status)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStatus):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update vendor status"})
		}
		return
	}
	c.JSON(http.StatusOK, v)
}
