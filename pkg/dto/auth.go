package dto

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	UserType string `json:"user_type" validate:"required,oneof=customer seller"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Username string `json:"username" validate:"omitempty,min=3"`
	Password string `json:"password" validate:"required"`
	UserType string `json:"user_type" validate:"required,oneof=customer seller"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
}

type TokenClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	UserType  string `json:"user_type"`
	IssuedAt  int64  `json:"issued_at"`
	ExpiresAt int64  `json:"expires_at"`
}

type UpdateAuthRequest struct {
	Email        string `json:"email" validate:"omitempty,email"`
	Username     string `json:"username" validate:"omitempty,min=3,max=50"`
	OldPassword  string `json:"old_password" validate:"required_with=NewPassword"`
	NewPassword  string `json:"new_password" validate:"omitempty,min=8"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type CustomerProfileRequest struct {
	Phone     string `json:"phone" validate:"omitempty,e164"`
	FirstName string `json:"first_name" validate:"omitempty,min=2,max=50"`
	LastName  string `json:"last_name" validate:"omitempty,min=2,max=50"`
	Address   string `json:"address" validate:"omitempty"`
	DateBirth string `json:"date_birth" validate:"omitempty,datetime=2006-01-02"` // ISO формат
}

type SellerProfileRequest struct {
	CompanyName string  `json:"company_name" validate:"omitempty,min=2,max=100"`
	Rating      float64 `json:"rating" validate:"omitempty,min=0,max=5"`
}

type CustomerProfileResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address   string `json:"address"`
	DateBirth string `json:"date_birth"`
	UserType  string `json:"user_type"`
}

type SellerProfileResponse struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	CompanyName string  `json:"company_name"`
	Rating      float64 `json:"rating"`
	UserType    string  `json:"user_type"`
}
