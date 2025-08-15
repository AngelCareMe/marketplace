package category

import (
	"context"
	"marketplace/internal/entity"
	"marketplace/pkg/dto"
)

type CategoryUsecase interface {
	Create(ctx context.Context, req *dto.CategoryDTO) (*dto.CategoryDTO, error)
	GetByID(ctx context.Context, id string) (*entity.Category, error)
	Update(ctx context.Context, req *dto.CategoryDTO) (*dto.CategoryDTO, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset string) ([]*dto.CategoryDTO, error)
}
