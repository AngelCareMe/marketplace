package token

import (
	"context"
	"marketplace/internal/entity"
)

type TokenRepository interface {
	GetRefreshTokenByUserID(ctx context.Context, user_id string) (*entity.RefreshToken, error)
	UpsertRefreshToken(ctx context.Context, token *entity.RefreshToken) error
}
