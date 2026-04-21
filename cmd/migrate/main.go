package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Database is not reachable:", err)
	}

	fmt.Println("Dropping existing schema for a clean RLS implementation...")
	// Drop all tables and types to avoid conflicts with enum changes
	dropCmd := `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO public;
		COMMENT ON SCHEMA public IS 'standard public schema';
	`
	_, err = db.Exec(dropCmd)
	if err != nil {
		log.Fatal("Error dropping schema:", err)
	}

	fmt.Println("Reading schema.sql...")
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatal("Could not read schema.sql:", err)
	}

	fmt.Println("Executing new schema with RLS and Roles...")
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Fatal("Could not start transaction:", err)
	}

	_, err = tx.Exec(string(schema))
	if err != nil {
		tx.Rollback()
		log.Fatal("Error executing schema:", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal("Error committing transaction:", err)
	}

	fmt.Println("Database schema with RLS and Superapp Roles pushed successfully to Neon!")
}
