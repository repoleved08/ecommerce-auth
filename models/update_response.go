package models

type UpdateResponse struct {
	Product ProductDetails `json:"product"`
	Message string         `json:"message"`
}

type ProductDetails struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}
