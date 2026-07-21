package products

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

// resolveVendorID maps the authenticated user to their vendor profile ID.
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

func (s *Service) List(ctx context.Context, f ListFilter) ([]Product, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) Get(ctx context.Context, id string) (*Product, error) {
	return s.repo.GetActiveByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, userID string, p *Product) (*Product, error) {
	vendorID, err := s.resolveVendorID(ctx, userID)
	if err != nil {
		return nil, err
	}
	p.VendorID = vendorID
	return s.repo.Create(ctx, p)
}

func (s *Service) Update(ctx context.Context, userID, productID string, p *Product) (*Product, error) {
	vendorID, err := s.resolveVendorID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, productID, vendorID, p)
}

func (s *Service) UpdateStock(ctx context.Context, userID, productID string, quantity int) (*Product, error) {
	vendorID, err := s.resolveVendorID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.UpdateStock(ctx, productID, vendorID, quantity)
}

func (s *Service) Delete(ctx context.Context, userID, productID string) error {
	vendorID, err := s.resolveVendorID(ctx, userID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, productID, vendorID)
}
