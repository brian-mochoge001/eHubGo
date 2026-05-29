package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"time"

	"ehubgo/db"
	"github.com/gin-gonic/gin"
	"github.com/gojuno/go.osrm"
	"github.com/google/uuid"
	"github.com/paulmach/go.geo"
)

type TaxiHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewTaxiHandler(queries *db.Queries, dbConn *sql.DB) *TaxiHandler {
	return &TaxiHandler{
		Queries: queries,
		DB:      dbConn,
	}
}

// UpdateLocation allows drivers to ping their GPS coordinates
func (h *TaxiHandler) UpdateLocation(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		Longitude float64 `json:"longitude" binding:"required"`
		Latitude  float64 `json:"latitude" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Snap coordinates to the nearest road node
	snappedLat, snappedLng := SnapToRoad(c.Request.Context(), req.Latitude, req.Longitude)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		driver, err := qtx.UpdateDriverLocation(c.Request.Context(), db.UpdateDriverLocationParams{
			UserID:        userID,
			StMakepoint:   snappedLng,
			StMakepoint_2: snappedLat,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, driver)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// UpdateStatus allows drivers to go online/offline
func (h *TaxiHandler) UpdateStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		Status string `json:"status" binding:"required"` // online, offline, busy
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		driver, err := qtx.UpdateDriverStatus(c.Request.Context(), db.UpdateDriverStatusParams{
			UserID: userID,
			Status: req.Status,
		})
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, driver)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetNearbyDrivers allows users to see available taxis on the map
func (h *TaxiHandler) GetNearbyDrivers(c *gin.Context) {
	var params struct {
		Longitude float64 `form:"longitude" binding:"required"`
		Latitude  float64 `form:"latitude" binding:"required"`
		Radius    float64 `form:"radius,default=5000"`
		Limit     int     `form:"limit,default=5"`
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if params.Radius <= 0 || params.Radius > 5000 {
		params.Radius = 5000
	}
	if params.Limit <= 0 {
		params.Limit = 5
	}
	if params.Limit > 5 {
		params.Limit = 5
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		drivers, err := qtx.GetNearbyDrivers(c.Request.Context(), db.GetNearbyDriversParams{
			StMakepoint:   params.Longitude,
			StMakepoint_2: params.Latitude,
			StDwithin:     params.Radius,
			Limit:         int32(params.Limit),
		})
		if err != nil {
			return err
		}

		if len(drivers) == 0 {
			c.JSON(http.StatusOK, []gin.H{})
			return nil
		}

		// Initialize OSRM client
		client := osrm.NewFromURL("http://router.project-osrm.org")

		var results []gin.H

		for _, d := range drivers {
			dLat := ParseCoordinate(d.Latitude)
			dLng := ParseCoordinate(d.Longitude)

			route, err := client.Route(c.Request.Context(), osrm.RouteRequest{
				Profile: "driving",
				Coordinates: osrm.NewGeometryFromPointSet(geo.PointSet{
					geo.Point{dLng, dLat},
					geo.Point{params.Longitude, params.Latitude},
				}),
			})

			etaMinutes := 15
			if err == nil && len(route.Routes) > 0 {
				etaMinutes = int(route.Routes[0].Duration / 60)
			}

			if etaMinutes <= 15 {
				results = append(results, gin.H{
					"driver":      d,
					"eta_minutes": etaMinutes,
				})
			}
		}

		sort.Slice(results, func(i, j int) bool {
			return results[i]["eta_minutes"].(int) < results[j]["eta_minutes"].(int)
		})

		c.JSON(http.StatusOK, results)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// DriverWS upgrades the connection and registers the driver for realtime messages
func (h *TaxiHandler) DriverWS(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	// Ensure the client can accept SSE
	w := c.Writer
	r := c.Request
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	var ch chan []byte
	if DriverHub != nil {
		ch = DriverHub.Register(userID)
	} else {
		ch = make(chan []byte)
		close(ch)
	}

	done := make(chan struct{})
	go func() {
		<-r.Context().Done()
		if DriverHub != nil {
			DriverHub.Unregister(userID)
		}
		close(done)
	}()

	StreamEvents(w, flusher, ch, done)
}

// GetRideHistory returns all completed taxi trips for the authenticated user
func (h *TaxiHandler) GetRideHistory(c *gin.Context) {
	// ...

	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(c.Request.Context(), `
			SELECT t.id, t.status, t.total_amount, t.currency, t.created_at,
			       d.name as driver_name
			FROM taxi_trips t
			LEFT JOIN drivers d ON t.driver_id = d.id
			WHERE t.user_id = $1 AND t.status = 'completed'
			ORDER BY t.created_at DESC
		`, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var history []gin.H
		for rows.Next() {
			var id, status, totalAmount, currency string
			var driverName sql.NullString
			var createdAt time.Time
			if err := rows.Scan(&id, &status, &totalAmount, &currency, &createdAt, &driverName); err != nil {
				return err
			}
			history = append(history, gin.H{
				"id":           id,
				"status":       status,
				"total_amount": totalAmount,
				"currency":     currency,
				"driver_name":  driverName.String,
				"created_at":   createdAt,
			})
		}
		c.JSON(http.StatusOK, history)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *TaxiHandler) GetDriverLocation(c *gin.Context) {
	driverID := c.Query("driver_id")
	if driverID == "" {
		driverID = c.MustGet("user_id").(string)
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(c.Request.Context(), `
			SELECT id, user_id, name, status, vehicle_type_id, rating, last_location, updated_at, created_at
			FROM drivers
			WHERE id = $1
		`, driverID)
		var driver db.Driver
		if err := row.Scan(&driver.ID, &driver.UserID, &driver.Name, &driver.Status, &driver.VehicleTypeID, &driver.Rating, &driver.LastLocation, &driver.UpdatedAt, &driver.CreatedAt); err != nil {
			return err
		}
		c.JSON(http.StatusOK, driver)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *TaxiHandler) GetDriverTasks(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(c.Request.Context(), `
			SELECT t.id, t.user_id, t.driver_id, t.status, t.total_amount, t.currency,
				ST_X(t.pickup_location) AS pickup_lng, ST_Y(t.pickup_location) AS pickup_lat,
				ST_X(t.dropoff_location) AS dropoff_lng, ST_Y(t.dropoff_location) AS dropoff_lat,
				u.first_name, u.last_name, t.accepted_at, t.started_at, t.completed_at, t.created_at, t.updated_at
			FROM taxi_trips t
			JOIN users u ON u.id = t.user_id
			WHERE t.driver_id = $1
			ORDER BY t.updated_at DESC
		`, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var tasks []gin.H
		for rows.Next() {
			var id, userID, status, totalAmount, currency, firstName, lastName string
			var driverID sql.NullString
			var pickupLng, pickupLat, dropoffLng, dropoffLat float64
			var acceptedAt, startedAt, completedAt, createdAt, updatedAt sql.NullTime
			if err := rows.Scan(&id, &userID, &driverID, &status, &totalAmount, &currency, &pickupLng, &pickupLat, &dropoffLng, &dropoffLat, &firstName, &lastName, &acceptedAt, &startedAt, &completedAt, &createdAt, &updatedAt); err != nil {
				return err
			}
			tasks = append(tasks, gin.H{
				"id":               id,
				"status":           status,
				"total_amount":     totalAmount,
				"currency":         currency,
				"customer_name":    firstName + " " + lastName,
				"pickup_location":  gin.H{"latitude": pickupLat, "longitude": pickupLng},
				"dropoff_location": gin.H{"latitude": dropoffLat, "longitude": dropoffLng},
				"polyline":         GetPointToPointRoute(c.Request.Context(), pickupLat, pickupLng, dropoffLat, dropoffLng),
				"accepted_at":      acceptedAt,
				"started_at":       startedAt,
				"completed_at":     completedAt,
				"created_at":       createdAt,
				"updated_at":       updatedAt,
			})
		}
		c.JSON(http.StatusOK, tasks)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *TaxiHandler) GetRideRequests(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(c.Request.Context(), `
			SELECT t.id, t.user_id, t.status, t.total_amount, t.currency,
				ST_X(t.pickup_location) AS pickup_lng, ST_Y(t.pickup_location) AS pickup_lat,
				ST_X(t.dropoff_location) AS dropoff_lng, ST_Y(t.dropoff_location) AS dropoff_lat,
				u.first_name, u.last_name, ST_Distance(t.pickup_location, d.last_location) AS distance
			FROM taxi_trips t
			JOIN drivers d ON d.user_id = $1
			JOIN users u ON u.id = t.user_id
			WHERE t.status = 'requested'
			AND ST_DWithin(t.pickup_location, d.last_location, 5000)
			ORDER BY distance ASC
			LIMIT 10
		`, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var requests []gin.H
		for rows.Next() {
			var id, userID, status, totalAmount, currency, firstName, lastName string
			var pickupLng, pickupLat, dropoffLng, dropoffLat, distance float64
			if err := rows.Scan(&id, &userID, &status, &totalAmount, &currency, &pickupLng, &pickupLat, &dropoffLng, &dropoffLat, &firstName, &lastName, &distance); err != nil {
				return err
			}
			requests = append(requests, gin.H{
				"id":               id,
				"status":           status,
				"total_amount":     totalAmount,
				"currency":         currency,
				"customer_name":    firstName + " " + lastName,
				"pickup_location":  gin.H{"latitude": pickupLat, "longitude": pickupLng},
				"dropoff_location": gin.H{"latitude": dropoffLat, "longitude": dropoffLng},
				"polyline":         GetPointToPointRoute(c.Request.Context(), pickupLat, pickupLng, dropoffLat, dropoffLng),
				"distance_m":       distance,
			})
		}
		c.JSON(http.StatusOK, requests)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *TaxiHandler) AcceptRideRequest(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		TripID string `json:"trip_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		var busyCount int
		if err := tx.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM taxi_trips WHERE driver_id = $1 AND status IN ('accepted', 'in_transit')`, userID).Scan(&busyCount); err != nil {
			return err
		}
		if busyCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You already have an active ride. Complete it before accepting another request."})
			return nil
		}

		row := tx.QueryRowContext(c.Request.Context(), `
			UPDATE taxi_trips
			SET driver_id = $1, status = 'accepted', accepted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2 AND status = 'requested'
			RETURNING id, user_id, driver_id, status, total_amount, currency
		`, userID, req.TripID)
		var id, tripUserID, tripStatus, totalAmount, currency string
		var driverID sql.NullString
		if err := row.Scan(&id, &tripUserID, &driverID, &tripStatus, &totalAmount, &currency); err != nil {
			return err
		}
		c.JSON(http.StatusOK, gin.H{"id": id, "status": tripStatus, "total_amount": totalAmount, "currency": currency})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *TaxiHandler) DeclineRideRequest(c *gin.Context) {
	var req struct {
		TripID string `json:"trip_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(c.Request.Context(), `
			UPDATE taxi_trips
			SET status = 'declined', updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND status = 'requested'
			RETURNING id, status
		`, req.TripID)
		var id, status string
		if err := row.Scan(&id, &status); err != nil {
			return err
		}
		c.JSON(http.StatusOK, gin.H{"id": id, "status": status})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *TaxiHandler) GetDeliveryRequests(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(c.Request.Context(), `
			SELECT o.id, o.user_id, o.status, o.total_amount, o.currency,
				a.longitude AS pickup_lng, a.latitude AS pickup_lat,
				u.first_name, u.last_name,
				ST_Distance(ST_SetSRID(ST_MakePoint(a.longitude, a.latitude), 4326), d.last_location) AS distance
			FROM orders o
			JOIN order_items oi ON oi.order_id = o.id
			JOIN businesses b ON b.id = oi.business_id
			JOIN addresses a ON a.id = b.address_id
			JOIN drivers d ON d.user_id = $1
			JOIN users u ON u.id = o.user_id
			WHERE o.status = 'requested'
			AND b.miniservice_type IN ('food', 'grocery', 'liquor')
			AND ST_DWithin(ST_SetSRID(ST_MakePoint(a.longitude, a.latitude), 4326), d.last_location, 5000)
			ORDER BY distance ASC
			LIMIT 20
		`, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var requests []gin.H
		for rows.Next() {
			var id, uid, status, totalAmount, currency, firstName, lastName string
			var pickupLng, pickupLat, distance float64
			if err := rows.Scan(&id, &uid, &status, &totalAmount, &currency, &pickupLng, &pickupLat, &firstName, &lastName, &distance); err != nil {
				return err
			}
			requests = append(requests, gin.H{
				"id":              id,
				"status":          status,
				"total_amount":    totalAmount,
				"currency":        currency,
				"customer_name":   firstName + " " + lastName,
				"pickup_location": gin.H{"latitude": pickupLat, "longitude": pickupLng},
				"distance_m":      distance,
				"service_type":    "delivery",
			})
		}
		c.JSON(http.StatusOK, requests)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *TaxiHandler) AcceptDeliveryRequest(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		var busyCount int
		if err := tx.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM orders WHERE driver_id = $1 AND status IN ('assigned', 'in_transit')`, userID).Scan(&busyCount); err != nil {
			return err
		}
		if busyCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You already have an active delivery. Complete it before accepting another request."})
			return nil
		}

		row := tx.QueryRowContext(c.Request.Context(), `
			UPDATE orders
			SET driver_id = $1, status = 'assigned', updated_at = CURRENT_TIMESTAMP
			WHERE id = $2 AND status = 'requested'
			RETURNING id, status, total_amount, currency
		`, userID, req.OrderID)
		var id, status, totalAmount, currency string
		if err := row.Scan(&id, &status, &totalAmount, &currency); err != nil {
			return err
		}
		c.JSON(http.StatusOK, gin.H{"id": id, "status": status, "total_amount": totalAmount, "currency": currency})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *TaxiHandler) DeclineDeliveryRequest(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(string)

	var notifyDriverIDs []string

	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		// mark updated timestamp to show driver responded
		row := tx.QueryRowContext(c.Request.Context(), `
			UPDATE orders
			SET updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND status = 'requested'
			RETURNING id, status
		`, req.OrderID)
		var id, status string
		if err := row.Scan(&id, &status); err != nil {
			return err
		}

		// record decline in in-memory tracker
		Declines.Add(req.OrderID, userID)

		// find pickup coordinates for the order to search nearby drivers
		var pickupLng, pickupLat float64
		if err := tx.QueryRowContext(c.Request.Context(), `SELECT a.longitude, a.latitude FROM order_items oi JOIN businesses b ON b.id = oi.business_id JOIN addresses a ON a.id = b.address_id WHERE oi.order_id = $1 LIMIT 1`, req.OrderID).Scan(&pickupLng, &pickupLat); err != nil {
			// if we cannot determine pickup location, return success but no broadcast
			c.JSON(http.StatusOK, gin.H{"id": id, "status": status})
			return nil
		}

		// get nearby motorbike drivers (ordered by distance)
		drivers, dErr := h.Queries.GetNearbyMotorbikeDrivers(c.Request.Context(), db.GetNearbyMotorbikeDriversParams{
			StMakepoint:   pickupLng,
			StMakepoint_2: pickupLat,
			StDwithin:     5000,
			Limit:         20,
		})
		if dErr != nil {
			c.JSON(http.StatusOK, gin.H{"id": id, "status": status})
			return nil
		}

		declined := make(map[string]bool)
		for _, d := range Declines.Get(req.OrderID) {
			declined[d] = true
		}

		// collect up to 10 next available drivers excluding declined and busy drivers
		for _, drv := range drivers {
			if len(notifyDriverIDs) >= 10 {
				break
			}
			if declined[drv.UserID] {
				continue
			}
			var busyCount int
			if err := tx.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM orders WHERE driver_id = $1 AND status IN ('assigned', 'in_transit')`, drv.UserID).Scan(&busyCount); err != nil {
				continue
			}
			if busyCount > 0 {
				continue
			}
			notifyDriverIDs = append(notifyDriverIDs, drv.UserID)
		}

		c.JSON(http.StatusOK, gin.H{"id": id, "status": status})
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// broadcast to next drivers (best-effort)
	if len(notifyDriverIDs) > 0 && DriverHub != nil {
		DriverHub.BroadcastToDrivers(notifyDriverIDs, gin.H{"type": "delivery_request", "order_id": req.OrderID})
	}
}

// RequestRide creates a new taxi trip request
func (h *TaxiHandler) RequestRide(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var req struct {
		PickupLng  float64 `json:"pickup_lng" binding:"required"`
		PickupLat  float64 `json:"pickup_lat" binding:"required"`
		DropoffLng float64 `json:"dropoff_lng" binding:"required"`
		DropoffLat float64 `json:"dropoff_lat" binding:"required"`
		Amount     float64 `json:"amount" binding:"required"`
		Currency   string  `json:"currency" default:"Ksh"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var createdTrip gin.H
	err := WithRLS(c, h.DB, func(tx *sql.Tx) error {
		qtx := h.Queries.WithTx(tx)
		trip, err := qtx.CreateTaxiTrip(c.Request.Context(), db.CreateTaxiTripParams{
			ID:            uuid.New().String(),
			UserID:        userID,
			StMakepoint:   req.PickupLng,
			StMakepoint_2: req.PickupLat,
			StMakepoint_3: req.DropoffLng,
			StMakepoint_4: req.DropoffLat,
			TotalAmount:   fmt.Sprintf("%.2f", req.Amount),
			Currency:      req.Currency,
		})
		if err != nil {
			return err
		}
		createdTrip = gin.H{
			"trip":     trip,
			"polyline": GetPointToPointRoute(c.Request.Context(), req.PickupLat, req.PickupLng, req.DropoffLat, req.DropoffLng),
		}
		c.JSON(http.StatusCreated, createdTrip)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Notify nearby drivers via SSE hub (best-effort)
	if DriverHub != nil {
		// find nearby drivers using DB queries
		drivers, dErr := h.Queries.GetNearbyDrivers(c.Request.Context(), db.GetNearbyDriversParams{
			StMakepoint:   req.PickupLng,
			StMakepoint_2: req.PickupLat,
			StDwithin:     5000,
			Limit:         10,
		})
		if dErr == nil && len(drivers) > 0 {
			var ids []string
			for _, d := range drivers {
				ids = append(ids, d.UserID)
			}
			DriverHub.BroadcastToDrivers(ids, gin.H{"type": "ride_request", "data": createdTrip})
		}
	}
}
