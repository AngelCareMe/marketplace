package usecase

import (
	"context"
	"marketplace/internal/entity"
)

type ProductUsecase interface {
	Create(ctx context.Context, product *entity.Product, categoryID string) error
	GetByTitle(ctx context.Context, title string) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, categoryID string, limit, offset int) ([]entity.Product, error)
}
