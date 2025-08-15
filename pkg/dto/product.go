package dto

type CreateProductRequest struct {
	SellerID    string  `json:"seller_id" validate:"required"`
	CategoryID  string  `json:"category_id" validate:"required"`
	Title       string  `json:"title" validate:"required,min=5,max=20"`
	Description string  `json:"description" validate:"omitempty,max=999"`
	Price       float64 `json:"price" validate:"required,min=0"`
}

type ProductResponse struct {
	SellerID   string  `json:"seller_id" validate:"required"`
	CategoryID string  `json:"category_id" validate:"required"`
	Title      string  `json:"title" validate:"required,min=5,max=20"`
	Price      float64 `json:"price" validate:"required,min=0"`
}

type UpdateProductRequest struct {
	ID          string  `json:"id" validate:"required"`
	CategoryID  string  `json:"category_id" validate:"required"`
	Title       string  `json:"title" validate:"required,min=5,max=20"`
	Description string  `json:"description" validate:"omitempty,max=999"`
	Price       float64 `json:"price" validate:"required,min=0"`
}

type CategoryDTO struct {
	CategoryID string `json:"category_id" validate:"required"`
	Name       string `json:"name" validate:"required,min=1,max=50"`
}

type ImageDTO struct {
	ProductID string `json:"product_id" validate:"required"`
	URL       string `json:"url" validate:"required"`
}
