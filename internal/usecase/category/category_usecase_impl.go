package category

import (
	"context"
	errorsLib "errors"
	"marketplace/internal/adapter/postgres/category"
	"marketplace/internal/entity"
	"marketplace/pkg/dto"
	"marketplace/pkg/errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type categoryUsecase struct {
	adapter  category.CategoryRepository
	logger   *logrus.Logger
	validate *validator.Validate
}

func NewCategoryUsecase(
	adapter category.CategoryRepository,
	logger *logrus.Logger,
	validate *validator.Validate,
) *categoryUsecase {
	return &categoryUsecase{
		adapter:  adapter,
		logger:   logger,
		validate: validate,
	}
}

func (uc *categoryUsecase) Create(ctx context.Context, req *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	if req == nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "create",
			"req":       req,
		}).Warn("Empty request")
		return nil, errors.NewAppError("INVALID_INPUT", "empty request", nil)
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

	category := &entity.Category{
		ID:        uuid.NewString(),
		Name:      req.Name,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := uc.adapter.Create(ctx, category); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "create",
			"req":       req,
			"error":     err,
		}).Warn("Failed create category")
		return nil, errors.NewAppError("CREATE_ERR", "failed create category", err)
	}

	resp := &dto.CategoryDTO{
		CategoryID: category.ID,
		Name:       category.Name,
	}

	uc.logger.WithFields(logrus.Fields{
		"operation":     "create",
		"category_id":   req.CategoryID,
		"category_name": req.Name,
	}).Info("Category created succesfully")

	return resp, nil
}

func (uc *categoryUsecase) GetByID(ctx context.Context, id string) (*entity.Category, error) {
	if id == "" {
		uc.logger.WithFields(logrus.Fields{
			"operation": "get_by_id",
			"id":        id,
		}).Warn("Empty input")
		return nil, errors.NewAppError("INPUT_ERR", "empty id", nil)
	}

	category, err := uc.adapter.GetByID(ctx, id)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "get_by_id",
			"id":        id,
			"error":     err,
		}).Warn("Failed get by ID")
		return nil, errors.NewAppError("GET_ERR", "failed get by id", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation":     "get_by_id",
		"id":            id,
		"category_name": category.Name,
	}).Info("Successfully get category by ID")

	return category, nil
}

func (uc *categoryUsecase) Update(ctx context.Context, req *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	if req == nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "update",
			"req":       req,
		}).Warn("Empty input")
		return nil, errors.NewAppError("INPUT_ERR", "empty input", nil)
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

	category := &entity.Category{
		ID:        req.CategoryID,
		Name:      req.Name,
		UpdatedAt: time.Now().UTC(),
	}

	if err := uc.adapter.Update(ctx, category); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "update",
			"id":        category.ID,
			"name":      category.Name,
			"error":     err,
		}).Warn("Failed update category")
		return nil, errors.NewAppError("UPDATE_ERR", "failed update category", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation": "update",
		"id":        category.ID,
		"name":      category.Name,
	}).Info("Category updated successfully")

	return req, nil
}

func (uc *categoryUsecase) Delete(ctx context.Context, id string) error {
	if id == "" {
		uc.logger.WithFields(logrus.Fields{
			"operation": "delete",
			"id":        id,
		}).Warn("Empty input")
		return errors.NewAppError("INPUT_ERR", "empty id", nil)
	}

	if err := uc.adapter.Delete(ctx, id); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "delete",
			"id":        id,
			"error":     err,
		}).Warn("Failed delete category")
		return errors.NewAppError("DELETE_ERR", "failed delete category", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation": "delete",
		"id":        id,
	}).Info("Category successfully deleted")

	return nil
}

func (uc *categoryUsecase) List(ctx context.Context, limit, offset int) ([]dto.CategoryDTO, error) {
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

	categories, err := uc.adapter.List(ctx, limit, offset)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "delete",
			"limit":     limit,
			"offset":    offset,
			"error":     err,
		}).Warn("Failed list categories")
		return nil, errors.NewAppError("LIST_ERR", "failed list categories", err)
	}

	var list []dto.CategoryDTO
	for _, category := range categories {
		dtoCategory := dto.CategoryDTO{
			CategoryID: category.ID,
			Name:       category.Name,
		}
		list = append(list, dtoCategory)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation":  "list",
		"list_count": len(list),
	}).Info("Categories successfully listed")

	return list, nil
}
