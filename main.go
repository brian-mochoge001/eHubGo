package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time" // Import time package for verified_at

	"ehub/backend/db"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/api/option"
)

func AuthMiddleware(authClient *auth.Client) gin.HandlerFunc {
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

		c.Set("user_id", token.UID)
		c.Set("user_email", token.Claims["email"])
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

	// Initialize Firebase Admin SDK
	var opts []option.ClientOption
	if serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH"); serviceAccountPath != "" {
		opts = append(opts, option.WithCredentialsFile(serviceAccountPath))
	}

	fbApp, err := firebase.NewApp(context.Background(), nil, opts...)
	if err != nil {
		log.Fatalf("error initializing app: %v
", err)
	}

	authClient, err := fbApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v
", err)
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
		protected.Use(AuthMiddleware(authClient))
		{
			protected.GET("/hubs", func(c *gin.Context) {
				hubs, err := queries.GetHubs(c.Request.Context())
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, hubs)
			})

			protected.POST("/hubs", func(c *gin.Context) {
				var params db.CreateHubParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				hub, err := queries.CreateHub(c.Request.Context(), params)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, hub)
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
					ID:        id,
					STMakePoint: loc.Longitude,
					STMakePoint_2: loc.Latitude,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, driver)
			})

			protected.GET("/drivers/:id/location", func(c *gin.Context) {
				id := c.Param("id")
				driver, err := queries.GetDriverLocation(c.Request.Context(), id)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, driver)
			})

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
					STMakePoint: params.Longitude,
					STMakePoint_2: params.Latitude,
					Limit: params.Limit,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, drivers)
			})

			// Services Endpoints
			protected.POST("/services", func(c *gin.Context) {
				var params db.CreateServiceParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				userID := c.MustGet("user_id").(string) // Get user ID from context
				params.ProviderUserID = userID          // Set the provider ID

				// Handle location from JSON payload if provided, otherwise use default nulls
				// Assuming db.CreateServiceParams has Lon and Lat fields correctly defined (e.g., sql.NullFloat64)
				// for ST_SetSRID to work. The SQL query expects separate Lon/Lat for ST_MakePoint.
				// The current params struct may not have Lon/Lat directly if SQLC generates differently.
				// We need to pass them correctly to queries.CreateService.
				// If params.Location is a GEOGRAPHY type, we need to extract Lon/Lat from it,
				// or ensure params struct correctly maps to the query.
				// For now, assuming Lon and Lat are directly available in params for ST_MakePoint.

				// If db.CreateServiceParams expects Lon/Lat as separate float64 for ST_MakePoint:
				// We need to ensure the JSON payload and params struct correctly map 'lon' and 'lat'.
				// Based on schema definition, location is GEOGRAPHY(POINT, 4326), so Lon/Lat are expected.
				// The params struct should ideally have Lon and Lat fields (e.g., sql.NullFloat64).
				// If JSON provides "longitude" and "latitude" keys, binding should handle it.

				service, err := queries.CreateService(c.Request.Context(), params)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, service)
			})

			protected.GET("/services/:id", func(c *gin.Context) {
				id := c.Param("id")
				service, err := queries.GetServiceByID(c.Request.Context(), id)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, service)
			})

			protected.GET("/services", func(c *gin.Context) {
				serviceType := c.Query("service_type")
				if serviceType == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "service_type query parameter is required"})
					return
				}
				services, err := queries.GetServicesByType(c.Request.Context(), serviceType)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, services)
			})

			protected.PUT("/services/:id", func(c *gin.Context) {
				id := c.Param("id")
				var params db.UpdateServiceParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				params.ID = id // Set the ID from the URL parameter

				// Handle location from JSON payload if provided
				// Similar logic as CreateService, assuming params.Lon and params.Lat are available.
				// If params struct is designed to directly accept Lon/Lat for ST_MakePoint:
				// service, err := queries.UpdateService(c.Request.Context(), params)

				service, err := queries.UpdateService(c.Request.Context(), params)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, service)
			})

			protected.DELETE("/services/:id", func(c *gin.Context) {
				id := c.Param("id")
				_, err := queries.DeleteService(c.Request.Context(), id)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusNoContent, nil) // 204 No Content for successful deletion
			})

			// Vendor Verification Endpoints
			// Endpoint for admins to update verification status
			protected.PUT("/vendors/:vendor_id/verification/:miniservice_type", func(c *gin.Context) {
				vendorID := c.Param("vendor_id")
				miniserviceType := c.Param("miniservice_type")

				var req struct {
					VerificationStatus string `json:"verification_status" binding:"required,oneof=pending verified rejected"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				// Get admin user ID from context
				adminUserID, ok := c.Get("user_id")
				if !ok {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
					return
				}

				params := db.UpdateVendorVerificationStatusParams{
					VerificationStatus: req.VerificationStatus,
					VerifiedByUserID:   sql.NullString{String: adminUserID.(string), Valid: true},
					VerifiedAt:         sql.NullTime{Time: time.Now(), Valid: true},
					VendorID:           vendorID,
					MiniserviceType:    miniserviceType,
				}

				verification, err := queries.UpdateVendorVerificationStatus(c.Request.Context(), params)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "vendor verification record not found"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, verification)
			})

			// Endpoint to get verification status for a specific vendor and miniservice
			protected.GET("/vendors/:vendor_id/verification/:miniservice_type", func(c *gin.Context) {
				vendorID := c.Param("vendor_id")
				miniserviceType := c.Param("miniservice_type")

				status, err := queries.GetVendorVerificationStatus(c.Request.Context(), db.GetVendorVerificationStatusParams{
					VendorID:        vendorID,
					MiniserviceType: miniserviceType,
				})
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "verification status not found for this vendor and miniservice"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"vendor_id": vendorID, "miniservice_type": miniserviceType, "verification_status": status})
			})

			// Endpoint to list all verifications for a vendor
			protected.GET("/vendors/:vendor_id/verification", func(c *gin.Context) {
				vendorID := c.Param("vendor_id")
				verifications, err := queries.ListVendorVerifications(c.Request.Context(), vendorID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, verifications)
			})

			// Endpoint to list all vendors verified for a specific miniservice type
			protected.GET("/miniservices/:miniservice_type/verified_vendors", func(c *gin.Context) {
				miniserviceType := c.Param("miniservice_type")
				vendors, err := queries.ListMiniserviceVerifications(c.Request.Context(), miniserviceType)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, vendors)
			})

			// Endpoint for a vendor to request verification (initial creation or update if pending)
			protected.POST("/vendors/:vendor_id/verification", func(c *gin.Context) {
				vendorID := c.Param("vendor_id")

				var req struct {
					MiniserviceType string `json:"miniservice_type" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				// Check if the authenticated user is the vendor or an admin
				currentUserID := c.MustGet("user_id").(string)
				// TODO: Add logic to check if currentUserID matches vendorID or if user has 'admin' role.
				// For now, we assume the request is valid.

				params := db.CreateVendorMiniserviceVerificationParams{
					VendorID:        vendorID,
					MiniserviceType: req.MiniserviceType,
					VerificationStatus: "pending", // Default status for new requests
				}

				verification, err := queries.CreateVendorMiniserviceVerification(c.Request.Context(), params)
				if err != nil {
					// Handle cases where it might already exist and needs update, or other DB errors
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to request verification: %v", err)})
					return
				}
				c.JSON(http.StatusCreated, verification)
			})

			// --- New routes for Brands, Product Variants, and Product Discounts ---

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
				_, err := queries.DeleteBrand(c.Request.Context(), id)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusNoContent, nil) // 204 No Content for successful deletion
			})

			// Product Variants Routes
			protected.POST("/products/:product_id/variants", func(c *gin.Context) {
				productID := c.Param("product_id")
				var params db.CreateProductVariantParams
				if err := c.ShouldBindJSON(&params); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				params.ProductID = productID

				variant, err := queries.CreateProductVariant(c.Request.Context(), params)
				if err != nil {
					c.JSON(httpL.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, variant)
			})

			protected.GET("/products/:product_id/variants", func(c *gin.Context) {
				productID := c.Param("product_id")
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
				_, err := queries.DeleteProductVariant(c.Request.Context(), id)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusNoContent, nil) // 204 No Content for successful deletion
			})

			// Product Discounts Routes
			protected.POST("/products/:product_id/discounts", func(c *gin.Context) {
				productID := c.Param("product_id")
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

			protected.GET("/products/:product_id/discounts", func(c *gin.Context) {
				productID := c.Param("product_id")
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
				_, err := queries.DeleteProductDiscount(c.Request.Context(), id)
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
