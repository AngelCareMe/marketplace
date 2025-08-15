package product

import (
	"context"
	"marketplace/internal/entity"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id string) (*entity.Product, error)
	GetByTitle(ctx context.Context, title string) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, categoryID string, limit, offset int) ([]entity.Product, error)
}
