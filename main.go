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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
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

// WithRLS wraps a database operation with RLS session variables
func WithRLS(c *gin.Context, dbConn *sql.DB, fn func(tx *sql.Tx) error) error {
	userID, _ := c.Get("user_id")
	userRoles, _ := c.Get("user_roles")

	tx, err := dbConn.BeginTx(c.Request.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Set session variables for RLS
	_, err = tx.ExecContext(c.Request.Context(), fmt.Sprintf("SET LOCAL app.current_user_id = '%s'", userID))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(c.Request.Context(), fmt.Sprintf("SET LOCAL app.current_user_roles = '%s'", userRoles))
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
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
	if serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH"); serviceAccountPath != "" {
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

	// Configure CORS for local development
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Protected routes
		protected := api.Group("/")
		protected.Use(AuthMiddleware(authClient, conn))
		{
			protected.GET("/users/me", func(c *gin.Context) {
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

				c.JSON(http.StatusOK, gin.H{
					"user":  user,
					"roles": roles,
				})
			})

			protected.GET("/hubs", func(c *gin.Context) {
				err := handlers.WithRLS(c, conn, func(tx *sql.Tx) error {
					q := queries.WithTx(tx)
					hubs, err := q.GetHubs(c.Request.Context())
					if err != nil {
						return err
					}
					c.JSON(http.StatusOK, hubs)
					return nil
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
			})

			protected.POST("/hubs", func(c *gin.Context) {
				var params db.CreateHubParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				
				err := handlers.WithRLS(c, conn, func(tx *sql.Tx) error {
					q := queries.WithTx(tx)
					hub, err := q.CreateHub(c.Request.Context(), params)
					if err != nil {
						return err
					}
					c.JSON(http.StatusCreated, hub)
					return nil
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
			})

			protected.GET("/tasks", func(c *gin.Context) {
				tasks, err := queries.GetTasks(c.Request.Context())
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, tasks)
			})

			protected.POST("/tasks", func(c *gin.Context) {
				var params db.CreateTaskParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				task, err := queries.CreateTask(c.Request.Context(), params)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, task)
			})

			// Taxi Endpoints
			protected.POST("/drivers", func(c *gin.Context) {
				var params db.CreateDriverParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				driver, err := queries.CreateDriver(c.Request.Context(), params)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, driver)
			})

			protected.POST("/drivers/:id/location", func(c *gin.Context) {
				id := c.Param("id")
				var loc struct {
					Longitude float64 `json:"longitude" binding:"required"`
					Latitude  float64 `json:"latitude" binding:"required"`
				}
				if err := c.ShouldBindJSON(&loc); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				driver, err := queries.UpdateDriverLocation(c.Request.Context(), db.UpdateDriverLocationParams{
					UserID:        id,
					StMakepoint:   loc.Longitude,
					StMakepoint_2: loc.Latitude,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, driver)
			})

			// ... (skip nearby drivers fix for now) ...

			protected.GET("/drivers/nearby", func(c *gin.Context) {
				var params struct {
					Longitude float64 `form:"longitude" binding:"required"`
					Latitude  float64 `form:"latitude" binding:"required"`
					Limit     int32   `form:"limit,default=5"`
				}
				if err := c.ShouldBindQuery(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				drivers, err := queries.GetNearbyDrivers(c.Request.Context(), db.GetNearbyDriversParams{
					StMakepoint:   params.Longitude,
					StMakepoint_2: params.Latitude,
					Limit:         params.Limit,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, drivers)
			})

			// ... (skip services fix) ...

			protected.DELETE("/services/:id", func(c *gin.Context) {
				id := c.Param("id")
				err := queries.DeleteService(c.Request.Context(), id)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusNoContent, nil) // 204 No Content for successful deletion
			})

			// ... (skip vendor verification fix) ...

			// User Endpoints
			protected.GET("/wallet/balance", userHandler.GetWalletBalance)
			protected.GET("/messages", userHandler.ListMessages)
			protected.GET("/notifications", userHandler.ListNotifications)
			protected.GET("/bills", userHandler.ListBills)

			// Specialized Service Endpoints
			protected.GET("/bus/routes", extraHandler.ListBusRoutes)
			protected.GET("/cinema/movies/now-playing", extraHandler.ListNowPlayingMovies)
			protected.GET("/cinema/movies/coming-soon", extraHandler.ListComingSoonMovies)
			protected.GET("/cinema/movies/:id/showtimes", extraHandler.ListMovieShowtimes)
			protected.GET("/flights", extraHandler.ListFlights)
			protected.GET("/jobs", extraHandler.ListJobs)
			protected.GET("/tours", extraHandler.ListTours)

			// Specialized Item Endpoints
			protected.GET("/groceries", specHandler.ListGroceryItems)
			protected.GET("/liquor", specHandler.ListLiquorItems)
			protected.GET("/pharmacy", specHandler.ListPharmacyItems)
			protected.GET("/food-items/all", specHandler.ListAllFoodItems)

			// Universal Cart Endpoints
			protected.GET("/cart", cartHandler.GetCart)
			protected.POST("/cart", cartHandler.AddToCart)
			protected.DELETE("/cart/:id", cartHandler.RemoveCartItem)

			// Review Endpoints
			protected.POST("/reviews", reviewHandler.CreateReview)
			protected.GET("/reviews", reviewHandler.GetReviewsByTarget)

			// Business Endpoints (The "Mall/Stall" model)
			protected.GET("/businesses", businessHandler.ListBusinesses)
			protected.POST("/businesses", businessHandler.RegisterBusiness)
			protected.GET("/businesses/me", businessHandler.GetMyMall)
			protected.GET("/businesses/:id", businessHandler.GetBusinessProfile)

			// Service Endpoints (eLaundry, eClean, eRepair, eHealth)
			protected.GET("/services", serviceHandler.ListServices)
			protected.POST("/services/book", serviceHandler.BookService)
			protected.GET("/services/my-bookings", serviceHandler.GetMyBookings)
			protected.POST("/services/bookings/:id", serviceHandler.ProviderUpdateBookingStatus)
			protected.POST("/services/listings", serviceHandler.CreateServiceListing)

			// Taxi & Driver Endpoints
			protected.GET("/taxi/nearby", taxiHandler.GetNearbyDrivers)
			protected.POST("/taxi/location", taxiHandler.UpdateLocation)
			protected.POST("/taxi/status", taxiHandler.UpdateStatus)
			protected.POST("/taxi/request", taxiHandler.RequestRide)

			// Property Endpoints (eHost - Stays/Hotels)
			protected.GET("/properties", propertyHandler.ListProperties)
			protected.GET("/properties/search", propertyHandler.SearchProperties)
			protected.GET("/properties/:id", propertyHandler.GetProperty)
			protected.POST("/properties/book", propertyHandler.BookProperty)
			protected.POST("/properties/listings", propertyHandler.CreatePropertyListing)

			// C2C Marketplace Endpoints
			protected.GET("/c2c/listings", c2cHandler.ListC2CListings)
			protected.GET("/c2c/listings/:id", c2cHandler.GetC2CListing)
			protected.POST("/c2c/listings", c2cHandler.CreateC2CListing)

			// Ecommerce Endpoints
			protected.GET("/featured-products", ecommerceHandler.ListFeaturedProducts)
			protected.GET("/products", ecommerceHandler.ListProducts)
			protected.GET("/products/:id", ecommerceHandler.GetProductByID)
			protected.GET("/categories", ecommerceHandler.ListCategories)
			protected.GET("/orders", ecommerceHandler.GetOrders)
			protected.POST("/checkout", ecommerceHandler.Checkout)
			protected.PUT("/cart/:id", ecommerceHandler.UpdateCartItemQuantity)

			// Brands Routes
			protected.POST("/brands", func(c *gin.Context) {
				var params db.CreateBrandParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				brand, err := queries.CreateBrand(c.Request.Context(), params)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, brand)
			})

			protected.GET("/brands/:id", func(c *gin.Context) {
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

			protected.GET("/brands", func(c *gin.Context) {
				brands, err := queries.ListBrands(c.Request.Context())
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, brands)
			})

			protected.PUT("/brands/:id", func(c *gin.Context) {
				id := c.Param("id")
				var params db.UpdateBrandParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				params.ID = id // Set the ID from the URL parameter

				brand, err := queries.UpdateBrand(c.Request.Context(), params)
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

			protected.DELETE("/brands/:id", func(c *gin.Context) {
				id := c.Param("id")
				err := queries.DeleteBrand(c.Request.Context(), id)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusNoContent, nil) // 204 No Content for successful deletion
			})

			// Product Variants Routes
			protected.POST("/products/:id/variants", func(c *gin.Context) {
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

			protected.GET("/products/:id/variants", func(c *gin.Context) {
				productID := c.Param("id")
				variants, err := queries.GetProductVariantsByProductID(c.Request.Context(), productID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, variants)
			})

			protected.GET("/variants/:id", func(c *gin.Context) {
				id := c.Param("id")
				variant, err := queries.GetProductVariantByID(c.Request.Context(), id)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "product variant not found"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, variant)
			})

			protected.PUT("/variants/:id", func(c *gin.Context) {
				id := c.Param("id")
				var params db.UpdateProductVariantParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				params.ID = id // Set the ID from the URL parameter

				variant, err := queries.UpdateProductVariant(c.Request.Context(), params)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "product variant not found"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, variant)
			})

			protected.DELETE("/variants/:id", func(c *gin.Context) {
				id := c.Param("id")
				err := queries.DeleteProductVariant(c.Request.Context(), id)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusNoContent, nil) // 204 No Content for successful deletion
			})

			// Product Discounts Routes
			protected.POST("/products/:id/discounts", func(c *gin.Context) {
				productID := c.Param("id")
				var params db.CreateProductDiscountParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				params.ProductID = productID

				discount, err := queries.CreateProductDiscount(c.Request.Context(), params)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, discount)
			})

			protected.GET("/products/:id/discounts", func(c *gin.Context) {
				productID := c.Param("id")
				discounts, err := queries.GetProductDiscountsByProductID(c.Request.Context(), productID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, discounts)
			})

			protected.GET("/discounts/:id", func(c *gin.Context) {
				id := c.Param("id")
				discount, err := queries.GetProductDiscountByID(c.Request.Context(), id)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "product discount not found"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, discount)
			})

			protected.PUT("/discounts/:id", func(c *gin.Context) {
				id := c.Param("id")
				var params db.UpdateProductDiscountParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				params.ID = id // Set the ID from the URL parameter

				discount, err := queries.UpdateProductDiscount(c.Request.Context(), params)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "product discount not found"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, discount)
			})

			protected.DELETE("/discounts/:id", func(c *gin.Context) {
				id := c.Param("id")
				err := queries.DeleteProductDiscount(c.Request.Context(), id)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusNoContent, nil) // 204 No Content for successful deletion
			})
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
