package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"ehubgo/db"
	"ehubgo/middleware"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	// Connect to database
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	ctx := context.Background()

	// Seed System User (Owner for in-house business)
	_, err = queries.CreateUser(ctx, db.CreateUserParams{
		ID:        "system",
		Email:     "system@ehub.internal",
		FirstName: "eHub",
		LastName:  sql.NullString{String: "System", Valid: true},
	})
	if err != nil {
		log.Printf("System user might already exist: %v", err)
	}

	// Seed In-House Business for system products
	_, err = queries.CreateBusiness(ctx, db.CreateBusinessParams{
		ID:              "in-house",
		OwnerID:         "system",
		Name:            "eHub Mall (In-House)",
		MiniserviceType: "ecommerce",
	})
	if err != nil {
		log.Printf("In-house business might already exist or failed: %v", err)
	}

	// Seed Casbin RBAC Policies into the database via enforcer
	enforcer, err := middleware.NewCasbinEnforcer(conn)
	if err != nil {
		log.Fatalf("error initializing enforcer for seeding: %v", err)
	}

	// Clear existing policies
	enforcer.ClearPolicy()

	// Define policies
	policies := [][]string{
		{"g", "executive_admin", "admin"},
		{"g", "admin", "staff"},
		{"g", "staff", "user"},

		// Executive Admin (Global Access)
		{"p", "executive_admin", "*", "*"},

		// Admin/Category/Brand/Model Management
		{"p", "admin", "/api/v1/categories", "*"},
		{"p", "admin", "/api/v1/categories/:id", "*"},
		{"p", "admin", "/api/v1/brands", "*"},
		{"p", "admin", "/api/v1/brands/:id", "*"},
		{"p", "admin", "/api/v1/models", "*"},
		{"p", "admin", "/api/v1/models/:id", "*"},
		{"p", "admin", "/api/v1/businesses/*/status", "PUT"},

		// Staff Permissions
		{"p", "staff", "/api/v1/analytics/*", "GET"},
		{"p", "staff", "/api/v1/businesses", "GET"},
		{"p", "staff", "/api/v1/products", "*"},
		{"p", "staff", "/api/v1/products/:id", "*"},
		{"p", "staff", "/api/v1/orders", "*"},
		{"p", "staff", "/api/v1/orders/:id", "*"},
		{"p", "staff", "/api/v1/categories", "*"},
		{"p", "staff", "/api/v1/categories/:id", "*"},
		{"p", "staff", "/api/v1/brands", "*"},
		{"p", "staff", "/api/v1/brands/:id", "*"},
		{"p", "staff", "/api/v1/models", "*"},
		{"p", "staff", "/api/v1/models/:id", "*"},

		// Vendor Permissions
		{"p", "vendor", "/api/v1/products", "*"},
		{"p", "vendor", "/api/v1/products/:id", "*"},
		{"p", "vendor", "/api/v1/businesses/me", "GET"},
		{"p", "vendor", "/api/v1/businesses/staff/*", "*"},

		// User Permissions (Default)
		{"p", "user", "/api/v1/health", "GET"},
		{"p", "user", "/api/v1/featured-products", "GET"},
		{"p", "user", "/api/v1/products", "GET"},
		{"p", "user", "/api/v1/products/:id", "GET"},
		{"p", "user", "/api/v1/categories", "GET"},
		{"p", "user", "/api/v1/categories/:id", "GET"},
		{"p", "user", "/api/v1/brands", "GET"},
		{"p", "user", "/api/v1/models", "GET"},
		{"p", "user", "/api/v1/cart", "*"},
		{"p", "user", "/api/v1/cart/:id", "*"},
		{"p", "user", "/api/v1/checkout", "POST"},
		{"p", "user", "/api/v1/orders", "GET"},
		{"p", "user", "/api/v1/orders/:id", "GET"},
	}

	for _, p := range policies {
		if p[0] == "g" {
			if _, err := enforcer.AddGroupingPolicy(p[1], p[2]); err != nil {
				log.Printf("failed to add grouping policy %v: %v", p, err)
			}
		} else if p[0] == "p" {
			if _, err := enforcer.AddPolicy(p[1], p[2], p[3]); err != nil {
				log.Printf("failed to add policy %v: %v", p, err)
			}
		}
	}
	if err := enforcer.SavePolicy(); err != nil {
		log.Fatalf("failed to save policies: %v", err)
	}

	fmt.Println("RBAC Policy and In-House seeding completed successfully!")
}
