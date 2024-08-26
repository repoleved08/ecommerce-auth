package controllers

import (
	"database/sql"
	"ecommerce-auth/config"
	"ecommerce-auth/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func AddProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product

	// Extracting form values
	product.Name = r.FormValue("name")
	product.Description = r.FormValue("description")
	product.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)

	// Getting the image from the form
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "error uploading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save the image
	fileExt := filepath.Ext(handler.Filename)
	imageName := fmt.Sprintf("%s%s", "product_image_", fileExt)
	imagePath := filepath.Join("uploads", imageName)
	outFile, err := os.Create(imagePath)
	if err != nil {
		http.Error(w, "Error saving image", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()
	io.Copy(outFile, file)

	product.ImageURL = "/uploads/" + imageName

	// Extract user ID from context (JWT middleware should set this)
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	product.CreatedBy = userID.(int)

	// Save product to the database
	_, err = config.DB.Exec("INSERT INTO products (name, description, price, image_url, created_by) VALUES ($1, $2, $3, $4, $5)",
		product.Name, product.Description, product.Price, product.ImageURL, product.CreatedBy)
	if err != nil {
		http.Error(w, "Error saving product", http.StatusInternalServerError)
		return
	}

	// Respond success
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}
	rows, err := config.DB.Query("SELECT id, name, description, price, image_url, created_by FROM products")
	if err != nil {
		http.Error(w, "Error fetching products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.ImageURL, &product.CreatedBy)
		if err != nil {
			http.Error(w, "Error scanning product", http.StatusInternalServerError)
			return
		}
		products = append(products, product)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func GetProductById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	var product models.Product
	err := config.DB.QueryRow("SELECT id, name, description,price, image_url,created_by, created_at FROM products WHERE id=$1", id).Scan(&product.ID, &product.Name, &product.Description, &product.CreatedBy, &product.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error fetching products", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}
