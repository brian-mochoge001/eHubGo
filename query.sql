-- Users
-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name, phone_number, profile_picture_url)
VALUES ($1::TEXT, $2::TEXT, $3::TEXT, $4::TEXT, $5::TEXT, $6::TEXT)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- User Roles
-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role) VALUES ($1, $2);

-- name: GetUserRoles :many
SELECT role FROM user_roles WHERE user_id = $1;

-- Businesses (The "Stalls")
-- name: CreateBusiness :one
INSERT INTO businesses (id, owner_id, name, description, logo_url, banner_url, miniservice_type, address_id, phone_number, email)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetBusinessByID :one
SELECT * FROM businesses WHERE id = $1;

-- name: GetBusinessesByOwnerID :many
SELECT * FROM businesses WHERE owner_id = $1;

-- name: UpdateBusinessStatus :one
UPDATE businesses 
SET verification_status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListBusinessesByType :many
SELECT * FROM businesses WHERE miniservice_type = $1 AND verification_status = 'approved';

-- Products
-- name: CreateProduct :one
INSERT INTO products (id, business_id, name, description, price, currency, stock_quantity, category_id, brand_id, is_flash_sale, discount_percentage)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetProductsByBusiness :many
SELECT * FROM products WHERE business_id = $1;

-- name: GetProductByIDWithDetails :one
SELECT p.*, b.name as brand_name, c.name as category_name
FROM products p
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.id = $1;

-- name: GetFeaturedProducts :many
SELECT p.*, b.name as brand_name, c.name as category_name
FROM products p
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.rating >= 4.0
LIMIT $1;

-- name: GetFlashSaleProducts :many
SELECT * FROM products WHERE is_flash_sale = TRUE AND (flash_sale_end_time IS NULL OR flash_sale_end_time > CURRENT_TIMESTAMP);

-- name: GetProducts :many
SELECT p.*, b.name as brand_name, c.name as category_name
FROM products p
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN categories c ON p.category_id = c.id;

-- name: GetCategories :many
SELECT id, name, description, image_url, parent_id, created_at, updated_at FROM categories
ORDER BY name;

-- name: LockAndDecrementStock :exec
UPDATE products
SET stock_quantity = stock_quantity - @stock_quantity, updated_at = CURRENT_TIMESTAMP
WHERE id = @id AND stock_quantity >= @stock_quantity;

-- Brands
-- name: CreateBrand :one
INSERT INTO brands (id, name, logo_url)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetBrandByID :one
SELECT * FROM brands WHERE id = $1;

-- name: ListBrands :many
SELECT * FROM brands ORDER BY name;

-- name: UpdateBrand :one
UPDATE brands
SET name = $1, logo_url = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $3
RETURNING *;

-- name: DeleteBrand :exec
DELETE FROM brands WHERE id = $1;

