package seller

import (
	"context"
	"marketplace/internal/entity"
)

type SellerRepository interface {
	UpdateProfile(ctx context.Context, profile *entity.SellerProfile) error
	GetByUsername(ctx context.Context, username string) (*entity.SellerProfile, error)
	GetByEmail(ctx context.Context, email string) (*entity.SellerProfile, error)
}
