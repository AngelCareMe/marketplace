package customer

import (
	"context"
	"marketplace/internal/entity"
)

type CustomerRepository interface {
	UpdateProfile(ctx context.Context, profile *entity.CustomerProfile) error
	GetByUsername(ctx context.Context, username string) (*entity.CustomerProfile, error)
	GetByEmail(ctx context.Context, email string) (*entity.CustomerProfile, error)
}
