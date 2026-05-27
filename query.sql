-- C2C Marketplace
-- name: CreateC2CSeller :one
INSERT INTO c2c_sellers (id, user_id, bio, avatar_url)
VALUES (@id, @user_id, @bio, @avatar_url)
RETURNING *;

-- name: GetC2CSellerByUserID :one
SELECT * FROM c2c_sellers WHERE user_id = @user_id;

-- name: CreateC2CListing :one
INSERT INTO c2c_listings (id, seller_id, title, description, price, currency, image_urls, is_negotiable, location, condition, status)
VALUES (@id, @seller_id, @title, @description, @price, @currency, @image_urls, @is_negotiable, @location, @condition, @status)
RETURNING *;

-- name: GetC2CListingByID :one
SELECT * FROM c2c_listings WHERE id = $1;

-- name: ListC2CListings :many
SELECT c.*, s.user_id as seller_user_id, s.avatar_url as seller_avatar_url
FROM c2c_listings c
JOIN c2c_sellers s ON c.seller_id = s.id
WHERE c.status = 'available' ORDER BY c.created_at DESC;

-- name: CountSellerListings :one
SELECT COUNT(*) FROM c2c_listings WHERE seller_id = $1;

-- Business
-- name: CreateBusiness :one
INSERT INTO businesses (id, owner_id, name, description, logo_url, banner_url, miniservice_type, address_id, phone_number, email)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetBusinessByID :one
SELECT * FROM businesses WHERE id = $1;

-- name: GetBusinessesByOwnerID :many
SELECT * FROM businesses WHERE owner_id = $1;

-- name: ListBusinessesByType :many
SELECT * FROM businesses WHERE miniservice_type = $1 AND verification_status = 'approved';

-- Services
-- name: CreateService :one
INSERT INTO services (id, business_id, service_type, name, description, base_price, currency, location)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetServiceByID :one
SELECT * FROM services WHERE id = $1;

-- name: ListServicesByType :many
SELECT s.*, b.name as business_name 
FROM services s
JOIN businesses b ON s.business_id = b.id
WHERE s.service_type = $1 AND s.is_active = TRUE;