-- Product Variants
-- name: CreateProductVariant :one
INSERT INTO product_variants (id, product_id, sku, price, stock_quantity, attributes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetProductVariantsByProductID :many
SELECT * FROM product_variants WHERE product_id = $1;

-- name: GetProductVariantByID :one
SELECT * FROM product_variants WHERE id = $1;

-- name: UpdateProductVariant :one
UPDATE product_variants
SET sku = $1, price = $2, stock_quantity = $3, attributes = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $5
RETURNING *;

-- name: DeleteProductVariant :exec
DELETE FROM product_variants WHERE id = $1;

-- Product Discounts
-- name: CreateProductDiscount :one
INSERT INTO product_discounts (id, product_id, discount_type, discount_value, start_at, end_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetProductDiscountsByProductID :many
SELECT * FROM product_discounts WHERE product_id = $1;

-- name: GetProductDiscountByID :one
SELECT * FROM product_discounts WHERE id = $1;

-- name: UpdateProductDiscount :one
UPDATE product_discounts
SET discount_type = $1, discount_value = $2, start_at = $3, end_at = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $5
RETURNING *;

-- name: DeleteProductDiscount :exec
DELETE FROM product_discounts WHERE id = $1;

-- Services
-- name: CreateService :one
INSERT INTO services (id, business_id, service_type, name, description, base_price, currency, location)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListServicesByBusiness :many
SELECT * FROM services WHERE business_id = $1;

-- name: ListServicesByType :many
SELECT * FROM services WHERE service_type = $1 AND is_active = TRUE;

-- name: GetServiceByID :one
SELECT * FROM services WHERE id = $1;

-- name: DeleteService :exec
DELETE FROM services WHERE id = $1;

-- Service Bookings
-- name: CreateServiceBooking :one
INSERT INTO service_bookings (id, user_id, service_type, service_item_id, provider_id, provider_type, start_time, end_time, total_amount, currency, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetServiceBookingsByUserID :many
SELECT * FROM service_bookings WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetServiceBookingsByProviderID :many
SELECT * FROM service_bookings WHERE provider_id = $1 ORDER BY created_at DESC;

-- name: UpdateServiceBookingStatus :one
UPDATE service_bookings
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- Properties
-- name: CreateProperty :one
INSERT INTO properties (id, business_id, title, description, address_id, price_per_night, currency, number_of_guests, number_of_bedrooms, type, image_urls)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetPropertyByID :one
SELECT * FROM properties WHERE id = $1;

-- name: ListPropertiesByBusiness :many
SELECT * FROM properties WHERE business_id = $1;

-- name: ListProperties :many
SELECT * FROM properties ORDER BY created_at DESC;

-- name: SearchPropertiesByLocation :many
SELECT p.*, b.name as business_name,
       ST_Distance(ST_SetSRID(ST_MakePoint(a.longitude, a.latitude), 4326)::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance_meters
FROM properties p
JOIN businesses b ON p.business_id = b.id
JOIN addresses a ON p.address_id = a.id
WHERE ST_DWithin(ST_SetSRID(ST_MakePoint(a.longitude, a.latitude), 4326)::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $3)
ORDER BY distance_meters ASC;

-- Property Bookings
-- name: CreatePropertyBooking :one
INSERT INTO property_bookings (id, user_id, property_id, check_in_date, check_out_date, total_amount, currency, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPropertyBookingsByUserID :many
SELECT * FROM property_bookings WHERE user_id = $1 ORDER BY created_at DESC;

-- Food Items
-- name: CreateFoodItem :one
INSERT INTO food_items (id, business_id, name, description, price, currency, image_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListFoodItemsByBusiness :many
SELECT * FROM food_items WHERE business_id = $1;

-- Drivers & Taxi
-- name: CreateDriver :one
INSERT INTO drivers (id, user_id, name, status, vehicle_type_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateDriverLocation :one
UPDATE drivers
SET last_location = ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
RETURNING *;

-- name: UpdateDriverStatus :one
UPDATE drivers
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
RETURNING *;

-- name: GetNearbyDrivers :many
SELECT d.*, 
       ST_Distance(d.last_location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance_meters
FROM drivers d
WHERE d.status = 'online'
  AND ST_DWithin(d.last_location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $3)
ORDER BY distance_meters ASC
LIMIT $4;

-- Hubs
-- name: GetHubs :many
SELECT * FROM hubs;

-- name: CreateHub :one
INSERT INTO hubs (id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- Tasks
-- name: GetTasks :many
SELECT * FROM tasks;

-- name: CreateTask :one
INSERT INTO tasks (id, title, priority)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateTaxiTrip :one
INSERT INTO taxi_trips (id, user_id, driver_id, pickup_location, dropoff_location, total_amount, currency, status)
VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography, ST_SetSRID(ST_MakePoint($6, $7), 4326)::geography, $8, $9, 'requested')
RETURNING *;

-- name: GetTaxiTripByID :one
SELECT * FROM taxi_trips WHERE id = $1;

-- name: UpdateTaxiTripStatus :one
UPDATE taxi_trips
SET status = $2, 
    accepted_at = CASE WHEN $2 = 'accepted' THEN CURRENT_TIMESTAMP ELSE accepted_at END,
    started_at = CASE WHEN $2 = 'in_progress' THEN CURRENT_TIMESTAMP ELSE started_at END,
    completed_at = CASE WHEN $2 = 'completed' THEN CURRENT_TIMESTAMP ELSE completed_at END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- Cart & Orders
-- name: AddItemToCart :one
INSERT INTO cart_items (id, user_id, business_id, item_id, item_type, quantity)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, item_id, item_type) DO UPDATE SET
  quantity = cart_items.quantity + EXCLUDED.quantity,
  updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetCartItemsByUserID :many
SELECT id, user_id, business_id, item_id, item_type, quantity, created_at, updated_at FROM cart_items WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateCartItemQuantity :exec
UPDATE cart_items SET quantity = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: RemoveCartItem :exec
DELETE FROM cart_items WHERE id = $1;

-- name: ClearCart :exec
DELETE FROM cart_items WHERE user_id = $1;

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

-- C2C Marketplace
-- name: CreateC2CSeller :one
INSERT INTO c2c_sellers (id, user_id, bio, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetC2CSellerByUserID :one
SELECT * FROM c2c_sellers WHERE user_id = $1;

-- name: CreateC2CListing :one
INSERT INTO c2c_listings (id, seller_id, title, description, price, currency, image_urls, is_negotiable, location, condition, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetC2CListingByID :one
SELECT * FROM c2c_listings WHERE id = $1;

-- name: ListC2CListings :many
SELECT * FROM c2c_listings WHERE status = 'available' ORDER BY created_at DESC;

-- Reviews
-- name: AddReview :one
INSERT INTO reviews (id, target_id, target_type, user_id, rating, comment)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetReviewsByTarget :many
SELECT * FROM reviews WHERE target_id = $1 AND target_type = $2 ORDER BY created_at DESC;
