package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/romina/pocket-market-api/internal/users"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRole        = errors.New("role must be customer or vendor")
	ErrAccountSuspended   = errors.New("this account has been suspended")
)

const tokenTTL = 24 * time.Hour

type Claims struct {
	UserID string     `json:"user_id"`
	Role   users.Role `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	userRepo  *users.Repository
	jwtSecret []byte
}

func NewService(userRepo *users.Repository, jwtSecret string) *Service {
	return &Service{userRepo: userRepo, jwtSecret: []byte(jwtSecret)}
}

type RegisterInput struct {
	Name     string
	Email    string
	Phone    string
	Password string
	Role     users.Role
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*users.User, error) {
	if in.Role != users.RoleCustomer && in.Role != users.RoleVendor {
		return nil, ErrInvalidRole
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &users.User{
		Name:         in.Name,
		Email:        in.Email,
		Phone:        in.Phone,
		PasswordHash: string(hash),
		Role:         in.Role,
	}

	return s.userRepo.Create(ctx, u)
}

func (s *Service) Login(ctx context.Context, email, password string) (string, *users.User, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if u.Status == users.StatusSuspended {
		return "", nil, ErrAccountSuspended
	}

	token, err := s.generateToken(u)
	if err != nil {
		return "", nil, err
	}

	return token, u, nil
}

func (s *Service) generateToken(u *users.User) (string, error) {
	claims := Claims{
		UserID: u.ID,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
