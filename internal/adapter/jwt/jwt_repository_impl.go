package jwt

import (
	"context"
	"fmt"
	"marketplace/internal/adapter/postgres/token"
	"marketplace/internal/entity"
	"marketplace/pkg/config"
	appErrors "marketplace/pkg/errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type jwtManager struct {
	tokenRepo token.TokenRepository
	logger    *logrus.Logger
	cfg       config.Config
}

func NewJWTManager(tokenRepo token.TokenRepository, logger *logrus.Logger, cfg config.Config) *jwtManager {
	return &jwtManager{
		tokenRepo: tokenRepo,
		logger:    logger,
		cfg:       cfg,
	}
}

func (j *jwtManager) GenerateAccessToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"user_type": user.UserType,
		"exp":       time.Now().Add(15 * time.Minute).Unix(),
		"iat":       time.Now().Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(j.cfg.JWT.SecretKey))
}

func (j *jwtManager) ValidateAccessToken(tokenString string) error {
	jwtToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, appErrors.NewAppError("JWT_VALIDATION", "unexpected signing method", nil)
		}
		return []byte(j.cfg.JWT.SecretKey), nil
	})

	if err != nil {
		j.logger.WithFields(logrus.Fields{
			"stage": "parse",
			"token": tokenString,
			"err":   err,
		}).Error("failed to parse access token")

		return appErrors.NewAppError("JWT_VALIDATION", "failed to parse access token", err)
	}

	if !jwtToken.Valid {
		return appErrors.NewAppError("JWT_VALIDATION", "invalid access token", nil)
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return appErrors.NewAppError("JWT_VALIDATION", "failed to parse access token claims", nil)
	}

	if _, ok := claims["user_id"].(string); !ok {
		return appErrors.NewAppError("JWT_VALIDATION", "user_id claim is missing", nil)
	}

	if _, ok := claims["user_type"].(string); !ok {
		return appErrors.NewAppError("JWT_VALIDATION", "user_type claim is missing", nil)
	}

	return nil
}

func (j *jwtManager) GenerateRefreshToken(ctx context.Context, user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"user_type": user.UserType,
		"exp":       time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(j.cfg.JWT.SecretKey))
	if err != nil {
		j.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"err":     err,
		}).Error("failed to sign refresh token")

		return "", appErrors.NewAppError("JWT_GENERATION", "failed to sign refresh token", err)
	}

	refreshToken := &entity.RefreshToken{
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		IsRevoked: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = j.tokenRepo.UpsertRefreshToken(ctx, refreshToken)
	if err != nil {
		j.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"err":     err,
		}).Error("failed to store refresh token in DB")

		return "", appErrors.NewAppError("JWT_DB", "failed to store refresh token", err)
	}

	return tokenString, nil
}

func (j *jwtManager) ValidateRefreshToken(ctx context.Context, tokenString string) error {
	jwtToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, appErrors.NewAppError("JWT_VALIDATION", "unexpected signing method", nil)
		}
		return []byte(j.cfg.JWT.SecretKey), nil
	})

	if err != nil {
		j.logger.WithFields(logrus.Fields{
			"stage": "parse",
			"token": tokenString,
			"err":   err,
		}).Error("failed to parse refresh token")

		return appErrors.NewAppError("JWT_VALIDATION", "failed to parse refresh token", err)
	}

	if !jwtToken.Valid {
		return appErrors.NewAppError("JWT_VALIDATION", "invalid refresh token", nil)
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return appErrors.NewAppError("JWT_VALIDATION", "failed to parse refresh token claims", nil)
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return appErrors.NewAppError("JWT_VALIDATION", "user_id claim is missing or invalid", nil)
	}

	_, ok = claims["user_type"].(string)
	if !ok {
		return appErrors.NewAppError("JWT_VALIDATION", "user_type claim is missing or invalid", nil)
	}

	dbToken, err := j.tokenRepo.GetRefreshTokenByUserID(ctx, userID)
	if err != nil {
		j.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"err":     err,
		}).Error("failed to fetch refresh token from DB")

		return appErrors.NewAppError("JWT_DB", "failed to fetch refresh token", err)
	}

	if dbToken.Token != tokenString {
		return appErrors.NewAppError("JWT_VALIDATION", fmt.Sprintf("refresh token mismatch for user %s", userID), nil)
	}

	if time.Now().After(dbToken.ExpiresAt) {
		return appErrors.NewAppError("JWT_VALIDATION", fmt.Sprintf("refresh token expired for user %s", userID), nil)
	}

	if dbToken.IsRevoked {
		return appErrors.NewAppError("JWT_VALIDATION", fmt.Sprintf("refresh token revoked for user %s", userID), nil)
	}

	return nil
}

func (j *jwtManager) Secret() string {
	return j.cfg.JWT.SecretKey
}
