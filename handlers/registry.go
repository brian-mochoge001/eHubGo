package handlers

import (
    "database/sql"

    "github.com/casbin/casbin/v2"
    "github.com/gin-gonic/gin"

    "ehubgo/cache"
    "ehubgo/db"
    "ehubgo/middleware"
)

type Registry struct {
    Auth       *AuthHandler
    Business   *BusinessHandler
    User       *UserHandler
    Ecommerce  *EcommerceHandler
    Cart       *CartHandler
    Review     *ReviewHandler
    Service    *ServiceHandler
    Property   *PropertyHandler
    Host       *HostHandler
    C2C        *C2CHandler
    Taxi       *TaxiHandler
    Laundry    *LaundryHandler
    Clean      *CleanHandler
    Repair     *RepairHandler
    Health     *HealthHandler
    Grocery    *GroceryHandler
    Liquor     *LiquorHandler
    Food       *FoodHandler
    Bus        *BusHandler
    Cinema     *CinemaHandler
    Flights    *FlightsHandler
    Jobs       *JobsHandler
    Travel     *TravelHandler
    Bills      *BillsHandler
    Pay        *PayHandler
    B2B        *B2BHandler
    Delivery   *DeliveryHandler
    Pricing    *PricingHandler
    Feedback   *FeedbackHandler
    Complaints *ComplaintHandler
    BusinessStaff *BusinessStaffHandler
}

// NewRegistry constructs all handlers used by the application.
func NewRegistry(queries *db.Queries, dbConn *sql.DB, redisStore cache.Store, jwtKey []byte, jwtExpiryMinutes int) *Registry {
    return &Registry{
        Auth:      NewAuthHandler(queries, dbConn, jwtKey, jwtExpiryMinutes),
        Business:  NewBusinessHandler(queries, dbConn),
        User:      NewUserHandler(queries, dbConn),
        Ecommerce: NewEcommerceHandler(queries, dbConn, redisStore),
        Cart:      NewCartHandler(queries, dbConn),
        Review:    NewReviewHandler(queries, dbConn),
        Service:   NewServiceHandler(queries, dbConn),
        Property:  NewPropertyHandler(queries, dbConn),
        Host:      NewHostHandler(queries, dbConn),
        C2C:       NewC2CHandler(queries, dbConn),
        Taxi:      NewTaxiHandler(queries, dbConn),
        Laundry:   NewLaundryHandler(queries, dbConn),
        Clean:     NewCleanHandler(queries, dbConn),
        Repair:    NewRepairHandler(queries, dbConn),
        Health:    NewHealthHandler(queries, dbConn),
        Grocery:   NewGroceryHandler(queries, dbConn, redisStore),
        Liquor:    NewLiquorHandler(queries, dbConn),
        Food:      NewFoodHandler(queries, dbConn),
        Bus:       NewBusHandler(queries, dbConn),
        Cinema:    NewCinemaHandler(queries, dbConn, redisStore),
        Flights:   NewFlightsHandler(queries, dbConn),
        Jobs:      NewJobsHandler(queries, dbConn),
        Travel:    NewTravelHandler(queries, dbConn),
        Bills:     NewBillsHandler(queries, dbConn),
        Pay:       NewPayHandler(queries, dbConn),
        B2B:       NewB2BHandler(queries, dbConn),
        Delivery:  NewDeliveryHandler(queries, dbConn),
        Pricing:   NewPricingHandler(queries, dbConn),
        Feedback:  NewFeedbackHandler(queries, dbConn),
        Complaints: NewComplaintHandler(queries, dbConn),
        BusinessStaff: NewBusinessStaffHandler(queries, dbConn),
    }
}

