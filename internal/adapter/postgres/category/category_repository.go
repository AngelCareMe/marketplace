package category

import (
	"context"
	"marketplace/internal/entity"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id string) (*entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]entity.Category, error)
}
