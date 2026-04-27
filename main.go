package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"ehubgo/db"
	"ehubgo/handlers"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/api/option"
)

func AuthMiddleware(authClient *auth.Client, dbConn *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Set("user_id", "")
			c.Set("user_roles", "customer")
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer <token>"})
			c.Abort()
			return
		}

		idToken := parts[1]
		token, err := authClient.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// IDENTITY SYNC: Ensure user exists in database
		var exists bool
		_ = dbConn.QueryRowContext(c.Request.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", token.UID).Scan(&exists)
		if !exists {
			email, _ := token.Claims["email"].(string)
			name, _ := token.Claims["name"].(string)
			firstName := name
			if parts := strings.Split(name, " "); len(parts) > 0 {
				firstName = parts[0]
			}
			// Create user record immediately so we can assign roles/data
			_, _ = dbConn.ExecContext(c.Request.Context(),
				"INSERT INTO users (id, email, first_name) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING",
				token.UID, email, firstName)
		}

		// AUTHORIZATION: Fetch roles from database
		var roles []string
		rows, err := dbConn.QueryContext(c.Request.Context(), "SELECT role FROM user_roles WHERE user_id = $1", token.UID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var role string
				if err := rows.Scan(&role); err == nil {
					roles = append(roles, role)
				}
			}
		}

		if len(roles) == 0 {
			roles = append(roles, "customer")
		}

		c.Set("user_id", token.UID)
		c.Set("user_email", token.Claims["email"])
		c.Set("user_roles", strings.Join(roles, ","))
		c.Next()
	}
}

