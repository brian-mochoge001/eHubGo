package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"ehubgo/db"
	"ehubgo/middleware"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
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

	// Initialize Firebase Admin SDK
	var opts []option.ClientOption
	if serviceAccountJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON"); serviceAccountJSON != "" {
		fmt.Println("Initializing Firebase with JSON from environment variable")
		opts = append(opts, option.WithCredentialsJSON([]byte(serviceAccountJSON)))
	} else if serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH"); serviceAccountPath != "" {
		fmt.Printf("Initializing Firebase with file from path: %s\n", serviceAccountPath)
		opts = append(opts, option.WithCredentialsFile(serviceAccountPath))
	} else {
		fmt.Println("Warning: No Firebase credentials provided via environment variables. Falling back to default credentials.")
	}

	fbApp, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		fmt.Printf("warning: error initializing firebase app: %v. Continuing with local DB only.\n", err)
	}

	var authClient *auth.Client
	if fbApp != nil {
		authClient, err = fbApp.Auth(ctx)
		if err != nil {
			fmt.Printf("warning: error getting Auth client: %v\n", err)
		}
	}

	// Admin Credentials
	email := "brinokemwa001@gmail.com"
	password := "7798Ray-2004"
	firstName := "Executive"
	lastName := "Admin"

	fmt.Printf("Seeding executive admin: %s\n", email)

	// 1. Create or Get User in Firebase
	var fbUID string
	if authClient != nil {
		fbUser, err := authClient.GetUserByEmail(ctx, email)
		if err != nil {
			// User doesn't exist, create them
			params := (&auth.UserToCreate{}).
				Email(email).
				Password(password).
				DisplayName(firstName + " " + lastName)
			
			fbUser, err = authClient.CreateUser(ctx, params)
			if err != nil {
				log.Fatalf("error creating user in Firebase: %v", err)
			}
			fbUID = fbUser.UID
			fmt.Printf("Successfully created user in Firebase: %s\n", fbUID)
		} else {
			fbUID = fbUser.UID
			fmt.Printf("User already exists in Firebase: %s\n", fbUID)
		}
	} else {
        // Fallback UID if Firebase is unreachable
        fbUID = "local-dev-user-uuid"
        fmt.Println("Firebase unreachable, using local-dev-user-uuid")
    }

	// 2. Hash password for local database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("error hashing password: %v", err)
	}

	// 3. Create user in local database if not exists
	var dbUser db.User
	dbUser, err = queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			dbUser, err = queries.CreateUser(ctx, db.CreateUserParams{
				ID:                fbUID, // Sync ID with Firebase UID
				Email:             email,
				PasswordHash:      string(hashedPassword),
				FirstName:         firstName,
				LastName:          sql.NullString{String: lastName, Valid: true},
				DateOfBirth:       sql.NullTime{Time: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				PhoneNumber:       sql.NullString{String: "", Valid: false},
				ProfilePictureUrl: sql.NullString{String: "", Valid: false},
			})
			if err != nil {
				log.Fatalf("error creating user in database: %v", err)
			}
			fmt.Printf("Successfully created user in database: %s\n", dbUser.ID)
		} else {
			log.Fatalf("error checking user in database: %v", err)
		}
	} else {
		fmt.Printf("User already exists in database: %s\n", dbUser.ID)
	}

	// 4. Assign Executive Admin Role
	_, err = queries.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID: dbUser.ID,
		Role:   db.UserRoleTypeExecutiveAdmin,
	})
	if err != nil {
		fmt.Printf("Note: Role assignment might already exist or failed: %v\n", err)
	}

	// 5. Seed Casbin RBAC Policies into the database via enforcer
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
		{"p", "admin", "/api/v1/categories", "*"},
		{"p", "admin", "/api/v1/brands", "*"},
		{"p", "admin", "/api/v1/models", "*"},
		{"p", "admin", "/api/v1/businesses", "GET"},
		{"p", "admin", "/api/v1/businesses/*/status", "PUT"},
		{"p", "staff", "/api/v1/analytics/*", "GET"},
		{"p", "staff", "/api/v1/businesses", "GET"},
		{"p", "vendor", "/api/v1/products", "*"},
		{"p", "vendor", "/api/v1/businesses/me", "GET"},
		{"p", "vendor", "/api/v1/businesses/staff/*", "*"},
		{"p", "store_staff", "/api/v1/products", "GET"},
		{"p", "store_staff", "/api/v1/orders", "GET"},
		{"p", "user", "/api/v1/health", "GET"},
		{"p", "admin", "/api/v1/health", "GET"},
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

	fmt.Println("Essential seeding completed successfully!")
}
