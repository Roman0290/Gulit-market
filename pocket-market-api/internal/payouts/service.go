package payouts

import (
	"context"
	"errors"

	"github.com/romina/pocket-market-api/internal/vendors"
)

var ErrNoVendorProfile = errors.New("no vendor profile found for this user")

type Service struct {
	repo       *Repository
	vendorRepo *vendors.Repository
}

func NewService(repo *Repository, vendorRepo *vendors.Repository) *Service {
	return &Service{repo: repo, vendorRepo: vendorRepo}
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

func (s *Service) BalanceForUser(ctx context.Context, userID string) (*Balance, error) {
	vendorID, err := s.resolveVendorID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.Balance(ctx, vendorID)
}

func (s *Service) ListForUser(ctx context.Context, userID string) ([]Payout, error) {
	vendorID, err := s.resolveVendorID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByVendor(ctx, vendorID)
}
