package user

import (
	"context"
	"marketplace/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, customer *entity.User) error
	GetByID(ctx context.Context, userID string) (*entity.User, error)
	UpdateAuth(ctx context.Context, id string, username, email, password string) error
	Delete(ctx context.Context, id string) error
}
