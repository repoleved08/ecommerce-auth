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
	"time"
)

func AddProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product

	// Extracting form values
	product.Name = r.FormValue("name")
	product.Description = r.FormValue("description")
	product.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)

	// Getting the image from the form
	file, handler, err := r.FormFile("image_url")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create unique image name
	// Using a timestamp to ensure the filename is unique
	timestamp := time.Now().Unix()
	imageName := fmt.Sprintf("%d_%s", timestamp, handler.Filename)
	imagePath := filepath.Join("uploads", imageName)

	// Save the image
	outFile, err := os.Create(imagePath)
	if err != nil {
		http.Error(w, "Error saving image", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	// Copy the file to the server
	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Error copying image", http.StatusInternalServerError)
		return
	}

	// Store the relative path to the image in the database
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

	// Respond with the product data
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}


func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
	err := config.DB.QueryRow("SELECT id, name, description,price, image_url,created_by FROM products WHERE id=$1", id).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.ImageURL, &product.CreatedBy)

	if err == sql.ErrNoRows {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
    if err != nil {
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }

	var product models.Product
	err = json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// extracting role from jwt token
	userRole := r.Context().Value("role").(string)

	//var productName string
	var storedProduct models.Product

	// Check if the product exists
	err = config.DB.QueryRow("SELECT id, name, description, price, image_url FROM products WHERE id=$1", id).Scan(
		&storedProduct.ID, &storedProduct.Name, &storedProduct.Description, &storedProduct.Price, &storedProduct.ImageURL)
	if err == sql.ErrNoRows {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error fetching products", http.StatusInternalServerError)
		return
	}

	// Check if the user is admin
	if userRole != "admin" {
		http.Error(w, "you are not allowed to update this product", http.StatusForbidden)
		return
	}

	// Perform the product update
	_, err = config.DB.Exec("UPDATE products SET name=$1, description=$2, price=$3, image_url=$4 WHERE id=$5",
		product.Name, product.Description, product.Price, product.ImageURL, id)
	if err != nil {
		http.Error(w, "error updating the product", http.StatusInternalServerError)
		return
	}

	// Return success response with the updated product details
	response := models.UpdateResponse{
		Product: models.ProductDetails{
			ID:          storedProduct.ID,
			Name:        product.Name,
			Description: product.Description,
			ImageURL:    product.ImageURL,
			Price:       product.Price,
		},
		Message: "Product Updated Successfully",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
    // Extract product ID from URL and ensure it's a valid integer
    id, err := strconv.Atoi(r.URL.Query().Get("id"))
    if err != nil {
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }

    // Extract user ID from JWT token
    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Extract user role (if applicable) to restrict actions to admins
    userRole := r.Context().Value("role").(string)

    // Verify that the product belongs to the logged-in user or user is admin
    var createdBy int
    err = config.DB.QueryRow("SELECT created_by FROM products WHERE id = $1", id).Scan(&createdBy)
    if err == sql.ErrNoRows {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "Error fetching product", http.StatusInternalServerError)
        return
    }

    // Ensure that only the product owner or admin can delete the product
    if createdBy != userID && userRole != "admin" {
        http.Error(w, "You are not allowed to delete this product", http.StatusForbidden)
        return
    }

    // Delete the product
    _, err = config.DB.Exec("DELETE FROM products WHERE id = $1", id)
    if err != nil {
        http.Error(w, "Error deleting product", http.StatusInternalServerError)
        return
    }

    // Respond with success message
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted successfully"})
}

