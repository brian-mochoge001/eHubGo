package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"ehubgo/db"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/google/uuid"
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
	var dbUser db.User
	dbUser, err = queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			dbUser, err = queries.CreateUser(ctx, db.CreateUserParams{
				Email:             email,
				PasswordHash:      string(hashedPassword),
				FirstName:         firstName,
				LastName:          sql.NullString{String: lastName, Valid: true},
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
	err = queries.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID: dbUser.ID,
		Role:   db.UserRoleTypeExecutiveAdmin,
	})
	if err != nil {
		fmt.Printf("Note: Role assignment might already exist or failed: %v\n", err)
	}

	// 5. Seed Categories
	categories := []string{"Electronics", "Groceries", "Pharmacy", "Food", "Repair", "Cleaning", "Laundry", "Bus", "Cinema", "Flights", "Travel", "Jobs", "Stays"}
	for _, catName := range categories {
		_, _ = queries.CreateHub(ctx, db.CreateHubParams{
			ID:          uuid.New().String(),
			Name:        catName,
			Description: fmt.Sprintf("All services related to %s", catName),
		})
		// Actually create in categories table if needed
		_, _ = conn.ExecContext(ctx, "INSERT INTO categories (id, name) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING", uuid.New().String(), catName)
	}

	// 6. Seed Brands
	brands := []struct{ name, logo string }{
		{"Apple", "https://logo.clearbit.com/apple.com"},
		{"Samsung", "https://logo.clearbit.com/samsung.com"},
		{"Coca-Cola", "https://logo.clearbit.com/cocacola.com"},
	}
	for _, b := range brands {
		_, _ = queries.CreateBrand(ctx, db.CreateBrandParams{
			ID:      uuid.New().String(),
			Name:    b.name,
			LogoUrl: sql.NullString{String: b.logo, Valid: true},
		})
	}

	// 7. Seed Businesses
	businessTypes := []struct{ name, mtype string }{
		{"eHub Grocery Store", "grocery"},
		{"eHub Pharmacy", "pharmacy"},
		{"eHub Eats", "restaurant"},
		{"FixIt Pro", "repair"},
		{"Green Clean", "cleaning"},
		{"Fresh Wash", "laundry"},
		{"Nairobi Bus Express", "bus"},
		{"Galaxy Cinema", "cinema"},
		{"SkyTravel", "flights"},
		{"Safari Tours", "travel"},
		{"Modern Stays", "hotel"},
	}
	
	businessIDs := make(map[string]string)
	for _, bt := range businessTypes {
		biz, err := queries.CreateBusiness(ctx, db.CreateBusinessParams{
			ID:              uuid.New().String(),
			OwnerID:         dbUser.ID,
			Name:            bt.name,
			MiniserviceType: bt.mtype,
			Description:     sql.NullString{String: fmt.Sprintf("The best %s in town", bt.mtype), Valid: true},
		})
		if err == nil {
			businessIDs[bt.mtype] = biz.ID
			// Set as approved
			_, _ = queries.UpdateBusinessStatus(ctx, db.UpdateBusinessStatusParams{
				ID:     biz.ID,
				Status: db.NullBusinessVerificationStatus{BusinessVerificationStatus: db.BusinessVerificationStatusApproved, Valid: true},
			})
		}
	}

	// 8. Seed Specific Items
	// Ecommerce Products
	ecommerceBusinessID := ""
	// Use grocery or restaurant as a fallback if no "ecommerce" type is defined yet
	if id, ok := businessIDs["grocery"]; ok {
		ecommerceBusinessID = id
	}

	if ecommerceBusinessID != "" {
		products := []struct {
			name        string
			price       string
			description string
			image       string
		}{
			{"iPhone 15 Pro", "120000.00", "Titanium design, A17 Pro chip.", "https://images.unsplash.com/photo-1695048133142-1a20484d2569?w=400"},
			{"AirPods Pro (2nd Gen)", "35000.00", "Active Noise Cancellation.", "https://images.unsplash.com/photo-1588423771073-b8903fbb85b5?w=400"},
			{"Samsung Galaxy S24", "110000.00", "Galaxy AI is here.", "https://images.unsplash.com/photo-1610945265064-0e34e5519bbf?w=400"},
		}

		for _, p := range products {
			_, _ = queries.CreateProduct(ctx, db.CreateProductParams{
				ID:            uuid.New().String(),
				BusinessID:    ecommerceBusinessID,
				Name:          p.name,
				Price:         p.price,
				Currency:      "Ksh",
				Description:   sql.NullString{String: p.description, Valid: true},
				StockQuantity: 10,
				Rating:        sql.NullString{String: "4.5", Valid: true},
			})
		}
	}

	// Groceries
	if id, ok := businessIDs["grocery"]; ok {
		_, _ = queries.CreateGroceryItem(ctx, db.CreateGroceryItemParams{
			ID:            uuid.New().String(),
			BusinessID:    id,
			Name:          "Fresh Whole Milk",
			Price:         "150.00",
			Currency:      "Ksh",
			Unit:          sql.NullString{String: "1L", Valid: true},
			StockQuantity: 100,
			IsAvailable:   sql.NullBool{Bool: true, Valid: true},
		})
	}

	// Pharmacy
	if id, ok := businessIDs["pharmacy"]; ok {
		_, _ = queries.CreatePharmacyItem(ctx, db.CreatePharmacyItemParams{
			ID:            uuid.New().String(),
			BusinessID:    id,
			Name:          "Paracetamol",
			Price:         "200.00",
			Currency:      "Ksh",
			StockQuantity: 50,
			IsAvailable:   sql.NullBool{Bool: true, Valid: true},
		})
	}

	// Food
	if id, ok := businessIDs["restaurant"]; ok {
		_, _ = queries.CreateFoodItem(ctx, db.CreateFoodItemParams{
			ID:          uuid.New().String(),
			BusinessID:  id,
			Name:        "Margherita Pizza",
			Price:       "1200.00",
			Currency:    "Ksh",
			IsAvailable: sql.NullBool{Bool: true, Valid: true},
		})
	}

	// Services
	if id, ok := businessIDs["repair"]; ok {
		_, _ = queries.CreateService(ctx, db.CreateServiceParams{
			ID:          uuid.New().String(),
			BusinessID:  id,
			ServiceType: "repair",
			Name:        "iPhone Screen Repair",
			BasePrice:   "5000.00",
			Currency:    "Ksh",
		})
	}

	// Bus
	if id, ok := businessIDs["bus"]; ok {
		_, _ = queries.CreateBusRoute(ctx, db.CreateBusRouteParams{
			ID:             uuid.New().String(),
			BusinessID:     id,
			Origin:         "Nairobi",
			Destination:    "Mombasa",
			DepartureTime:  time.Now().Add(24 * time.Hour),
			Price:          "1500.00",
			Currency:       "Ksh",
			AvailableSeats: 40,
			BusType:        sql.NullString{String: "Executive", Valid: true},
		})
	}

	// Cinema
	_, _ = queries.CreateMovie(ctx, db.CreateMovieParams{
		ID:           uuid.New().String(),
		Title:        "The Batman",
		IsNowPlaying: sql.NullBool{Bool: true, Valid: true},
		Rating:       sql.NullString{String: "4.8", Valid: true},
	})

	// Driver & Taxi
	_, _ = queries.CreateDriver(ctx, db.CreateDriverParams{
		ID:     uuid.New().String(),
		UserID: dbUser.ID,
		Name:   "John Taxi",
		Status: "online",
	})

	fmt.Println("Seeding completed successfully!")
}
