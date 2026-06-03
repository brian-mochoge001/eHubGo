package handlers

import (
    "database/sql"

    "github.com/casbin/casbin/v2"
    "github.com/gin-gonic/gin"

    "ehubgo/cache"
    "ehubgo/db"
    "ehubgo/middleware"
    firebase "firebase.google.com/go/v4/auth"
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
    Compliance    *ComplianceHandler
    OrderCoordinator *OrderCoordinator
}

// NewRegistry constructs all handlers used by the application.
func NewRegistry(queries *db.Queries, dbConn *sql.DB, redisStore cache.Store, authClient *firebase.Client) *Registry {
    oc := NewOrderCoordinator(queries, dbConn)
    return &Registry{
        Auth:      NewAuthHandler(queries, dbConn, authClient),
        Business:  NewBusinessHandler(queries, dbConn),
        User:      NewUserHandler(queries, dbConn),
        Ecommerce: NewEcommerceHandler(queries, dbConn, redisStore, oc),
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
        Compliance: NewComplianceHandler(queries, dbConn),
        OrderCoordinator: oc,
    }
}

// RegisterPublicRoutes registers public API routes.
func RegisterPublicRoutes(api *gin.RouterGroup, reg *Registry) {
    api.POST("/feedback", reg.Feedback.SubmitFeedback)
    api.POST("/complaints", reg.Complaints.SubmitComplaint)

    api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

    api.GET("/featured-products", middleware.ChaosMiddleware(), reg.Ecommerce.ListFeaturedProducts)
    api.GET("/products", reg.Ecommerce.ListProducts)
    api.GET("/products/search", middleware.ChaosMiddleware(), reg.Ecommerce.SearchProducts)
    api.GET("/products/filter", reg.Ecommerce.SearchAndFilterProducts)
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
}

// RegisterProtectedRoutes registers protected API routes.
func RegisterProtectedRoutes(api *gin.RouterGroup, reg *Registry, enforcer *casbin.Enforcer) {
    // Protected routes
    
    // Analytics
    api.GET("/analytics/miniservices", reg.Ecommerce.GetMiniserviceAnalytics)
    api.GET("/analytics/stores/:id", reg.Ecommerce.GetStoreAnalytics)

    api.GET("/users", reg.User.ListUsers)
    api.GET("/users/me", reg.User.GetMe)

    api.POST("/businesses", reg.Business.RegisterBusiness)
    api.GET("/businesses/me", reg.Business.GetMyMall)
    api.POST("/businesses/staff/invite", reg.BusinessStaff.InviteStaff)
    api.GET("/businesses/:id/staff", reg.BusinessStaff.ListStaff)
    api.PUT("/businesses/:id/status", middleware.RBACMiddleware(enforcer), reg.Compliance.UpdateApplicationStatus)
    api.PUT("/businesses/documents/:doc_id/status", middleware.RBACMiddleware(enforcer), reg.Compliance.ReviewDocument)

    api.POST("/products", reg.Ecommerce.CreateProduct)
    api.PUT("/products/:id", reg.Ecommerce.UpdateProduct)
    api.DELETE("/products/:id", reg.Ecommerce.DeleteProduct)

    api.POST("/categories", middleware.RBACMiddleware(enforcer), reg.Ecommerce.CreateCategory)
    api.PUT("/categories/:id", middleware.RBACMiddleware(enforcer), reg.Ecommerce.UpdateCategory)
    api.DELETE("/categories/:id", middleware.RBACMiddleware(enforcer), reg.Ecommerce.DeleteCategory)

    api.POST("/brands", middleware.RBACMiddleware(enforcer), reg.Ecommerce.CreateBrand)
    api.PUT("/brands/:id", middleware.RBACMiddleware(enforcer), reg.Ecommerce.UpdateBrand)
    api.DELETE("/brands/:id", middleware.RBACMiddleware(enforcer), reg.Ecommerce.DeleteBrand)

    api.POST("/models", middleware.RBACMiddleware(enforcer), reg.Ecommerce.CreateProductModel)
    api.PUT("/models/:id", middleware.RBACMiddleware(enforcer), reg.Ecommerce.UpdateProductModel)
    api.DELETE("/models/:id", middleware.RBACMiddleware(enforcer), reg.Ecommerce.DeleteProductModel)

    api.GET("/cart", reg.Cart.GetCart)
    api.POST("/cart", reg.Cart.AddToCart)
    api.PUT("/cart/:id", reg.Ecommerce.UpdateCartItemQuantity)
    api.DELETE("/cart/:id", reg.Cart.RemoveCartItem)
    api.POST("/checkout", reg.Ecommerce.Checkout)
    api.GET("/orders", reg.Ecommerce.GetOrders)

    api.POST("/services/book", reg.Service.BookService)
    api.GET("/services/my-bookings", reg.Service.GetMyBookings)
    api.POST("/services/bookings/:id/status", reg.Service.ProviderUpdateBookingStatus)
    api.POST("/services/listings", reg.Service.CreateServiceListing)

    api.POST("/taxi/location", reg.Taxi.UpdateLocation)
    api.GET("/taxi/location", reg.Taxi.GetDriverLocation)
    api.POST("/taxi/status", reg.Taxi.UpdateStatus)
    api.GET("/taxi/driver/tasks", reg.Taxi.GetDriverTasks)
    api.GET("/taxi/requests", reg.Taxi.GetRideRequests)
    api.GET("/delivery/requests", reg.Taxi.GetDeliveryRequests)
    api.POST("/delivery/driver/accept", reg.Taxi.AcceptDeliveryRequest)
    api.POST("/delivery/driver/decline", reg.Taxi.DeclineDeliveryRequest)
    api.POST("/taxi/driver/accept", reg.Taxi.AcceptRideRequest)
    api.POST("/taxi/driver/decline", reg.Taxi.DeclineRideRequest)
    api.POST("/taxi/request", reg.Taxi.RequestRide)
    api.GET("/ws/driver", reg.Taxi.DriverWS)
    api.POST("/food/delivery/assign", reg.Food.AssignDelivery)

    api.POST("/properties/book", reg.Property.BookProperty)
    api.POST("/properties/listings", reg.Property.CreatePropertyListing)

    api.POST("/c2c/listings", reg.C2C.CreateC2CListing)

    api.POST("/reviews", reg.Review.CreateReview)

    api.GET("/wallet/balance", reg.Pay.GetWalletBalance)
    api.POST("/pay/mock", reg.Pay.ProcessMockPayment)
    api.GET("/bills", reg.Bills.ListUserBills)

    api.GET("/b2b/dashboard", reg.B2B.GetB2BDashboard)
    api.GET("/b2b/items", reg.B2B.ListWholesaleItems)
    api.GET("/b2b/items/:id", reg.B2B.GetWholesaleItem)
    api.POST("/b2b/rfqs", reg.B2B.CreateRFQ)
    api.GET("/b2b/rfqs", reg.B2B.ListMyRFQs)
    api.POST("/b2b/quotes", reg.B2B.SubmitQuote)

    api.GET("/messages", reg.User.ListMessages)
    api.GET("/notifications", reg.User.ListNotifications)
}
