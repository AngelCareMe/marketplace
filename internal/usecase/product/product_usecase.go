package usecase

import (
	"context"
	"marketplace/internal/entity"
	"marketplace/pkg/dto"
)

type ProductUsecase interface {
	// TODO: РЕАЛИЗОВАТЬ СОЗДАНИЕ ПРОДУКТА В КАТЕГОРИИ
	Create(ctx context.Context, product *dto.CreateProductRequest, categoryID string) (*dto.ProductResponse, error)
	GetByTitle(ctx context.Context, title string) (*entity.Product, error)
	Update(ctx context.Context, product *dto.UpdateProductRequest, id string) (*dto.ProductResponse, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, categoryID string, limit, offset int) ([]dto.ProductResponse, error)
}
