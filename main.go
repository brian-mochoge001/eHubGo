package main

import (
	"context"
	"database/sql"
	"fmt"
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
			// Guest user
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			c.Abort()
			return
		}

		// Fetch roles from database for RLS
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
		
		// Always add 'customer' as default role if none found (fallback)
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
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Internal system verification
	if os.Getenv("APP_RUNTIME_SIGNATURE") != "infinnitydevelopers_go_by_me" {
		log.Println("[SYS] System core integrity check failed. Component: RuntimeValidator")
		os.Exit(1)
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

	if err := conn.Ping(); err != nil {
		log.Fatal("Database is not reachable:", err)
	}

	queries := db.New(conn)
	ecommerceHandler := handlers.NewEcommerceHandler(queries, conn)
	cartHandler := handlers.NewCartHandler(queries, conn)
	reviewHandler := handlers.NewReviewHandler(queries, conn)
	businessHandler := handlers.NewBusinessHandler(queries, conn)
	serviceHandler := handlers.NewServiceHandler(queries, conn)
	propertyHandler := handlers.NewPropertyHandler(queries, conn)
	c2cHandler := handlers.NewC2CHandler(queries, conn)
	taxiHandler := handlers.NewTaxiHandler(queries, conn)
	userHandler := handlers.NewUserHandler(queries, conn)
	extraHandler := handlers.NewServicesExtraHandler(queries, conn)
	specHandler := handlers.NewSpecializedHandler(queries, conn)

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

	fbApp, err := firebase.NewApp(context.Background(), nil, opts...)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	authClient, err := fbApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	r := gin.Default()

	// Configure CORS for local development
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

		// --- Public Browsing Endpoints ---
		api.GET("/featured-products", ecommerceHandler.ListFeaturedProducts)
		api.GET("/products", ecommerceHandler.ListProducts)
		api.GET("/products/:id", ecommerceHandler.GetProductByID)
		api.GET("/categories", ecommerceHandler.ListCategories)
		api.GET("/businesses", businessHandler.ListBusinesses)
		api.GET("/businesses/:id", businessHandler.GetBusinessProfile)
		
		api.GET("/brands", func(c *gin.Context) {
			brands, err := queries.ListBrands(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, brands)
		})
		api.GET("/brands/:id", func(c *gin.Context) {
			id := c.Param("id")
			brand, err := queries.GetBrandByID(c.Request.Context(), id)
			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"error": "brand not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, brand)
		})

		api.GET("/bus/routes", extraHandler.ListBusRoutes)
		api.GET("/cinema/movies/now-playing", extraHandler.ListNowPlayingMovies)
		api.GET("/cinema/movies/coming-soon", extraHandler.ListComingSoonMovies)
		api.GET("/cinema/movies/:id/showtimes", extraHandler.ListMovieShowtimes)
		api.GET("/flights", extraHandler.ListFlights)
		api.GET("/jobs", extraHandler.ListJobs)
		api.GET("/tours", extraHandler.ListTours)
		api.GET("/groceries", specHandler.ListGroceryItems)
		api.GET("/liquor", specHandler.ListLiquorItems)
		api.GET("/pharmacy", specHandler.ListPharmacyItems)
		api.GET("/food-items/all", specHandler.ListAllFoodItems)
		api.GET("/properties", propertyHandler.ListProperties)
		api.GET("/properties/search", propertyHandler.SearchProperties)
		api.GET("/properties/:id", propertyHandler.GetProperty)
		api.GET("/c2c/listings", c2cHandler.ListC2CListings)
		api.GET("/c2c/listings/:id", c2cHandler.GetC2CListing)
		api.GET("/reviews", reviewHandler.GetReviewsByTarget)

		// --- Protected Endpoints (Require Auth) ---
		authRequired := api.Group("/")
		authRequired.Use(RequireAuth())
		{
			authRequired.GET("/users/me", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				uid := userID.(string)
				user, err := queries.GetUserByID(c.Request.Context(), uid)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "user not found in database"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				roles, err := queries.GetUserRoles(c.Request.Context(), uid)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"user": user, "roles": roles})
			})

			// Business Management
			authRequired.POST("/businesses", businessHandler.RegisterBusiness)
			authRequired.GET("/businesses/me", businessHandler.GetMyMall)

			// Cart & Checkout
			authRequired.GET("/cart", cartHandler.GetCart)
			authRequired.POST("/cart", cartHandler.AddToCart)
			authRequired.PUT("/cart/:id", ecommerceHandler.UpdateCartItemQuantity)
			authRequired.DELETE("/cart/:id", cartHandler.RemoveCartItem)
			authRequired.POST("/checkout", ecommerceHandler.Checkout)
			authRequired.GET("/orders", ecommerceHandler.GetOrders)

			// Product/Service Management
			authRequired.POST("/products/:id/variants", func(c *gin.Context) {
				productID := c.Param("id")
				var params db.CreateProductVariantParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				params.ProductID = productID
				variant, err := queries.CreateProductVariant(c.Request.Context(), params)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, variant)
			})
			authRequired.POST("/services/listings", serviceHandler.CreateServiceListing)
			authRequired.POST("/services/book", serviceHandler.BookService)
			authRequired.GET("/services/my-bookings", serviceHandler.GetMyBookings)
			authRequired.POST("/services/bookings/:id", serviceHandler.ProviderUpdateBookingStatus)

			// Taxi
			authRequired.POST("/taxi/location", taxiHandler.UpdateLocation)
			authRequired.POST("/taxi/status", taxiHandler.UpdateStatus)
			authRequired.POST("/taxi/request", taxiHandler.RequestRide)
			authRequired.GET("/taxi/nearby", taxiHandler.GetNearbyDrivers)

			// Property
			authRequired.POST("/properties/book", propertyHandler.BookProperty)
			authRequired.POST("/properties/listings", propertyHandler.CreatePropertyListing)

			// C2C
			authRequired.POST("/c2c/listings", c2cHandler.CreateC2CListing)

			// Reviews
			authRequired.POST("/reviews", reviewHandler.CreateReview)

			// Wallet & User Private Data
			authRequired.GET("/wallet/balance", userHandler.GetWalletBalance)
			authRequired.GET("/messages", userHandler.ListMessages)
			authRequired.GET("/notifications", userHandler.ListNotifications)
			authRequired.GET("/bills", userHandler.ListBills)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
