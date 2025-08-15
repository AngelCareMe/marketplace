package usecase

import (
	"context"
	errorsLib "errors"
	"marketplace/internal/adapter/postgres/product"
	"marketplace/internal/entity"
	"marketplace/pkg/dto"
	"marketplace/pkg/errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type productUsecase struct {
	adapter  product.ProductRepository
	logger   *logrus.Logger
	validate *validator.Validate
}

func NewProductUsecase(adapter product.ProductRepository, logger *logrus.Logger, validate *validator.Validate) *productUsecase {
	return &productUsecase{
		adapter:  adapter,
		logger:   logger,
		validate: validate,
	}
}

// TODO: РЕАЛИЗОВАТЬ СОЗДАНИЕ ПРОДУКТА В КАТЕГОРИИ
func (uc *productUsecase) Create(ctx context.Context, req *dto.CreateProductRequest, categoryID string) (*dto.ProductResponse, error) {
	if req == nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "create",
			"req":       req,
		}).Warn("Request is empty")
		return nil, errors.NewAppError("INVALID_INPUT", "bad request", nil)
	}

	if err := uc.validate.StructCtx(ctx, &req); err != nil {
		var validatorErrs validator.ValidationErrors
		if errorsLib.As(err, &validatorErrs) {
			var msgs []string
			for _, e := range validatorErrs {
				msgs = append(msgs, e.Field())
			}
			uc.logger.WithFields(logrus.Fields{
				"operation": "create",
				"error":     err,
				"req":       req,
				"msgs":      msgs,
			}).Warn("Failed validation")
			return nil, errors.NewAppError("VALIDATE_ERR", "failed validate create request", err)
		}
		uc.logger.WithFields(logrus.Fields{"error": err}).Warn("Failed validation")
		return nil, errors.NewAppError("VALIDATE_ERR", "unexpected validation error", err)
	}

	existing, err := uc.adapter.GetByTitle(ctx, req.Title)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "create",
			"error":     err,
			"title":     req.Title,
		}).Warn("Failed create product")
		return nil, errors.NewAppError("CHECK_ERR", "failed check product", err)
	}
	if existing != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "create",
			"title":     req.Title,
		}).Warn("Product already exists")
		return nil, errors.NewAppError("BUSINESS_ERR", "product already exists", nil)
	}

	p := entity.Product{
		ID:         uuid.NewString(),
		SellerID:   req.SellerID,
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Price:      req.Price,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		IsActive:   true,
	}

	if err := uc.adapter.Create(ctx, &p); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "create",
			"error":     err,
			"data":      p,
		}).Warn("Failed to create product")
		return nil, errors.NewAppError("CREATE_ERR", "failed create product", err)
	}

	resp := dto.ProductResponse{
		SellerID:   p.SellerID,
		CategoryID: p.CategoryID,
		Title:      p.Title,
		Price:      p.Price,
	}

	uc.logger.WithFields(logrus.Fields{
		"operation": "create",
		"id":        p.ID,
		"title":     p.Title,
	}).Info("Product created successfully")

	return &resp, nil
}

func (uc *productUsecase) GetByTitle(ctx context.Context, title string) (*entity.Product, error) {
	if title == "" {
		uc.logger.WithFields(logrus.Fields{
			"operation": "get_by_title",
			"title":     title,
		}).Warn("Invalid input: empty title")
		return nil, errors.NewAppError("INVALID_INPUT", "empty title", nil)
	}

	product, err := uc.adapter.GetByTitle(ctx, title)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "get_by_title",
			"title":     title,
			"error":     err,
		}).Warn("Failed get by title")
		return nil, errors.NewAppError("GET_ERROR", "failed get product by title", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation": "get_by_title",
		"title":     title,
	}).Info("Successfully get product by title")

	return product, nil
}

