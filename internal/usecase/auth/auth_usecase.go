package usecase

import (
	"context"
	"marketplace/pkg/dto"
)

type AuthUsecase interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	UpdateAuth(ctx context.Context, tokenString, userID string, req dto.UpdateAuthRequest) error
	UpdateProfile(ctx context.Context, userID string, userType string, payload any) error
	DeleteUser(ctx context.Context, userID string) error
}
