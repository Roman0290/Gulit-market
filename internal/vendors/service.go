package vendors

import (
	"context"
	"errors"
)

var ErrInvalidStatus = errors.New("status must be one of: pending, approved, rejected, suspended")

func (s Status) valid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusRejected, StatusSuspended:
		return true
	}
	return false
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	ShopName    string
	Description string
	Location    string
}

func (s *Service) Register(ctx context.Context, userID string, in CreateInput) (*Vendor, error) {
	v := &Vendor{
		UserID:      userID,
		ShopName:    in.ShopName,
		Description: in.Description,
		Location:    in.Location,
		Status:      StatusPending,
	}
	return s.repo.Create(ctx, v)
}

func (s *Service) ListApproved(ctx context.Context) ([]Vendor, error) {
	return s.repo.ListApproved(ctx)
}

func (s *Service) GetApproved(ctx context.Context, id string) (*Vendor, error) {
	return s.repo.GetApprovedByID(ctx, id)
}

type UpdateInput struct {
	ShopName    string
	Description string
	Location    string
}

func (s *Service) Update(ctx context.Context, userID, vendorID string, in UpdateInput) (*Vendor, error) {
	v := &Vendor{
		ShopName:    in.ShopName,
		Description: in.Description,
		Location:    in.Location,
	}
	return s.repo.Update(ctx, vendorID, userID, v)
}

func (s *Service) UpdateStatus(ctx context.Context, vendorID string, status Status) (*Vendor, error) {
	if !status.valid() {
		return nil, ErrInvalidStatus
	}
	return s.repo.UpdateStatus(ctx, vendorID, status)
}
