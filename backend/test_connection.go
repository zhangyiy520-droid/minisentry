package main

import (
	"log"
	"minisentry/internal/config"
	"minisentry/internal/database"
)

func main() {
	cfg := config.Load()
	log.Printf("Using database URL: %s", cfg.DatabaseURL)
	
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer db.Close()
	
	// Test by querying existing tables
	var count int64
	if err := db.Model(&struct{}{}).Table("users").Count(&count).Error; err != nil {
		log.Fatal("Failed to query users table:", err)
	}
	
	log.Printf("Successfully connected to database. Users table has %d rows", count)
}