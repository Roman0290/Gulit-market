package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/users"
)

type Handler struct {
	service  *Service
	userRepo *users.Repository
}

func NewHandler(service *Service, userRepo *users.Repository) *Handler {
	return &Handler{service: service, userRepo: userRepo}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/auth/register", h.Register)
	rg.POST("/auth/login", h.Login)
	rg.GET("/auth/me", h.service.RequireAuth(), h.Me)
}

type registerRequest struct {
	Name     string     `json:"name" binding:"required"`
	Email    string     `json:"email" binding:"required,email"`
	Phone    string     `json:"phone" binding:"required"`
	Password string     `json:"password" binding:"required,min=8"`
	Role     users.Role `json:"role" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.service.Register(c.Request.Context(), RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, users.ErrDuplicateUser):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		}
		return
	}

	c.JSON(http.StatusCreated, u)
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, u, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, ErrAccountSuspended):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log in"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": u})
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)

	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, u)
}
