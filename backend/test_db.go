package main

import (
	"database/sql"
	"fmt"
	"log"
	
	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgres://postgres:password@127.0.0.1:5432/minisentry?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open connection:", err)
	}
	defer db.Close()
	
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	
	var dbName string
	err = db.QueryRow("SELECT current_database()").Scan(&dbName)
	if err != nil {
		log.Fatal("Failed to query database:", err)
	}
	
	fmt.Printf("Successfully connected to database: %s\n", dbName)
}