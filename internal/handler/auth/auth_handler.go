package auth

import (
	"errors"
	"marketplace/internal/handler/response"
	usecase "marketplace/internal/usecase/auth"
	"marketplace/pkg/dto"
	appErrors "marketplace/pkg/errors"
	"marketplace/pkg/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
	responder   *response.Responder
	validate    validator.Validator
}

func NewAuthHandler(authUsecase usecase.AuthUsecase, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		responder:   response.New(logger),
		validate:    validator.NewValidator(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responder.Error(c, err)
		return
	}
	if err := h.validate.Validate(req); err != nil {
		h.responder.Error(c, appErrors.NewAppError("VALIDATION", "invalid input", err))
		return
	}

	resp, err := h.authUsecase.Register(c.Request.Context(), req)
	if err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.Success(c, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responder.Error(c, err)
		return
	}
	if err := h.validate.Validate(req); err != nil {
		h.responder.Error(c, appErrors.NewAppError("VALIDATION", "invalid input", err))
		return
	}

	resp, err := h.authUsecase.Login(c.Request.Context(), req)
	if err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.Success(c, http.StatusOK, resp)
}

func (h *AuthHandler) UpdateAuth(c *gin.Context) {
	var req dto.UpdateAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responder.Error(c, err)
		return
	}
	if err := h.validate.Validate(req); err != nil {
		h.responder.Error(c, appErrors.NewAppError("VALIDATION", "invalid input", err))
		return
	}

	userID := c.GetString("userID")
	refreshToken := req.RefreshToken

	if refreshToken == "" {
		h.responder.Error(c, appErrors.NewAppError("VALIDATION", "missing refresh token", nil))
		return
	}

	if err := h.authUsecase.UpdateAuth(c.Request.Context(), refreshToken, userID, req); err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.NoContent(c)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")
	userType := c.GetString("userType")

	switch userType {
	case "customer":
		var req dto.CustomerProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.responder.Error(c, err)
			return
		}
		if err := h.validate.Validate(req); err != nil {
			h.responder.Error(c, appErrors.NewAppError("VALIDATION", "invalid input", err))
			return
		}
		if err := h.authUsecase.UpdateProfile(c.Request.Context(), userID, userType, req); err != nil {
			h.responder.Error(c, err)
			return
		}

	case "seller":
		var req dto.SellerProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.responder.Error(c, err)
			return
		}
		if err := h.validate.Validate(req); err != nil {
			h.responder.Error(c, appErrors.NewAppError("VALIDATION", "invalid input", err))
			return
		}
		if err := h.authUsecase.UpdateProfile(c.Request.Context(), userID, userType, req); err != nil {
			h.responder.Error(c, err)
			return
		}

	default:
		h.responder.Error(c, errors.New("unsupported user type"))
		return
	}

	h.responder.NoContent(c)
}

func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID := c.GetString("userID")

	if err := h.authUsecase.DeleteUser(c.Request.Context(), userID); err != nil {
		h.responder.Error(c, err)
		return
	}

	h.responder.NoContent(c)
}
