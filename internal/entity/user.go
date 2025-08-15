package entity

import (
	"database/sql"
	"time"
)

type User struct {
	ID           string    `db:"id" json:"id,omitempty"`
	UserType     string    `db:"user_type" json:"user_type,omitempty"`
	Username     string    `db:"username" json:"username,omitempty"`
	PasswordHash string    `db:"password_hash" json:"password_hash,omitempty"`
	Email        string    `db:"email" json:"email,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type CustomerProfile struct {
	User
	FirstName sql.NullString `json:"first_name" db:"first_name"`
	LastName  sql.NullString `json:"last_name" db:"last_name"`
	Phone     sql.NullString `json:"phone" db:"phone"`
	DateBirth sql.NullTime   `json:"date_birth" db:"date_birth"`
	Address   sql.NullString `json:"address" db:"address"`
}

type SellerProfile struct {
	User
	CompanyName sql.NullString  `db:"company_name" json:"company_name,omitempty"`
	Rating      sql.NullFloat64 `db:"rating" json:"rating,omitempty"`
}
