package entity

import "time"

type Product struct {
	ID          string    `db:"id" json:"id"`
	SellerID    string    `db:"seller_id" json:"seller_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Price       float64   `db:"price" json:"price"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	CategoryID  string    `db:"category_id" json:"category_id"`
	IsActive    bool      `db:"is_active" json:"is_active"`
}

type ProductImage struct {
	ID        string    `db:"id" json:"id"`
	ProductID string    `db:"product_id" json:"product_id"`
	URL       string    `db:"url" json:"url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Category struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
