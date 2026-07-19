package orders

import (
	"context"
	"errors"

	"github.com/romina/pocket-market-api/internal/vendors"
)

var (
	ErrInvalidPaymentMethod    = errors.New("payment_method must be one of: card, cod")
	ErrNoVendorProfile         = errors.New("no vendor profile found for this user")
	ErrInvalidStatusValue      = errors.New("status must be one of: pending, accepted, preparing, out_for_delivery, delivered, cancelled")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// allowedTransitions defines the vendor-driven order lifecycle. Terminal
// states (delivered, cancelled) have no outgoing transitions.
var allowedTransitions = map[Status][]Status{
	StatusPending:        {StatusAccepted, StatusCancelled},
	StatusAccepted:       {StatusPreparing, StatusCancelled},
	StatusPreparing:      {StatusOutForDelivery, StatusCancelled},
	StatusOutForDelivery: {StatusDelivered},
	StatusDelivered:      {},
	StatusCancelled:      {},
}

func (s Status) valid() bool {
	_, ok := allowedTransitions[s]
	return ok
}

type Service struct {
	repo       *Repository
	vendorRepo *vendors.Repository
}

func NewService(repo *Repository, vendorRepo *vendors.Repository) *Service {
	return &Service{repo: repo, vendorRepo: vendorRepo}
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

func (s *Service) resolveVendorID(ctx context.Context, userID string) (string, error) {
	v, err := s.vendorRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, vendors.ErrNotFound) {
			return "", ErrNoVendorProfile
		}
		return "", err
	}
	return v.ID, nil
}

func (s *Service) ListForVendor(ctx context.Context, userID string) ([]Order, error) {
	vendorID, err := s.resolveVendorID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByVendor(ctx, vendorID)
}

func (s *Service) UpdateStatus(ctx context.Context, userID, orderID string, newStatus Status) (*Order, error) {
	if !newStatus.valid() {
		return nil, ErrInvalidStatusValue
	}

	vendorID, err := s.resolveVendorID(ctx, userID)
	if err != nil {
		return nil, err
	}

	currentStatus, err := s.repo.GetStatusForVendor(ctx, orderID, vendorID)
	if err != nil {
		return nil, err
	}

	if !isTransitionAllowed(currentStatus, newStatus) {
		return nil, ErrInvalidStatusTransition
	}

	return s.repo.UpdateStatusForVendor(ctx, orderID, vendorID, newStatus)
}

func isTransitionAllowed(from, to Status) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
