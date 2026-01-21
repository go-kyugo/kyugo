package dto

import "time"

type Product struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Price     float64   `json:"price"`
    CreatedAt time.Time `json:"created_at"`
}

type CreateProductRequest struct {
    Name  string  `json:"name" validate:"required"`
    Price float64 `json:"price" validate:"required,gt=0"`
}

type UpdateProductRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}