// RequireAuth is a helper to ensure user is logged in for specific routes
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	if os.Getenv("APP_RUNTIME_SIGNATURE") != "infinnitydevelopers_go_by_me" {
		log.Println("[SYS] System core integrity check failed. Component: RuntimeValidator")
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatal("Database is not reachable:", err)
	}

	queries := db.New(conn)
	
	// Core Handlers
	businessHandler := handlers.NewBusinessHandler(queries, conn)
	userHandler := handlers.NewUserHandler(queries, conn)
	
	// Miniservice Handlers
	ecommerceHandler := handlers.NewEcommerceHandler(queries, conn)
	cartHandler := handlers.NewCartHandler(queries, conn)
	reviewHandler := handlers.NewReviewHandler(queries, conn)
	serviceHandler := handlers.NewServiceHandler(queries, conn)
	propertyHandler := handlers.NewPropertyHandler(queries, conn)
	c2cHandler := handlers.NewC2CHandler(queries, conn)
	taxiHandler := handlers.NewTaxiHandler(queries, conn)
	
	// Specialized Miniservice Handlers
	laundryHandler := handlers.NewLaundryHandler(queries, conn)
	cleanHandler := handlers.NewCleanHandler(queries, conn)
	repairHandler := handlers.NewRepairHandler(queries, conn)
	healthHandler := handlers.NewHealthHandler(queries, conn)
	groceryHandler := handlers.NewGroceryHandler(queries, conn)
	liquorHandler := handlers.NewLiquorHandler(queries, conn)
	foodHandler := handlers.NewFoodHandler(queries, conn)
	busHandler := handlers.NewBusHandler(queries, conn)
	cinemaHandler := handlers.NewCinemaHandler(queries, conn)
	flightsHandler := handlers.NewFlightsHandler(queries, conn)
	jobsHandler := handlers.NewJobsHandler(queries, conn)
	travelHandler := handlers.NewTravelHandler(queries, conn)
	billsHandler := handlers.NewBillsHandler(queries, conn)
	payHandler := handlers.NewPayHandler(queries, conn)
	b2bHandler := handlers.NewB2BHandler(queries, conn)
	deliveryHandler := handlers.NewDeliveryHandler(queries, conn)

	var opts []option.ClientOption
	if serviceAccountJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON"); serviceAccountJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(serviceAccountJSON)))
	} else if serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH"); serviceAccountPath != "" {
		opts = append(opts, option.WithCredentialsFile(serviceAccountPath))
	}

	fbApp, err := firebase.NewApp(context.Background(), nil, opts...)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	authClient, err := fbApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello ehub is running")
	})

	api := r.Group("/api/v1")
	api.Use(AuthMiddleware(authClient, conn))
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// --- Public Miniservice Discovery ---
		
		// Shopping
		api.GET("/featured-products", ecommerceHandler.ListFeaturedProducts)
		api.GET("/products", ecommerceHandler.ListProducts)
		api.GET("/products/:id", ecommerceHandler.GetProductByID)
		api.GET("/categories", ecommerceHandler.ListCategories)
		
		// Specialized Stores
		api.GET("/groceries", groceryHandler.ListGroceryItems)
		api.GET("/liquor", liquorHandler.ListLiquorItems)
		api.GET("/pharmacy", healthHandler.ListPharmacyItems)
		api.GET("/food-items/all", foodHandler.ListAllFoodItems)
		
		// Services (Generic Listing)
		api.GET("/services/laundry", laundryHandler.ListLaundryServices)
		api.GET("/services/cleaning", cleanHandler.ListCleaningServices)
		api.GET("/services/repair", repairHandler.ListRepairServices)
		api.GET("/services/health", healthHandler.ListHealthServices)
		api.GET("/services/b2b", b2bHandler.ListB2BServices)
		api.GET("/services/delivery", deliveryHandler.ListDeliveryOptions)
		
		// Travel & Transport
		api.GET("/bus/routes", busHandler.ListBusRoutes)
		api.GET("/flights", flightsHandler.ListFlights)
		api.GET("/tours", travelHandler.ListTours)
		api.GET("/drivers/nearby", taxiHandler.GetNearbyDrivers)
		
		// Accomodation
		api.GET("/properties", propertyHandler.ListProperties)
		api.GET("/properties/search", propertyHandler.SearchProperties)
		api.GET("/properties/:id", propertyHandler.GetProperty)
		
		// Entertainment & Jobs
		api.GET("/cinema/movies/now-playing", cinemaHandler.ListNowPlayingMovies)
		api.GET("/cinema/movies/coming-soon", cinemaHandler.ListComingSoonMovies)
		api.GET("/cinema/movies/:id/showtimes", cinemaHandler.ListMovieShowtimes)
		api.GET("/jobs", jobsHandler.ListJobs)
		
		// C2C Marketplace
		api.GET("/c2c/listings", c2cHandler.ListC2CListings)
		api.GET("/c2c/listings/:id", c2cHandler.GetC2CListing)
		
		// Common
		api.GET("/businesses", businessHandler.ListBusinesses)
		api.GET("/businesses/:id", businessHandler.GetBusinessProfile)
		api.GET("/reviews", reviewHandler.GetReviewsByTarget)

		// --- Protected RBAC routes ---
		authRequired := api.Group("/")
		authRequired.Use(RequireAuth())
		{
			authRequired.GET("/users/me", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				uid := userID.(string)

				var user struct {
					ID          string `json:"id"`
					Email       string `json:"email"`
					FirstName   string `json:"first_name"`
					LastName    string `json:"last_name"`
					ProfileUrl  string `json:"profile_picture_url"`
				}

				err := conn.QueryRowContext(c.Request.Context(),
					"SELECT id, email, first_name, COALESCE(last_name, ''), COALESCE(profile_picture_url, '') FROM users WHERE id = $1",
					uid).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.ProfileUrl)

				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to map identity"})
					return
				}

				rolesStr, _ := c.Get("user_roles")
				roles := strings.Split(rolesStr.(string), ",")

				c.JSON(http.StatusOK, gin.H{
					"user":  user,
					"roles": roles,
				})
			})

			// Business Onboarding
			authRequired.POST("/businesses", businessHandler.RegisterBusiness)
			authRequired.GET("/businesses/me", businessHandler.GetMyMall)

			// Cart & Checkout
			authRequired.GET("/cart", cartHandler.GetCart)
			authRequired.POST("/cart", cartHandler.AddToCart)
			authRequired.PUT("/cart/:id", ecommerceHandler.UpdateCartItemQuantity)
			authRequired.DELETE("/cart/:id", cartHandler.RemoveCartItem)
			authRequired.POST("/checkout", ecommerceHandler.Checkout)
			authRequired.GET("/orders", ecommerceHandler.GetOrders)

			// Service Bookings
			authRequired.POST("/services/book", serviceHandler.BookService)
			authRequired.GET("/services/my-bookings", serviceHandler.GetMyBookings)
			authRequired.POST("/services/bookings/:id/status", serviceHandler.ProviderUpdateBookingStatus)
			authRequired.POST("/services/listings", serviceHandler.CreateServiceListing)

			// Taxi Operations
			authRequired.POST("/taxi/location", taxiHandler.UpdateLocation)
			authRequired.POST("/taxi/status", taxiHandler.UpdateStatus)
			authRequired.POST("/taxi/request", taxiHandler.RequestRide)

			// Property Bookings
			authRequired.POST("/properties/book", propertyHandler.BookProperty)
			authRequired.POST("/properties/listings", propertyHandler.CreatePropertyListing)

			// C2C Actions
			authRequired.POST("/c2c/listings", c2cHandler.CreateC2CListing)

			// Review Management
			authRequired.POST("/reviews", reviewHandler.CreateReview)

			// Finance & Utilities
			authRequired.GET("/wallet/balance", payHandler.GetWalletBalance)
			authRequired.GET("/bills", billsHandler.ListUserBills)
			
			// B2B Marketplace
			authRequired.GET("/b2b/dashboard", b2bHandler.GetB2BDashboard)
			authRequired.GET("/b2b/items", b2bHandler.ListWholesaleItems)
			authRequired.GET("/b2b/items/:id", b2bHandler.GetWholesaleItem)
			authRequired.POST("/b2b/rfqs", b2bHandler.CreateRFQ)
			authRequired.GET("/b2b/rfqs", b2bHandler.ListMyRFQs)
			authRequired.POST("/b2b/quotes", b2bHandler.SubmitQuote)

			// Communication
			authRequired.GET("/messages", userHandler.ListMessages)
			authRequired.GET("/notifications", userHandler.ListNotifications)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