// RegisterRoutes registers all API routes onto the provided router group.
func RegisterRoutes(api *gin.RouterGroup, reg *Registry, enforcer *casbin.Enforcer, redisStore cache.Store) {
    // Public auth routes
    api.POST("/auth/login", reg.Auth.Login)
    api.POST("/auth/register", reg.Auth.Register)
    
    // Public routes
    api.POST("/feedback", reg.Feedback.SubmitFeedback)
    api.POST("/complaints", reg.Complaints.SubmitComplaint)

    api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

    api.GET("/featured-products", middleware.ChaosMiddleware(), reg.Ecommerce.ListFeaturedProducts)
    api.GET("/products", reg.Ecommerce.ListProducts)
    api.GET("/products/search", middleware.ChaosMiddleware(), reg.Ecommerce.SearchProducts)
    api.GET("/products/:id", reg.Ecommerce.GetProductByID)
    api.GET("/categories", reg.Ecommerce.ListCategories)
    api.GET("/brands", reg.Ecommerce.ListBrands)
    api.GET("/models", reg.Ecommerce.ListProductModels)

    api.GET("/groceries", reg.Grocery.ListGroceryItems)
    api.GET("/grocery/stores", reg.Grocery.SearchGroceryStores)
    api.POST("/groceries/delivery/estimate", reg.Grocery.CalculateGroceryDeliveryQuote)
    api.GET("/liquor", reg.Liquor.ListLiquorItems)
    api.GET("/liquor/stores", reg.Liquor.SearchLiquorStores)
    api.GET("/pharmacy", reg.Health.ListPharmacyItems)
    api.GET("/food-items", reg.Food.ListAllFoodItems)
    api.GET("/food-items/:id", reg.Food.GetFoodItem)
    api.GET("/food-items/category", reg.Food.ListFoodByCategory)

    api.GET("/delivery/drivers/nearby", reg.Food.GetNearbyMotorbikeDrivers)
    api.POST("/food/delivery/estimate", reg.Food.EstimateDelivery)
    api.POST("/delivery/pricing/estimate", reg.Delivery.CalculateDeliveryQuote)
    api.POST("/pricing/estimate", reg.Pricing.GetPriceEstimate)

    api.GET("/services/laundry", reg.Laundry.ListLaundryServices)
    api.GET("/services/cleaning", reg.Clean.ListCleaningServices)
    api.GET("/services/repair", reg.Repair.ListRepairServices)
    api.GET("/services/health", reg.Health.ListHealthServices)
    api.GET("/services/b2b", reg.B2B.ListB2BServices)
    api.GET("/services/delivery", reg.Delivery.ListDeliveryOptions)

    api.GET("/bus/routes", reg.Bus.ListBusRoutes)
    api.GET("/flights", reg.Flights.ListFlights)
    api.GET("/tours", reg.Travel.ListTours)
    api.GET("/drivers/nearby", reg.Taxi.GetNearbyDrivers)

    api.GET("/properties", reg.Host.ListProperties)
    api.GET("/properties/:id", reg.Host.GetProperty)

    api.GET("/cinema/movies/now-playing", reg.Cinema.ListNowPlayingMovies)
    api.GET("/cinema/movies/coming-soon", reg.Cinema.ListComingSoonMovies)
    api.GET("/cinema/movies/:id", reg.Cinema.GetMovieDetails)
    api.GET("/cinema/refreshments", reg.Cinema.ListRefreshments)
    api.GET("/cinema/movies/:id/showtimes", reg.Cinema.ListMovieShowtimes)
    api.GET("/jobs", reg.Jobs.ListJobs)

    api.GET("/c2c/listings", reg.C2C.ListC2CListings)
    api.GET("/c2c/listings/:id", reg.C2C.GetC2CListing)

    api.GET("/businesses", reg.Business.ListBusinesses)
    api.GET("/businesses/:id", reg.Business.GetBusinessProfile)
    api.GET("/reviews", reg.Review.GetReviewsByTarget)

    // Protected routes
    authRequired := api.Group("/")
    authRequired.Use(middleware.RequireAuth())
    {
        // Analytics
        authRequired.GET("/analytics/miniservices", reg.Ecommerce.GetMiniserviceAnalytics)
        authRequired.GET("/analytics/stores/:id", reg.Ecommerce.GetStoreAnalytics)

        authRequired.GET("/users", reg.User.ListUsers)
        authRequired.GET("/users/me", reg.User.GetMe)

        authRequired.POST("/businesses", reg.Business.RegisterBusiness)
        authRequired.GET("/businesses/me", reg.Business.GetMyMall)
        authRequired.POST("/businesses/staff/invite", reg.BusinessStaff.InviteStaff)
        authRequired.GET("/businesses/:id/staff", reg.BusinessStaff.ListStaff)

        authRequired.POST("/products", reg.Ecommerce.CreateProduct)
        authRequired.PUT("/products/:id", reg.Ecommerce.UpdateProduct)
        authRequired.DELETE("/products/:id", reg.Ecommerce.DeleteProduct)

        authRequired.POST("/categories", reg.Ecommerce.CreateCategory)
        authRequired.PUT("/categories/:id", reg.Ecommerce.UpdateCategory)
        authRequired.DELETE("/categories/:id", reg.Ecommerce.DeleteCategory)

        authRequired.POST("/brands", reg.Ecommerce.CreateBrand)
        authRequired.PUT("/brands/:id", reg.Ecommerce.UpdateBrand)
        authRequired.DELETE("/brands/:id", reg.Ecommerce.DeleteBrand)

        // Attribute Management
        authRequired.GET("/attributes", reg.Ecommerce.ListAttributes)
        authRequired.POST("/attributes", reg.Ecommerce.CreateAttribute)
        authRequired.POST("/attributes/:id/values", reg.Ecommerce.AddAttributeValue)

        authRequired.POST("/models", reg.Ecommerce.CreateProductModel)
        authRequired.PUT("/models/:id", reg.Ecommerce.UpdateProductModel)
        authRequired.DELETE("/models/:id", reg.Ecommerce.DeleteProductModel)

        authRequired.GET("/cart", reg.Cart.GetCart)
        authRequired.POST("/cart", reg.Cart.AddToCart)
        authRequired.PUT("/cart/:id", reg.Ecommerce.UpdateCartItemQuantity)
        authRequired.DELETE("/cart/:id", reg.Cart.RemoveCartItem)
        authRequired.POST("/checkout", reg.Ecommerce.Checkout)
        authRequired.GET("/orders", reg.Ecommerce.GetOrders)

        authRequired.POST("/services/book", reg.Service.BookService)
        authRequired.GET("/services/my-bookings", reg.Service.GetMyBookings)
        authRequired.POST("/services/bookings/:id/status", reg.Service.ProviderUpdateBookingStatus)
        authRequired.POST("/services/listings", reg.Service.CreateServiceListing)

        authRequired.POST("/taxi/location", reg.Taxi.UpdateLocation)
        authRequired.GET("/taxi/location", reg.Taxi.GetDriverLocation)
        authRequired.POST("/taxi/status", reg.Taxi.UpdateStatus)
        authRequired.GET("/taxi/driver/tasks", reg.Taxi.GetDriverTasks)
        authRequired.GET("/taxi/requests", reg.Taxi.GetRideRequests)
        authRequired.GET("/delivery/requests", reg.Taxi.GetDeliveryRequests)
        authRequired.POST("/delivery/driver/accept", reg.Taxi.AcceptDeliveryRequest)
        authRequired.POST("/delivery/driver/decline", reg.Taxi.DeclineDeliveryRequest)
        authRequired.POST("/taxi/driver/accept", reg.Taxi.AcceptRideRequest)
        authRequired.POST("/taxi/driver/decline", reg.Taxi.DeclineRideRequest)
        authRequired.POST("/taxi/request", reg.Taxi.RequestRide)
        authRequired.GET("/ws/driver", reg.Taxi.DriverWS)
        authRequired.POST("/food/delivery/assign", reg.Food.AssignDelivery)

        authRequired.POST("/properties/book", reg.Property.BookProperty)
        authRequired.POST("/properties/listings", reg.Property.CreatePropertyListing)

        authRequired.POST("/c2c/listings", reg.C2C.CreateC2CListing)

        authRequired.POST("/reviews", reg.Review.CreateReview)

        authRequired.GET("/wallet/balance", reg.Pay.GetWalletBalance)
        authRequired.GET("/bills", reg.Bills.ListUserBills)

        authRequired.GET("/b2b/dashboard", reg.B2B.GetB2BDashboard)
        authRequired.GET("/b2b/items", reg.B2B.ListWholesaleItems)
        authRequired.GET("/b2b/items/:id", reg.B2B.GetWholesaleItem)
        authRequired.POST("/b2b/rfqs", reg.B2B.CreateRFQ)
        authRequired.GET("/b2b/rfqs", reg.B2B.ListMyRFQs)
        authRequired.POST("/b2b/quotes", reg.B2B.SubmitQuote)

        authRequired.GET("/messages", reg.User.ListMessages)
        authRequired.GET("/notifications", reg.User.ListNotifications)
    }

    // Admin routes with RBAC
    adminApi := api.Group("/admin")
    adminApi.Use(middleware.RBACMiddleware(enforcer))
    adminApi.Use(middleware.RequireRole("executive_admin", "admin"))
    {
        adminApi.GET("/status", func(c *gin.Context) { c.JSON(200, gin.H{"status": "admin access granted"}) })
    }
}
