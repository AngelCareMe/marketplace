package images

import (
	"context"
	"marketplace/internal/entity"
	"marketplace/pkg/dto"
)

type ImageUsecase interface {
	Create(ctx context.Context, req *dto.ImageDTO) (*dto.ImageDTO, error)
	GetByID(ctx context.Context, id string) (*entity.ProductImage, error)
	Delete(ctx context.Context, id string) error
	ListByProductID(ctx context.Context, productID string, limit, offset int) ([]dto.ImageDTO, error)
}
