package product

import (
	"marketplace/internal/handler/response"
	usecase "marketplace/internal/usecase/product"
	appError "marketplace/pkg/errors"
	"marketplace/pkg/validator"
	"net/http"
	"strconv"

	"marketplace/pkg/dto"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type productHandler struct {
	usecase   usecase.ProductUsecase
	validate  validator.Validator
	responder *response.Responder
}

func NewProductHandler(usecase usecase.ProductUsecase, logger *logrus.Logger) *productHandler {
	return &productHandler{
		usecase:   usecase,
		responder: response.New(logger),
		validate:  validator.NewValidator(),
	}
}

func (h *productHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	categoryID := c.Param("categoryID")

	if err := c.ShouldBindJSON(&req); err != nil {
		h.responder.Error(c, err)
		return
	}

	if err := h.validate.Validate(req); err != nil {
		h.responder.Error(c, appError.NewAppError("VALIDATION", "invalid input", err))
		return
	}

	resp, err := h.usecase.Create(c, &req, categoryID)
	if err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.Success(c, http.StatusCreated, resp)
}

func (h *productHandler) GetByTitle(c *gin.Context) {
	title := c.Param("title")

	product, err := h.usecase.GetByTitle(c, title)
	if err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.Success(c, http.StatusOK, product)
}

func (h *productHandler) Update(c *gin.Context) {
	var req dto.UpdateProductRequest
	productId := c.Param("productID")

	if err := c.ShouldBindJSON(&req); err != nil {
		h.responder.Error(c, err)
		return
	}

	if err := h.validate.Validate(req); err != nil {
		h.responder.Error(c, err)
		return
	}

	resp, err := h.usecase.Update(c, &req, productId)
	if err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.Success(c, http.StatusOK, resp)
}

func (h *productHandler) Delete(c *gin.Context) {
	productID := c.Param("productID")

	if err := h.usecase.Delete(c, productID); err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.NoContent(c)
}

func (h *productHandler) List(c *gin.Context) {
	categoryID := c.Param("categoryID")
	limitStr := c.Query("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	offsetStr := c.Query("offset")
	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	products, err := h.usecase.List(c, categoryID, limit, offset)
	if err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.Success(c, http.StatusOK, products)
}
