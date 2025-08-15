package productimage

import (
	"context"
	"marketplace/internal/entity"
)

type ProductImageRepository interface {
	Create(ctx context.Context, image *entity.ProductImage) error
	GetByID(ctx context.Context, id string) (*entity.ProductImage, error)
	Delete(ctx context.Context, id string) error
	ListByProductID(ctx context.Context, productID string, limit, offset int) ([]entity.ProductImage, error)
}