func (uc *productUsecase) Update(ctx context.Context, req *dto.UpdateProductRequest, id string) (*dto.ProductResponse, error) {
	if req == nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "update",
			"req":       req,
		}).Warn("Request is empty")
		return nil, errors.NewAppError("INVALID_INPUT", "bad request", nil)
	}

	if err := uc.validate.StructCtx(ctx, &req); err != nil {
		var validatorErrs validator.ValidationErrors
		if errorsLib.As(err, &validatorErrs) {
			var msgs []string
			for _, e := range validatorErrs {
				msgs = append(msgs, e.Field())
			}
			uc.logger.WithFields(logrus.Fields{
				"operation": "update",
				"error":     err,
				"req":       req,
				"msgs":      msgs,
			}).Warn("Failed validation")
			return nil, errors.NewAppError("VALIDATE_ERR", "failed validate update request", err)
		}
		uc.logger.WithFields(logrus.Fields{"error": err}).Warn("Failed validation")
		return nil, errors.NewAppError("VALIDATE_ERR", "unexpected validation error", err)
	}

	if _, err := uc.adapter.GetByID(ctx, id); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "update",
			"id":        id,
			"error":     err,
		}).Warn("User not found")
		return nil, errors.NewAppError("NOT_FOUND", "user not found", err)
	}

	p := entity.Product{
		ID:         req.ID,
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Price:      req.Price,
		UpdatedAt:  time.Now().UTC(),
	}

	if err := uc.adapter.Update(ctx, &p); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "update",
			"req":       req,
			"error":     err,
		}).Warn("Failed update product")
		return nil, errors.NewAppError("UPDATE_ERR", "failed update product", err)
	}

	resp := dto.ProductResponse{
		SellerID:   p.SellerID,
		CategoryID: p.CategoryID,
		Title:      p.Title,
		Price:      p.Price,
	}

	uc.logger.WithFields(logrus.Fields{
		"operation": "update",
		"resp":      resp,
	}).Info("Product updated successfully")

	return &resp, nil
}

func (uc *productUsecase) Delete(ctx context.Context, id string) error {
	if id == "" {
		uc.logger.WithFields(logrus.Fields{
			"operation": "delete",
			"id":        id,
		}).Warn("Invalid input")
		return errors.NewAppError("INVALID_INPUT", "empty id string", nil)
	}

	if err := uc.adapter.Delete(ctx, id); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "delete",
			"id":        id,
			"error":     err,
		}).Warn("Failed delete product")
		return errors.NewAppError("DELETE_ERR", "failed delete product", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation": "delete",
		"id":        id,
	}).Info("Product deleted successfully")

	return nil
}

func (uc *productUsecase) List(ctx context.Context, categoryID string, limit, offset int) ([]dto.ProductResponse, error) {
	if categoryID == "" {
		uc.logger.WithFields(logrus.Fields{
			"operation":   "list",
			"cetegory_id": categoryID,
		}).Warn("Invalid input")
		return nil, errors.NewAppError("INVALID_INPUT", "category id is empty", nil)
	}

	if limit < 0 || limit > 100 {
		uc.logger.WithFields(logrus.Fields{
			"operation": "list",
			"limit":     limit,
		}).Warn("Invalid limit")
		limit = 40
	}

	if offset < 0 {
		uc.logger.WithFields(logrus.Fields{
			"operation": "list",
			"offset":    offset,
		}).Warn("Invalid offset")
		offset = 0
	}

	products, err := uc.adapter.List(ctx, categoryID, limit, offset)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation":   "list",
			"category_id": categoryID,
			"error":       err,
		}).Warn("Failed list products")
		return nil, errors.NewAppError("LIST_ERR", "failed list products", err)
	}

	var list []dto.ProductResponse
	for _, p := range products {
		dtoProduct := dto.ProductResponse{
			SellerID:   p.SellerID,
			CategoryID: p.CategoryID,
			Title:      p.Title,
			Price:      p.Price,
		}
		list = append(list, dtoProduct)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation":   "list",
		"category_id": categoryID,
		"list_count":  len(list),
	}).Info("Products successfully listed by category")

	return list, nil
}