-- Service Bookings
-- name: CreateServiceBooking :one
INSERT INTO service_bookings (id, user_id, service_type, service_item_id, provider_id, provider_type, start_time, end_time, total_amount, currency, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetServiceBookingsByUserID :many
SELECT * FROM service_bookings WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateServiceBookingStatus :one
UPDATE service_bookings SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- Grocery
-- name: ListGroceryItems :many
SELECT g.*, b.name as business_name
FROM grocery_items g
JOIN businesses b ON g.business_id = b.id
WHERE g.is_available = TRUE;

-- name: ListGroceryItemsByBusiness :many
SELECT * FROM grocery_items WHERE business_id = $1 AND is_available = TRUE;

-- name: SearchStoresByLocation :many
SELECT DISTINCT b.*, a.city, a.address_line1, a.longitude, a.latitude, ST_Distance(ST_SetSRID(ST_MakePoint(a.longitude, a.latitude), 4326), ST_SetSRID(ST_MakePoint($1, $2), 4326)) as distance
FROM businesses b
JOIN addresses a ON b.address_id = a.id
WHERE b.miniservice_type = ANY($3::text[])
AND b.verification_status = 'approved'
AND ST_DWithin(ST_SetSRID(ST_MakePoint(a.longitude, a.latitude), 4326), ST_SetSRID(ST_MakePoint($1, $2), 4326), $4)
ORDER BY distance;

-- Liquor
-- name: ListLiquorItems :many
SELECT l.*, b.name as business_name
FROM liquor_items l
JOIN businesses b ON l.business_id = b.id
WHERE l.is_available = TRUE;

-- name: ListLiquorItemsByBusiness :many
SELECT * FROM liquor_items WHERE business_id = $1 AND is_available = TRUE;

-- Health
-- name: CreateDoctor :one
INSERT INTO doctors (id, user_id, specialty, bio, license_number)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListDoctors :many
SELECT d.*, u.first_name, u.last_name, b.name as center_name
FROM doctors d
JOIN users u ON d.user_id = u.id
LEFT JOIN businesses b ON d.business_id = b.id;

-- name: CreateAppointment :one
INSERT INTO appointments (id, patient_id, doctor_id, appointment_time, notes)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListAppointmentsByPatient :many
SELECT a.*, d.specialty, u.first_name, u.last_name
FROM appointments a
JOIN doctors d ON a.doctor_id = d.id
JOIN users u ON d.user_id = u.id
WHERE a.patient_id = $1 ORDER BY a.appointment_time ASC;

-- Pharmacy Items (Already exists, ensures prescription check is respected for OTC)
-- name: ListPharmacyItems :many
SELECT * FROM pharmacy_items WHERE is_available = TRUE;

-- name: ListPharmacyItemsByBusiness :many
SELECT * FROM pharmacy_items WHERE business_id = $1 AND is_available = TRUE;

-- Food
-- name: ListAllFoodItems :many
SELECT f.*, b.name as restaurant_name, c.name as category_name
FROM food_items f
JOIN businesses b ON f.business_id = b.id
LEFT JOIN categories c ON f.category_id = c.id
WHERE f.is_available = TRUE;

-- name: ListFoodItemsByCategory :many
SELECT f.*, b.name as restaurant_name, c.name as category_name
FROM food_items f
JOIN businesses b ON f.business_id = b.id
JOIN categories c ON f.category_id = c.id
WHERE c.name = $1 AND f.is_available = TRUE;

-- name: GetFoodItemByID :one
SELECT f.*, b.name as restaurant_name, c.name as category_name
FROM food_items f
JOIN businesses b ON f.business_id = b.id
LEFT JOIN categories c ON f.category_id = c.id
WHERE f.id = $1;

-- Travel
-- name: ListBusRoutes :many
SELECT r.*, b.name as business_name
FROM bus_routes r
JOIN businesses b ON r.business_id = b.id
WHERE (sqlc.narg('origin')::text IS NULL OR r.origin ILIKE '%' || sqlc.narg('origin') || '%')
AND (sqlc.narg('destination')::text IS NULL OR r.destination ILIKE '%' || sqlc.narg('destination') || '%')
AND r.departure_time >= CURRENT_TIMESTAMP
ORDER BY r.departure_time ASC;

-- name: CreateBusTicket :one
INSERT INTO tickets (id, user_id, showtime_id, seat_number, ticket_number, qr_code_data, total_amount, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'booked')
RETURNING *;

-- Cinema
-- name: ListNowPlayingMovies :many
SELECT * FROM movies WHERE is_now_playing = TRUE ORDER BY release_date DESC;

-- name: ListComingSoonMovies :many
SELECT * FROM movies WHERE is_now_playing = FALSE ORDER BY release_date ASC;

-- name: GetMovieDetails :one
SELECT * FROM movies WHERE id = $1;

-- name: ListMovieShowtimes :many
SELECT ms.*, b.name as cinema_name
FROM movie_showtimes ms
JOIN businesses b ON ms.business_id = b.id
WHERE ms.movie_id = $1 AND ms.show_time >= CURRENT_TIMESTAMP
ORDER BY ms.show_time ASC;

-- name: ListRefreshmentsByCinema :many
SELECT * FROM refreshments WHERE business_id = $1;

-- name: CreateTicket :one
INSERT INTO tickets (id, user_id, showtime_id, seat_number, ticket_number, qr_code_data, refreshment_ids, total_amount, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'booked')
RETURNING *;

-- Flights
-- name: ListFlights :many
SELECT * FROM flights WHERE departure_time > CURRENT_TIMESTAMP ORDER BY departure_time ASC;

-- name: CreateFlightTicket :one
INSERT INTO tickets (id, user_id, showtime_id, seat_number, ticket_number, qr_code_data, total_amount, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'booked')
RETURNING *;

-- Jobs
-- name: ListActiveJobs :many
SELECT j.*, b.name as business_name
FROM jobs j
JOIN businesses b ON j.business_id = b.id
WHERE j.is_active = TRUE ORDER BY j.created_at DESC;

-- Travel/Tours
-- name: ListTours :many
SELECT t.*, b.name as business_name
FROM tours t
JOIN businesses b ON t.business_id = b.id
ORDER BY t.created_at DESC;

-- Bills
-- name: ListUserBills :many
SELECT * FROM bills WHERE user_id = $1 ORDER BY due_date ASC;

-- Wallet & Transactions
-- name: GetWalletBalance :one
SELECT * FROM user_wallet WHERE user_id = $1;

-- name: UpdateWalletBalance :one
INSERT INTO user_wallet (user_id, balance)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET balance = $2, last_updated = CURRENT_TIMESTAMP
RETURNING *;

-- messaging & notifications
-- name: ListUserMessages :many
SELECT * FROM messages WHERE sender_id = $1 OR receiver_id = $1 ORDER BY created_at DESC;

-- name: ListUserNotifications :many
SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC;

-- Products (Ecommerce)
-- name: GetProducts :many
SELECT p.*, c.name as category_name, b.name as brand_name, m.name as model_name
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN product_models m ON p.model_id = m.id
ORDER BY p.created_at DESC LIMIT $1 OFFSET $2;

-- name: GetProductsByBusiness :many
SELECT p.*, c.name as category_name, b.name as brand_name, m.name as model_name
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN product_models m ON p.model_id = m.id
WHERE p.business_id = $1
ORDER BY p.created_at DESC LIMIT $2 OFFSET $3;

-- name: GetProductByIDWithDetails :one
SELECT p.*, c.name as category_name, b.name as brand_name, m.name as model_name
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN product_models m ON p.model_id = m.id
WHERE p.id = $1;

-- name: GetFeaturedProducts :many
SELECT * FROM products WHERE rating >= 4.0 ORDER BY review_count DESC LIMIT $1 OFFSET $2;

-- name: GetCategories :many
SELECT * FROM categories ORDER BY name ASC;

-- name: CreateCategory :one
INSERT INTO categories (id, name, description, image_url, parent_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCategory :one
UPDATE categories SET name = $2, description = $3, image_url = $4, parent_id = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- name: GetBrands :many
SELECT b.*, c.name as category_name 
FROM brands b
LEFT JOIN categories c ON b.category_id = c.id
ORDER BY b.name ASC;

-- name: ListBrandsByCategory :many
SELECT * FROM brands WHERE category_id = $1 ORDER BY name ASC;

-- name: CreateBrand :one
INSERT INTO brands (id, name, logo_url, category_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateBrand :one
UPDATE brands SET name = $2, logo_url = $3, category_id = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 RETURNING *;

-- name: DeleteBrand :exec
DELETE FROM brands WHERE id = $1;

-- name: ListProductModels :many
SELECT m.*, b.name as brand_name
FROM product_models m
JOIN brands b ON m.brand_id = b.id
ORDER BY m.name ASC;

-- name: ListModelsByBrand :many
SELECT * FROM product_models WHERE brand_id = $1 ORDER BY name ASC;

-- name: CreateProductModel :one
INSERT INTO product_models (id, brand_id, name, image_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateProductModel :one
UPDATE product_models SET brand_id = $2, name = $3, image_url = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 RETURNING *;

-- name: DeleteProductModel :exec
DELETE FROM product_models WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (id, business_id, name, description, price, currency, stock_quantity, category_id, brand_id, model_id, image_urls, rating, review_count, is_flash_sale, discount_percentage)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products SET 
    name = $2, description = $3, price = $4, stock_quantity = $5, 
    category_id = $6, brand_id = $7, model_id = $8, image_urls = $9, 
    is_flash_sale = $10, discount_percentage = $11, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 RETURNING *;

-- name: CreateGroceryItem :one
INSERT INTO grocery_items (id, business_id, name, description, price, currency, image_url, unit, stock_quantity, category, is_available)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: CreatePharmacyItem :one
INSERT INTO pharmacy_items (id, business_id, name, description, price, currency, image_url, requires_prescription, stock_quantity, category, is_available)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: CreateFoodItem :one
INSERT INTO food_items (id, business_id, category_id, name, description, price, currency, image_urls, is_available)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetNearbyMotorbikeDrivers :many
SELECT d.*, ST_X(last_location) as longitude, ST_Y(last_location) as latitude, ST_Distance(last_location, ST_SetSRID(ST_MakePoint($1, $2), 4326)) as distance
FROM drivers d
JOIN vehicle_types vt ON d.vehicle_type_id = vt.id
WHERE d.status = 'online' 
AND vt.name = 'motorbike'
AND ST_DWithin(last_location, ST_SetSRID(ST_MakePoint($1, $2), 4326), $3)
ORDER BY distance LIMIT $4;

-- name: AssignDriverToOrder :one
UPDATE orders SET driver_id = $2, delivery_fee = $3, status = 'assigned', updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- name: CreateBusRoute :one
INSERT INTO bus_routes (id, business_id, origin, destination, departure_time, arrival_time, price, currency, available_seats, bus_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: CreateMovie :one
INSERT INTO movies (id, title, description, genre, rating, duration_minutes, poster_url, is_now_playing, release_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: CreateDriver :one
INSERT INTO drivers (id, user_id, name, status, vehicle_type_id, rating)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- Role management
-- name: AssignRoleToUser :one
INSERT INTO user_roles (user_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING *;

-- Users
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, first_name, last_name, date_of_birth, phone_number, profile_picture_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- Hubs
-- name: CreateHub :one
INSERT INTO hubs (id, name, description) VALUES ($1, $2, $3) RETURNING *;

-- Business Status
-- name: UpdateBusinessStatus :one
UPDATE businesses SET verification_status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- Delete Product
-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- B2B / RFQs
-- name: CreateRFQ :one
INSERT INTO rfqs (id, buyer_id, business_id, title, description, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: AddRFQItem :one
INSERT INTO rfq_items (id, rfq_id, product_id, item_name, quantity, unit, target_price)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListRFQsForBuyer :many
SELECT r.*, b.name as business_name
FROM rfqs r
JOIN businesses b ON r.business_id = b.id
WHERE r.buyer_id = $1 ORDER BY r.created_at DESC;

-- name: ListRFQsForBusiness :many
SELECT r.*, u.first_name, u.last_name
FROM rfqs r
JOIN users u ON r.buyer_id = u.id
WHERE r.business_id = $1 ORDER BY r.created_at DESC;

-- name: GetRFQWithItems :many
SELECT r.*, ri.id as item_id, ri.product_id, ri.item_name, ri.quantity, ri.unit, ri.target_price
FROM rfqs r
LEFT JOIN rfq_items ri ON r.id = ri.rfq_id
WHERE r.id = $1;

-- name: UpdateRFQStatus :one
UPDATE rfqs SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- B2B Quotes
-- name: CreateB2BQuote :one
INSERT INTO b2b_quotes (id, rfq_id, business_id, total_amount, currency, valid_until, status, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetQuotesForRFQ :many
SELECT q.*, b.name as business_name
FROM b2b_quotes q
JOIN businesses b ON q.business_id = b.id
WHERE q.rfq_id = $1 ORDER BY q.created_at DESC;

-- name: UpdateB2BQuoteStatus :one
UPDATE b2b_quotes SET status = $2 WHERE id = $1 RETURNING *;

-- name: GetB2BDashboardStats :one
SELECT 
    (SELECT COUNT(*) FROM rfqs WHERE buyer_id = $1 AND status = 'pending') as active_rfqs,
    (SELECT COUNT(*) FROM orders WHERE user_id = $1 AND status = 'in_transit') as in_transit_orders;

-- Wholesale Items
-- name: CreateWholesaleItem :one
INSERT INTO wholesale_items (id, business_id, name, description, image_urls, unit_price, bulk_price, bulk_quantity, category)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListWholesaleItems :many
SELECT w.*, b.name as business_name, b.rating as business_rating, b.logo_url as business_logo
FROM wholesale_items w
JOIN businesses b ON w.business_id = b.id
WHERE w.is_available = TRUE ORDER BY w.created_at DESC;

-- name: GetWholesaleItemByID :one
SELECT w.*, b.name as business_name, b.rating as business_rating, b.logo_url as business_logo, b.description as business_description
FROM wholesale_items w
JOIN businesses b ON w.business_id = b.id
WHERE w.id = $1;

-- name: ListWholesaleItemsByBusiness :many
SELECT * FROM wholesale_items WHERE business_id = $1 AND is_available = TRUE;

-- Business Locations
-- name: AddBusinessLocation :exec
INSERT INTO business_locations (business_id, address_id) VALUES ($1, $2);

-- name: ListBusinessLocations :many
SELECT a.* FROM addresses a
JOIN business_locations bl ON a.id = bl.address_id
WHERE bl.business_id = $1;

-- Cart Items
-- name: GetCartItemsByUserID :many
SELECT c.*, p.name, p.price, p.currency, b.name as business_name
FROM cart_items c
JOIN products p ON c.item_id = p.id
JOIN businesses b ON c.business_id = b.id
WHERE c.user_id = $1;

-- name: AddItemToCart :one
INSERT INTO cart_items (id, user_id, business_id, item_id, item_type, quantity)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, item_id, item_type) DO UPDATE SET quantity = cart_items.quantity + $6, updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: UpdateCartItemQuantity :exec
UPDATE cart_items SET quantity = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: RemoveCartItem :exec
DELETE FROM cart_items WHERE id = $1;

-- Orders
-- name: CreateOrder :one
INSERT INTO orders (id, user_id, total_amount, currency, status, shipping_address_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetOrdersByUserID :many
SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateOrderItem :one
INSERT INTO order_items (id, order_id, business_id, item_id, item_type, quantity, unit_price)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ClearCart :exec
DELETE FROM cart_items WHERE user_id = $1;

-- name: LockAndDecrementStock :exec
UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2 AND stock_quantity >= $1;

-- Addresses
-- name: ListUserAddresses :many
SELECT * FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC;

-- name: CreateAddress :one
INSERT INTO addresses (id, user_id, label, address_line1, city, country, is_default)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- Property
-- name: ListProperties :many
SELECT p.*, b.name as business_name, a.city, a.address_line1
FROM properties p
JOIN businesses b ON p.business_id = b.id
LEFT JOIN addresses a ON p.address_id = a.id
WHERE (sqlc.narg('city')::text IS NULL OR a.city ILIKE '%' || sqlc.narg('city') || '%')
AND (sqlc.narg('max_price')::numeric IS NULL OR p.price_per_night <= sqlc.narg('max_price'))
AND (sqlc.narg('min_rooms')::int IS NULL OR p.number_of_bedrooms >= sqlc.narg('min_rooms'));

-- name: GetPropertyByID :one
SELECT p.*, b.name as business_name, a.city, a.address_line1
FROM properties p
JOIN businesses b ON p.business_id = b.id
LEFT JOIN addresses a ON p.address_id = a.id
WHERE p.id = $1;

-- name: CreateProperty :one
INSERT INTO properties (id, business_id, title, description, address_id, price_per_night, currency, number_of_guests, number_of_bedrooms, type, image_urls)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: CreatePropertyBooking :one
INSERT INTO property_bookings (id, user_id, property_id, check_in_date, check_out_date, total_amount, currency, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')
RETURNING *;

-- name: SearchPropertiesByLocation :many
SELECT p.* FROM properties p
JOIN addresses a ON p.address_id = a.id
WHERE ST_DWithin(ST_SetSRID(ST_MakePoint(a.longitude, a.latitude), 4326), ST_SetSRID(ST_MakePoint($1, $2), 4326), $3);

-- Taxi
-- name: UpdateDriverLocation :one
UPDATE drivers SET last_location = ST_SetSRID(ST_MakePoint($2, $3), 4326), updated_at = CURRENT_TIMESTAMP WHERE user_id = $1 RETURNING *;

-- name: UpdateDriverStatus :one
UPDATE drivers SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1 RETURNING *;

-- name: GetNearbyDrivers :many
SELECT *, ST_X(last_location) as longitude, ST_Y(last_location) as latitude, ST_Distance(last_location, ST_SetSRID(ST_MakePoint($1, $2), 4326)) as distance
FROM drivers
WHERE status = 'online' AND ST_DWithin(last_location, ST_SetSRID(ST_MakePoint($1, $2), 4326), $3)
ORDER BY distance LIMIT $4;

-- name: CreateTaxiTrip :one
INSERT INTO taxi_trips (id, user_id, pickup_location, dropoff_location, total_amount, currency)
VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), ST_SetSRID(ST_MakePoint($5, $6), 4326), $7, $8)
RETURNING *;

-- Reviews
-- name: AddReview :one
INSERT INTO reviews (id, target_id, target_type, user_id, rating, comment)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetReviewsByTarget :many
SELECT * FROM reviews WHERE target_id = $1 AND target_type = $2 ORDER BY created_at DESC;

