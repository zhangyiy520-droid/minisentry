package main

import (
	"database/sql"
	"fmt"
	"log"
	
	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgres://postgres:password@127.0.0.1:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open connection:", err)
	}
	defer db.Close()
	
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	
	rows, err := db.Query("SELECT datname FROM pg_database WHERE datname = 'minisentry'")
	if err != nil {
		log.Fatal("Failed to query:", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal("Failed to scan:", err)
		}
		fmt.Printf("Found database: %s\n", name)
	}
	
	fmt.Println("Test completed")
}