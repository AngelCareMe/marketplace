package jwt

import (
	"context"
	"marketplace/internal/entity"
)

type JWTManager interface {
	GenerateAccessToken(user *entity.User) (string, error)
	ValidateAccessToken(tokenString string) error
	GenerateRefreshToken(ctx context.Context, user *entity.User) (string, error)
	ValidateRefreshToken(ctx context.Context, tokenString string) error
	Secret() string
}
