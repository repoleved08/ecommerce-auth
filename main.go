package main

import (
	"ecommerce-auth/config"
	"ecommerce-auth/controllers"
	"ecommerce-auth/middleware"
	"log"
	"net/http"
)



func main() {
	config.InitDB()
	router := http.NewServeMux()
	// routes
	router.HandleFunc("/register", controllers.Register)
	router.HandleFunc("/login", controllers.Login)
	// public routes
	router.HandleFunc("/products", controllers.GetAllProducts)
	router.HandleFunc("/product", controllers.GetProductById)
	// PROTECTED ROUTES
	router.HandleFunc("/products/add", middleware.JWTMiddleware(controllers.AddProduct))
	router.HandleFunc("/products/update", middleware.JWTMiddleware(controllers.UpdateProduct))
	router.HandleFunc("/product/delete", controllers.DeleteProduct)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
