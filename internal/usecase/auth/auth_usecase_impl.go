package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"marketplace/internal/adapter/bcrypt"
	"marketplace/internal/adapter/jwt"
	"marketplace/internal/adapter/postgres/customer"
	"marketplace/internal/adapter/postgres/seller"
	"marketplace/internal/adapter/postgres/token"
	"marketplace/internal/adapter/postgres/user"
	"marketplace/internal/entity"
	"marketplace/pkg/dto"
	appErrors "marketplace/pkg/errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type authUsecase struct {
	userRepo     user.UserRepository
	customerRepo customer.CustomerRepository
	sellerRepo   seller.SellerRepository
	tokenRepo    token.TokenRepository
	jwtManager   jwt.JWTManager
	hashManager  bcrypt.Hasher
	validator    *validator.Validate
	logger       *logrus.Logger
}

func NewAuthUsecase(
	userRepo user.UserRepository,
	customerRepo customer.CustomerRepository,
	sellerRepo seller.SellerRepository,
	tokenRepo token.TokenRepository,
	jwtManager jwt.JWTManager,
	hashManager bcrypt.Hasher,
	logger *logrus.Logger,
) *authUsecase {
	return &authUsecase{
		userRepo:     userRepo,
		customerRepo: customerRepo,
		sellerRepo:   sellerRepo,
		tokenRepo:    tokenRepo,
		jwtManager:   jwtManager,
		hashManager:  hashManager,
		validator:    validator.New(),
		logger:       logger,
	}
}

