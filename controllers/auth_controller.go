package controllers

import (
	"ecommerce-auth/config"
	"ecommerce-auth/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	// hash password
	hasshedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hasshedPassword)
	// Insert user to the database
	_, err = config.DB.Exec("INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4)", user.Username, user.Email, user.Password, user.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
	fmt.Fprint(w, "User Registered")
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	// retrieving stored user
	var storedUser models.User
	err := config.DB.QueryRow("SELECT id, username, email, password, role FROM users WHERE username=$1", user.Username).Scan(&storedUser.ID, &storedUser.Username, &storedUser.Email, &storedUser.Password, &storedUser.Role)
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	// compare password with bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	// Generate jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": storedUser.Username,
		"user_id":  storedUser.ID,
		"role":     storedUser.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	response := models.LoginResponse{
		User: models.UserDetails{
			ID:       storedUser.ID,
			Username: storedUser.Username,
			Email:    storedUser.Email,
			Role:     storedUser.Role,
		},
		Token: tokenString,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
