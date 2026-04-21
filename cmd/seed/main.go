package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"ehub/backend/db"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found at ../../.env, trying current directory")
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found")
		}
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
	// Use path relative to cmd/seed directory
	serviceAccountPath := "../../service-account.json"
	opts = append(opts, option.WithCredentialsFile(serviceAccountPath))

	fbApp, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}

	authClient, err := fbApp.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	// Admin Credentials
	email := "brinokemwa001@gmail.com"
	password := "7798Ray-2004"
	firstName := "Executive"
	lastName := "Admin"

	fmt.Printf("Seeding executive admin: %s\n", email)

	// 1. Create or Get User in Firebase
	var fbUser *auth.UserRecord
	fbUser, err = authClient.GetUserByEmail(ctx, email)
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
		fmt.Printf("Successfully created user in Firebase: %s\n", fbUser.UID)
	} else {
		fmt.Printf("User already exists in Firebase: %s\n", fbUser.UID)
	}

	// 2. Hash password for local database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("error hashing password: %v", err)
	}

	// 3. Create user in local database if not exists
	// We use the Firebase UID as the ID in our database for easier lookup
	// Note: The schema has gen_random_uuid() but we can override it if we want, 
	// or just let it generate and link by email. 
	// Let's check if user already exists in DB
	var dbUser db.User
	dbUser, err = queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create user
			// Since sqlc generated CreateUser with positional parameters (Column1, etc.)
			// We match the order: email, password_hash, first_name, last_name, phone_number, profile_picture_url
			dbUser, err = queries.CreateUser(ctx, db.CreateUserParams{
				Column1: email,
				Column2: string(hashedPassword),
				Column3: firstName,
				Column4: lastName,
				Column5: "", // phone
				Column6: "", // profile_pic
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
	err = queries.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID: dbUser.ID,
		Role:   db.UserRoleTypeExecutiveAdmin,
	})
	if err != nil {
		// Check if it's a unique constraint violation (already assigned)
		fmt.Printf("Note: Role assignment might already exist or failed: %v\n", err)
	} else {
		fmt.Println("Successfully assigned executive_admin role.")
	}

	fmt.Println("Seeding completed successfully!")
}
