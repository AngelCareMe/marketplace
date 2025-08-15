package images

import (
	"context"
	errorsLib "errors"
	productimage "marketplace/internal/adapter/postgres/product_image"
	"marketplace/internal/entity"
	"marketplace/pkg/dto"
	"marketplace/pkg/errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type imageUsecase struct {
	adapter  productimage.ProductImageRepository
	logger   *logrus.Logger
	validate *validator.Validate
}

func NewImageUsecase(
	adapter productimage.ProductImageRepository,
	logger *logrus.Logger,
	validate *validator.Validate,
) *imageUsecase {
	return &imageUsecase{
		adapter:  adapter,
		logger:   logger,
		validate: validate,
	}
}

func (uc *imageUsecase) Create(ctx context.Context, req *dto.ImageDTO) (*dto.ImageDTO, error) {
	if req == nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "create",
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

	image := &entity.ProductImage{
		ID:        uuid.NewString(),
		ProductID: req.ProductID,
		URL:       req.URL,
		CreatedAt: time.Now().UTC(),
	}

	if err := uc.adapter.Create(ctx, image); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "create",
			"req":       req,
			"error":     err,
		}).Warn("Failed create image")
		return nil, errors.NewAppError("CREATE_ERR", "failed create image", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation":  "create",
		"product_id": req.ProductID,
		"url":        req.URL,
	}).Info("Image successfully created")

	return req, nil
}

func (uc *imageUsecase) GetByID(ctx context.Context, id string) (*entity.ProductImage, error) {
	if id == "" {
		uc.logger.WithFields(logrus.Fields{
			"operation": "get_by_id",
			"id":        id,
		}).Warn("Empty input")
		return nil, errors.NewAppError("INPUT_ERR", "empty id", nil)
	}

	image, err := uc.adapter.GetByID(ctx, id)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation": "get_by_id",
			"id":        id,
			"error":     err,
		}).Warn("Failed get by ID")
		return nil, errors.NewAppError("GET_ERR", "failed get by id", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation": "get_by_id",
		"id":        id,
		"url":       image.URL,
	}).Info("Successfully get image by ID")

	return image, nil
}

func (uc *imageUsecase) Delete(ctx context.Context, id string) error {
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
		}).Warn("Failed delete image")
		return errors.NewAppError("DELETE_ERR", "failed delete image", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation": "delete",
		"id":        id,
	}).Info("Image successfully deleted")

	return nil
}

func (uc *imageUsecase) ListByProductID(ctx context.Context, productID string, limit, offset int) ([]dto.ImageDTO, error) {
	if productID == "" {
		uc.logger.WithFields(logrus.Fields{
			"operation": "delete",
			"id":        productID,
		}).Warn("Empty input")
		return nil, errors.NewAppError("INPUT_ERR", "empty id", nil)
	}

	if limit < 0 || limit > 20 {
		uc.logger.WithFields(logrus.Fields{
			"operation": "list",
			"limit":     limit,
		}).Warn("Invalid limit")
		limit = 20
	}

	if offset < 0 {
		uc.logger.WithFields(logrus.Fields{
			"operation": "list",
			"offset":    offset,
		}).Warn("Invalid offset")
		offset = 0
	}

	images, err := uc.adapter.ListByProductID(ctx, productID, limit, offset)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"operation":  "list",
			"product_id": productID,
			"error":      err,
		}).Warn("Failed list images")
		return nil, errors.NewAppError("LIST_ERR", "failed list images", err)
	}

	var list []dto.ImageDTO
	for _, image := range images {
		dtoImage := dto.ImageDTO{
			ProductID: image.ID,
			URL:       image.URL,
		}
		list = append(list, dtoImage)
	}

	uc.logger.WithFields(logrus.Fields{
		"operation":  "list",
		"product_id": productID,
		"list_count": len(list),
	}).Info("Images successfully listed by category")

	return list, nil
}
