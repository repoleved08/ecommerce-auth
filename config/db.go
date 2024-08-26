package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_"github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error
	// Load env
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}
	// Connection string
	connString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", os.Getenv("DB_USER"),os.Getenv("DB_PASSWORD"),os.Getenv("DB_NAME"),os.Getenv("DB_HOST"),os.Getenv("DB_PORT"),os.Getenv("DB_SSLMODE"))
	// Open sql
	DB, err = sql.Open("postgres", connString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	// Ping
	err = DB.Ping()
	if err != nil {
		log.Fatalf("error : %v", err)
	}
	log.Println("Connected to the Database successfully")
}
