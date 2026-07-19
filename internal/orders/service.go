package orders

import (
	"context"
	"errors"
)

var ErrInvalidPaymentMethod = errors.New("payment_method must be one of: card, cod")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Checkout(ctx context.Context, customerID, addressID string, paymentMethod PaymentMethod) ([]Order, error) {
	if !paymentMethod.Valid() {
		return nil, ErrInvalidPaymentMethod
	}
	return s.repo.Checkout(ctx, customerID, addressID, paymentMethod)
}

func (s *Service) ListByCustomer(ctx context.Context, customerID string) ([]Order, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}

func (s *Service) GetForParticipant(ctx context.Context, orderID, userID string) (*Order, error) {
	return s.repo.GetByIDForParticipant(ctx, orderID, userID)
}