func (uc *authUsecase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	if err := uc.validator.Struct(req); err != nil {
		return nil, appErrors.NewAppError("VALIDATION", "invalid registration data", err)
	}

	userType := strings.ToLower(strings.TrimSpace(req.UserType))
	if userType != "customer" && userType != "seller" {
		uc.logger.WithField("user_type", req.UserType).Warn("invalid user_type")
		return nil, appErrors.NewAppError("INVALID_TYPE", "unsupported user_type", nil)
	}

	// Проверка уникальности
	var existingErr error
	switch userType {
	case "customer":
		_, existingErr = uc.customerRepo.GetByEmail(ctx, req.Email)
		if existingErr == nil {
			return nil, appErrors.NewAppError("DUPLICATE", "email already exists", nil)
		}
		_, existingErr = uc.customerRepo.GetByUsername(ctx, req.Username)
		if existingErr == nil {
			return nil, appErrors.NewAppError("DUPLICATE", "username already exists", nil)
		}
	case "seller":
		_, existingErr = uc.sellerRepo.GetByEmail(ctx, req.Email)
		if existingErr == nil {
			return nil, appErrors.NewAppError("DUPLICATE", "email already exists", nil)
		}
		_, existingErr = uc.sellerRepo.GetByUsername(ctx, req.Username)
		if existingErr == nil {
			return nil, appErrors.NewAppError("DUPLICATE", "username already exists", nil)
		}
	}
	if existingErr != nil && !errors.Is(existingErr, appErrors.ErrNotFound) {
		uc.logger.WithError(existingErr).Error("failed to check uniqueness")
		return nil, appErrors.NewAppError("REPO", "uniqueness check failed", existingErr)
	}

	hashed, err := uc.hashManager.GenerateHashPassword(req.Password)
	if err != nil {
		uc.logger.WithField("email", req.Email).Error("failed to hash password")
		return nil, appErrors.NewAppError("HASHING", "failed to hash password", err)
	}

	now := time.Now()
	u := &entity.User{
		ID:           uuid.NewString(),
		UserType:     userType,
		Username:     req.Username,
		PasswordHash: hashed,
		Email:        req.Email,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.userRepo.Create(ctx, u); err != nil {
		uc.logger.WithFields(logrus.Fields{"user_id": u.ID, "type": u.UserType}).Error("user create failed")
		return nil, appErrors.NewAppError("USER_CREATE_FAIL", "failed to create user", err)
	}

	access, err := uc.jwtManager.GenerateAccessToken(u)
	if err != nil {
		return nil, appErrors.NewAppError("JWT_GENERATION", "failed to generate access token", err)
	}

	refresh, err := uc.jwtManager.GenerateRefreshToken(ctx, u)
	if err != nil {
		return nil, appErrors.NewAppError("JWT_GENERATION", "failed to generate refresh token", err)
	}

	uc.logger.WithFields(logrus.Fields{"user_id": u.ID, "type": u.UserType}).Info("user registered")

	return &dto.AuthResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func (uc *authUsecase) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	if err := uc.validator.Struct(req); err != nil {
		return nil, appErrors.NewAppError("VALIDATION", "invalid login data", err)
	}

	userType := strings.ToLower(strings.TrimSpace(req.UserType))
	if userType != "customer" && userType != "seller" {
		uc.logger.WithField("user_type", req.UserType).Warn("invalid user_type")
		return nil, appErrors.NewAppError("INVALID_TYPE", "unsupported user_type", nil)
	}

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	if username != "" && email != "" {
		return nil, appErrors.NewAppError("VALIDATION", "provide either username or email, not both", nil)
	}
	if username == "" && email == "" {
		return nil, appErrors.NewAppError("VALIDATION", "email or username is required", nil)
	}

	lookupBy := "email"
	identifier := email
	if username != "" {
		lookupBy = "username"
		identifier = username
	}

	var u entity.User
	var err error

	switch userType {
	case "customer":
		var c *entity.CustomerProfile
		if lookupBy == "email" {
			c, err = uc.customerRepo.GetByEmail(ctx, identifier)
		} else {
			c, err = uc.customerRepo.GetByUsername(ctx, identifier)
		}
		if err != nil {
			if errors.Is(err, appErrors.ErrNotFound) {
				return nil, appErrors.NewAppError("INVALID_CREDENTIALS", "invalid credentials", nil)
			}
			return nil, appErrors.NewAppError("REPO", "failed to fetch user", err)
		}
		if err = uc.hashManager.CompareHashPassword(c.PasswordHash, req.Password); err != nil {
			return nil, appErrors.NewAppError("INVALID_CREDENTIALS", "invalid credentials", nil)
		}
		u = entity.User{ID: c.ID, UserType: userType, Username: c.Username, Email: c.Email, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}

	case "seller":
		var s *entity.SellerProfile
		if lookupBy == "email" {
			s, err = uc.sellerRepo.GetByEmail(ctx, identifier)
		} else {
			s, err = uc.sellerRepo.GetByUsername(ctx, identifier)
		}
		if err != nil {
			if errors.Is(err, appErrors.ErrNotFound) {
				return nil, appErrors.NewAppError("INVALID_CREDENTIALS", "invalid credentials", nil)
			}
			return nil, appErrors.NewAppError("REPO", "failed to fetch user", err)
		}
		if err = uc.hashManager.CompareHashPassword(s.PasswordHash, req.Password); err != nil {
			return nil, appErrors.NewAppError("INVALID_CREDENTIALS", "invalid credentials", nil)
		}
		u = entity.User{ID: s.ID, UserType: userType, Username: s.Username, Email: s.Email, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
	}

	access, err := uc.jwtManager.GenerateAccessToken(&u)
	if err != nil {
		return nil, appErrors.NewAppError("JWT_GENERATION", "failed to generate access token", err)
	}

	refresh, err := uc.jwtManager.GenerateRefreshToken(ctx, &u)
	if err != nil {
		return nil, appErrors.NewAppError("JWT_GENERATION", "failed to generate refresh token", err)
	}

	uc.logger.WithFields(logrus.Fields{"user_id": u.ID, "type": u.UserType}).Info("user logged in")

	return &dto.AuthResponse{AccessToken: access, RefreshToken: refresh}, nil
}

func (uc *authUsecase) UpdateAuth(ctx context.Context, tokenString, userID string, req dto.UpdateAuthRequest) error {
	if err := uc.validator.Struct(req); err != nil {
		return appErrors.NewAppError("VALIDATION", "invalid update data", err)
	}

	if err := uc.jwtManager.ValidateRefreshToken(ctx, tokenString); err != nil {
		return appErrors.NewAppError("INVALID_TOKEN", "invalid refresh token", err)
	}

	userByID, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return appErrors.NewAppError("NOT_FOUND", "user not found", err)
	}

	var newHash string = userByID.PasswordHash
	if req.NewPassword != "" {
		if req.OldPassword == "" {
			return appErrors.NewAppError("VALIDATION", "old password required", nil)
		}
		if err := uc.hashManager.CompareHashPassword(userByID.PasswordHash, req.OldPassword); err != nil {
			return appErrors.NewAppError("INVALID_CREDENTIALS", "old password incorrect", nil)
		}
		newHash, err = uc.hashManager.GenerateHashPassword(req.NewPassword)
		if err != nil {
			return appErrors.NewAppError("HASHING", "failed to hash new password", err)
		}
	}

	email := userByID.Email
	if req.Email != "" {
		email = req.Email
	}

	username := userByID.Username
	if req.Username != "" {
		username = req.Username
	}

	if err := uc.userRepo.UpdateAuth(ctx, userID, username, email, newHash); err != nil {
		return appErrors.NewAppError("UPDATE_FAILED", "failed to update user", err)
	}

	if err := uc.revokeRefreshToken(ctx, userID); err != nil {
		uc.logger.WithField("user_id", userID).Warn("failed to revoke token after update")
	}

	return nil
}

func (uc *authUsecase) UpdateProfile(ctx context.Context, userID string, userType string, payload interface{}) error {
	userType = strings.ToLower(strings.TrimSpace(userType))
	now := time.Now()

	switch userType {
	case "customer":
		req, ok := payload.(dto.CustomerProfileRequest)
		if !ok {
			return appErrors.NewAppError("INVALID_PAYLOAD", "invalid customer payload", errors.New("type mismatch"))
		}
		if err := uc.validator.Struct(req); err != nil {
			return appErrors.NewAppError("VALIDATION", "invalid customer profile data", err)
		}

		profile := &entity.CustomerProfile{
			User:      entity.User{ID: userID, UpdatedAt: now},
			Phone:     sql.NullString{String: req.Phone, Valid: req.Phone != ""},
			FirstName: sql.NullString{String: req.FirstName, Valid: req.FirstName != ""},
			LastName:  sql.NullString{String: req.LastName, Valid: req.LastName != ""},
			Address:   sql.NullString{String: req.Address, Valid: req.Address != ""},
		}
		if req.DateBirth != "" {
			dt, err := time.Parse("2006-01-02", req.DateBirth)
			if err != nil {
				return appErrors.NewAppError("INVALID_FORMAT", "invalid date format", err)
			}
			profile.DateBirth = sql.NullTime{Time: dt, Valid: true}
		}
		return uc.customerRepo.UpdateProfile(ctx, profile)

	case "seller":
		req, ok := payload.(dto.SellerProfileRequest)
		if !ok {
			return appErrors.NewAppError("INVALID_PAYLOAD", "invalid seller payload", errors.New("type mismatch"))
		}
		if err := uc.validator.Struct(req); err != nil {
			return appErrors.NewAppError("VALIDATION", "invalid seller profile data", err)
		}

		profile := &entity.SellerProfile{
			User:        entity.User{ID: userID, UpdatedAt: now},
			CompanyName: sql.NullString{String: req.CompanyName, Valid: req.CompanyName != ""},
			Rating:      sql.NullFloat64{Float64: req.Rating, Valid: true},
		}
		return uc.sellerRepo.UpdateProfile(ctx, profile)

	default:
		return appErrors.NewAppError("INVALID_TYPE", "unsupported user type", nil)
	}
}

func (uc *authUsecase) DeleteUser(ctx context.Context, userID string) error {
	if err := uc.revokeRefreshToken(ctx, userID); err != nil && !errors.Is(err, appErrors.ErrNotFound) {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	if err := uc.userRepo.Delete(ctx, userID); err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return appErrors.NewAppError("NOT_FOUND", "user not found", err)
		}
		return appErrors.NewAppError("DELETE_FAIL", "failed to delete user", err)
	}

	uc.logger.WithField("user_id", userID).Info("user deleted")
	return nil
}

func (uc *authUsecase) revokeRefreshToken(ctx context.Context, userID string) error {
	t, err := uc.tokenRepo.GetRefreshTokenByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return nil
		}
		return err
	}
	t.IsRevoked = true
	t.UpdatedAt = time.Now()
	return uc.tokenRepo.UpsertRefreshToken(ctx, t)
}
