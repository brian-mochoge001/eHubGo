-- Existing Tables Queries
-- name: GetHubs :many
SELECT id, name, description FROM hubs;

-- name: CreateHub :one
INSERT INTO hubs (id, name, description)
VALUES ($1, $2, $3)
RETURNING id, name, description;

-- name: GetTasks :many
SELECT id, title, priority FROM tasks;

-- name: CreateTask :one
-- Corrected: Explicitly include 'id' in INSERT and VALUES for compatibility with SERIAL PK and RETURNING.
INSERT INTO tasks (id, title, priority)
VALUES ($1, $2, $3)
RETURNING id, title, priority;

-- Taxi Queries (Drivers)


-- name: UpdateDriverLocation :one
UPDATE drivers
SET last_location = ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, user_id, name, status, vehicle_make, vehicle_model, license_plate, rating, last_location, updated_at, created_at;

-- name: GetDriverLocation :one
SELECT id, user_id, name, status, vehicle_make, vehicle_model, license_plate, rating,
       ST_X(last_location::geometry) as longitude,
       ST_Y(last_location::geometry) as latitude,
       updated_at, created_at
FROM drivers
WHERE id = $1;

-- name: GetNearbyDrivers :many
SELECT d.id, d.user_id, d.name, d.status, d.vehicle_make, d.vehicle_model, d.license_plate, d.rating,
       ST_X(d.last_location::geometry) as longitude,
       ST_Y(d.last_location::geometry) as latitude,
       ST_Distance(d.last_location, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance_meters,
       d.updated_at, d.created_at
FROM drivers d
WHERE d.status = 'online'
ORDER BY d.last_location <-> ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
LIMIT $3;

-- name: GetDriverByUserID :one
SELECT id, user_id, name, status, vehicle_make, vehicle_model, license_plate, rating, last_location, updated_at, created_at
FROM drivers
WHERE user_id = $1;

-- New Tables Queries

-- Users
-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name, phone_number, profile_picture_url)
VALUES ($1::TEXT, $2::TEXT, $3::TEXT, $4::TEXT, $5::TEXT, $6::TEXT)
RETURNING id, email, first_name, last_name, phone_number, profile_picture_url, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, first_name, last_name, phone_number, profile_picture_url, created_at, updated_at FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, first_name, last_name, phone_number, profile_picture_url, created_at, updated_at FROM users
WHERE id = $1;

-- User Roles
-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role) VALUES ($1, $2);

-- name: GetUserRoles :many
SELECT role FROM user_roles WHERE user_id = $1;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles WHERE user_id = $1 AND role = $2;

-- Addresses
-- name: CreateAddress :one
INSERT INTO addresses (id, user_id, label, recipient_name, recipient_phone, address_line1, address_line2, city, state, postal_code, country, is_default, latitude, longitude)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id, user_id, label, recipient_name, recipient_phone, address_line1, address_line2, city, state, postal_code, country, is_default, latitude, longitude, created_at, updated_at;

-- name: GetAddressesByUserID :many
SELECT id, user_id, label, recipient_name, recipient_phone, address_line1, address_line2, city, state, postal_code, country, is_default, latitude, longitude, created_at, updated_at FROM addresses
WHERE user_id = $1;

-- name: GetAddressByID :one
SELECT id, user_id, label, recipient_name, recipient_phone, address_line1, address_line2, city, state, postal_code, country, is_default, latitude, longitude, created_at, updated_at FROM addresses
WHERE id = $1;

-- name: UpdateAddress :one
UPDATE addresses
SET label = $1, recipient_name = $2, recipient_phone = $3, address_line1 = $4, address_line2 = $5, city = $6, state = $7, postal_code = $8, country = $9, is_default = $10, latitude = $11, longitude = $12, updated_at = CURRENT_TIMESTAMP
WHERE id = $13
RETURNING id, user_id, label, recipient_name, recipient_phone, address_line1, address_line2, city, state, postal_code, country, is_default, latitude, longitude, created_at, updated_at;

-- name: DeleteAddress :exec
DELETE FROM addresses WHERE id = $1;

-- Categories
-- name: CreateCategory :one
INSERT INTO categories (id, name, description, image_url, parent_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, image_url, parent_id, created_at, updated_at;

-- name: GetCategoryByID :one
SELECT id, name, description, image_url, parent_id, created_at, updated_at FROM categories
WHERE id = $1;

-- name: GetCategories :many
SELECT id, name, description, image_url, parent_id, created_at, updated_at FROM categories
ORDER BY name;

-- name: GetSubcategories :many
SELECT id, name, description, image_url, parent_id, created_at, updated_at FROM categories
WHERE parent_id = $1
ORDER BY name;

-- Products
-- name: CreateProduct :one
INSERT INTO products (id, name, description, price, currency, stock_quantity, category_id, vendor_id, rating, review_count, is_flash_sale, discount_percentage, flash_sale_start_time, flash_sale_end_time)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id, name, description, price, currency, stock_quantity, category_id, vendor_id, rating, review_count, is_flash_sale, discount_percentage, flash_sale_start_time, flash_sale_end_time, created_at, updated_at;

-- name: GetProductByID :one
SELECT id, name, description, price, currency, stock_quantity, category_id, vendor_id, rating, review_count, is_flash_sale, discount_percentage, flash_sale_start_time, flash_sale_end_time, created_at, updated_at FROM products
WHERE id = $1;

-- name: GetProductsByCategory :many
SELECT id, name, description, price, currency, stock_quantity, category_id, vendor_id, rating, review_count, is_flash_sale, discount_percentage, flash_sale_start_time, flash_sale_end_time, created_at, updated_at FROM products
WHERE category_id = $1
ORDER BY created_at DESC;

-- name: GetProductsByVendor :many
SELECT id, name, description, price, currency, stock_quantity, category_id, vendor_id, rating, review_count, is_flash_sale, discount_percentage, flash_sale_start_time, flash_sale_end_time, created_at, updated_at FROM products
WHERE vendor_id = $1
ORDER BY created_at DESC;

-- name: GetFlashSaleProducts :many
SELECT id, name, description, price, currency, stock_quantity, category_id, vendor_id, rating, review_count, is_flash_sale, discount_percentage, flash_sale_start_time, flash_sale_end_time, created_at, updated_at FROM products
WHERE is_flash_sale = TRUE AND (flash_sale_end_time IS NULL OR flash_sale_end_time > CURRENT_TIMESTAMP)
ORDER BY flash_sale_end_time ASC, created_at DESC;

-- name: UpdateProductStock :exec
UPDATE products
SET stock_quantity = stock_quantity - $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND stock_quantity >= $2;

-- name: UpdateProductRating :exec
UPDATE products
SET rating = $2, review_count = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- Product Images
-- name: CreateProductImage :one
INSERT INTO product_images (id, product_id, image_url, is_thumbnail, sort_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, product_id, image_url, is_thumbnail, sort_order, created_at;

-- name: GetProductImagesByProductID :many
SELECT id, product_id, image_url, is_thumbnail, sort_order, created_at FROM product_images
WHERE product_id = $1
ORDER BY sort_order ASC, created_at ASC;

-- name: DeleteProductImagesByProductID :exec
DELETE FROM product_images WHERE product_id = $1;

-- Product Videos
-- name: CreateProductVideo :one
INSERT INTO product_videos (id, product_id, video_url)
VALUES ($1, $2, $3)
RETURNING id, product_id, video_url, created_at;

-- name: GetProductVideoByProductID :one
SELECT id, product_id, video_url, created_at FROM product_videos
WHERE product_id = $1;

-- name: DeleteProductVideoByProductID :exec
DELETE FROM product_videos WHERE product_id = $1;

-- Cart Items (Universal Cart)
-- name: AddItemToCart :one
INSERT INTO cart_items (id, user_id, item_id, item_type, quantity)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, item_id, item_type) DO UPDATE SET
  quantity = cart_items.quantity + EXCLUDED.quantity,
  updated_at = CURRENT_TIMESTAMP
RETURNING id, user_id, item_id, item_type, quantity, created_at, updated_at;

-- name: GetCartItemsByUserID :many
SELECT id, user_id, item_id, item_type, quantity, created_at, updated_at FROM cart_items
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateCartItemQuantity :exec
UPDATE cart_items
SET quantity = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND quantity + $2 > 0; -- Ensure quantity doesn't go below zero implicitly

-- name: RemoveCartItem :exec
DELETE FROM cart_items WHERE id = $1;

-- name: ClearCart :exec
DELETE FROM cart_items WHERE user_id = $1;

-- Orders
-- name: CreateOrder :one
INSERT INTO orders (id, user_id, total_amount, currency, status, shipping_address_id, delivery_fee, discount_amount, estimated_delivery_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, order_date, total_amount, currency, status, shipping_address_id, delivery_fee, discount_amount, estimated_delivery_date, created_at, updated_at;

-- name: GetOrderByID :one
SELECT id, user_id, order_date, total_amount, currency, status, shipping_address_id, delivery_fee, discount_amount, estimated_delivery_date, created_at, updated_at FROM orders
WHERE id = $1;

-- name: GetOrdersByUserID :many
SELECT id, user_id, order_date, total_amount, currency, status, shipping_address_id, delivery_fee, discount_amount, estimated_delivery_date, created_at, updated_at FROM orders
WHERE user_id = $1
ORDER BY order_date DESC;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- Order Items
-- name: CreateOrderItem :one
INSERT INTO order_items (id, order_id, item_id, item_type, quantity, unit_price)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, order_id, item_id, item_type, quantity, unit_price, created_at;

-- name: GetOrderItemsByOrderID :many
SELECT id, order_id, item_id, item_type, quantity, unit_price, created_at FROM order_items
WHERE order_id = $1;

-- Transactions
-- name: CreateTransaction :one
INSERT INTO transactions (id, user_id, order_id, transaction_type, amount, currency, status, payment_gateway_ref, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, order_id, transaction_type, amount, currency, status, payment_gateway_ref, description, created_at, updated_at;

-- name: GetTransactionsByUserID :many
SELECT id, user_id, order_id, transaction_type, amount, currency, status, payment_gateway_ref, description, created_at, updated_at FROM transactions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetUserWalletBalance :one
SELECT balance FROM user_wallet WHERE user_id = $1;

-- name: UpdateUserWalletBalance :one
UPDATE user_wallet
SET balance = balance + $2, last_updated = CURRENT_TIMESTAMP
WHERE user_id = $1
RETURNING user_id, balance, last_updated;

-- name: CreateUserWallet :one
INSERT INTO user_wallet (user_id, balance) VALUES ($1, $2)
RETURNING user_id, balance, last_updated;

-- Vendor specific queries
-- name: CreateVendor :one
INSERT INTO vendors (id, user_id, shop_name, description, logo_url, address_id, phone_number, email, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, shop_name, description, logo_url, address_id, phone_number, email, is_active, created_at, updated_at;

-- name: GetVendorByID :one
SELECT id, user_id, shop_name, description, logo_url, address_id, phone_number, email, is_active, created_at, updated_at FROM vendors
WHERE id = $1;

-- name: GetVendorByUserID :one
SELECT id, user_id, shop_name, description, logo_url, address_id, phone_number, email, is_active, created_at, updated_at FROM vendors
WHERE user_id = $1;

-- name: GetVendorProducts :many
SELECT p.id, p.name, p.description, p.price, p.currency, p.stock_quantity, p.category_id, p.vendor_id, p.rating, p.review_count, p.is_flash_sale, p.discount_percentage, p.flash_sale_start_time, p.flash_sale_end_time, p.created_at, p.updated_at
FROM products p
WHERE p.vendor_id = $1
ORDER BY p.created_at DESC;

-- name: UpdateVendor :one
UPDATE vendors
SET shop_name = $1, description = $2, logo_url = $3, address_id = $4, phone_number = $5, email = $6, is_active = $7, updated_at = CURRENT_TIMESTAMP
WHERE id = $8
RETURNING id, user_id, shop_name, description, logo_url, address_id, phone_number, email, is_active, created_at, updated_at;

-- C2C Sellers
-- name: CreateC2CSeller :one
INSERT INTO c2c_sellers (id, user_id, bio, avatar_url, reputation_score)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, bio, avatar_url, reputation_score, created_at, updated_at;

-- name: GetC2CSellerByUserID :one
SELECT id, user_id, bio, avatar_url, reputation_score, created_at, updated_at FROM c2c_sellers
WHERE user_id = $1;

-- name: GetC2CListingsByUserID :many
-- Assuming a c2c_listings table exists, not created in schema.sql yet.
-- SELECT id, seller_id, title, description, price, location, created_at FROM c2c_listings
-- WHERE seller_id = $1 ORDER BY created_at DESC;

-- Hosts
-- name: CreateHost :one
INSERT INTO hosts (id, user_id, bio, response_rate, is_superhost)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, bio, response_rate, is_superhost, created_at, updated_at;

-- name: GetHostByUserID :one
SELECT id, user_id, bio, response_rate, is_superhost, created_at, updated_at FROM hosts
WHERE user_id = $1;

-- name: GetPropertiesByHost :many
SELECT id, host_id, title, description, address_id, latitude, longitude, price_per_night, currency, number_of_guests, number_of_bedrooms, number_of_beds, number_of_bathrooms, image_urls, video_url, available_from, available_to, created_at, updated_at FROM properties
WHERE host_id = $1
ORDER BY created_at DESC;

-- Properties
-- name: CreateProperty :one
INSERT INTO properties (id, host_id, title, description, address_id, latitude, longitude, price_per_night, currency, number_of_guests, number_of_bedrooms, number_of_beds, number_of_bathrooms, image_urls, video_url, available_from, available_to)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING id, host_id, title, description, address_id, latitude, longitude, price_per_night, currency, number_of_guests, number_of_bedrooms, number_of_beds, number_of_bathrooms, image_urls, video_url, available_from, available_to, created_at, updated_at;

-- name: GetPropertyByID :one
SELECT id, host_id, title, description, address_id, latitude, longitude, price_per_night, currency, number_of_guests, number_of_bedrooms, number_of_beds, number_of_bathrooms, image_urls, video_url, available_from, available_to, created_at, updated_at FROM properties
WHERE id = $1;

-- name: UpdatePropertyAvailability :exec
UPDATE properties
SET available_from = $2, available_to = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- Food Items
-- name: CreateFoodItem :one
INSERT INTO food_items (id, vendor_id, name, description, price, currency, image_url, menu_category_id, is_available)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, vendor_id, name, description, price, currency, image_url, menu_category_id, is_available, created_at, updated_at;

-- name: GetFoodItemsByVendor :many
SELECT id, vendor_id, name, description, price, currency, image_url, menu_category_id, is_available, created_at, updated_at FROM food_items
WHERE vendor_id = $1 AND is_available = TRUE
ORDER BY created_at DESC;

-- Food Item Queries
-- name: GetFoodItemByID :one
SELECT id, vendor_id, name, description, price, currency, image_url, menu_category_id, is_available, created_at, updated_at FROM food_items
WHERE id = $1;

-- Better Auth Tables
CREATE TABLE "user" (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    image TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE session (
    id TEXT PRIMARY KEY,
    expires_at TIMESTAMP NOT NULL,
    token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE TABLE account (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    access_token TEXT,
    refresh_token TEXT,
    id_token TEXT,
    access_token_expires_at TIMESTAMP,
    refresh_token_expires_at TIMESTAMP,
    scope TEXT,
    password TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE verification (
    id TEXT PRIMARY KEY,
    identifier TEXT NOT NULL,
    value TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);









-- Property Bookings (for eHost mini-service)
-- name: CreatePropertyBooking :one
INSERT INTO property_bookings (id, user_id, property_id, check_in_date, check_out_date, total_amount, currency, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, user_id, property_id, check_in_date, check_out_date, total_amount, currency, status, created_at, updated_at;

-- name: GetPropertyBookingByID :one
SELECT id, user_id, property_id, check_in_date, check_out_date, total_amount, currency, status, created_at, updated_at FROM property_bookings
WHERE id = $1;

-- name: GetPropertyBookingsByUserID :many
SELECT pb.*, p.title as property_title, p.image_urls as property_images FROM property_bookings pb
JOIN properties p ON pb.property_id = p.id
WHERE pb.user_id = $1
ORDER BY pb.check_in_date DESC;

-- name: GetPropertyBookingsByPropertyID :many
SELECT pb.*, u.first_name, u.last_name, u.email FROM property_bookings pb
JOIN users u ON pb.user_id = u.id
WHERE pb.property_id = $1 AND pb.status = 'confirmed' AND pb.check_out_date >= CURRENT_DATE
ORDER BY pb.check_in_date ASC;

-- name: UpdatePropertyBookingStatus :exec
UPDATE property_bookings
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- Generic Service Bookings
-- name: CreateServiceBooking :one
INSERT INTO service_bookings (id, user_id, service_type, service_item_id, provider_id, provider_type, start_time, end_time, total_amount, currency, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, user_id, service_type, service_item_id, provider_id, provider_type, start_time, end_time, total_amount, currency, status, created_at, updated_at;

-- name: GetServiceBookingByID :one
SELECT id, user_id, service_type, service_item_id, provider_id, provider_type, start_time, end_time, total_amount, currency, status, created_at, updated_at FROM service_bookings
WHERE id = $1;

-- name: GetServiceBookingsByUserID :many
SELECT sb.*, s.name as service_name FROM service_bookings sb
JOIN services s ON sb.service_item_id = s.id -- This assumes service_item_id directly maps to services.id. Adjust if using a different approach.
WHERE sb.user_id = $1
ORDER BY sb.start_time DESC;

-- name: UpdateServiceBookingStatus :exec
UPDATE service_bookings
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- Shipments
-- name: CreateShipment :one
INSERT INTO shipments (id, user_id, tracking_number, sender_address_id, recipient_address_id, current_location, status, estimated_delivery_date)
VALUES ($1, $2, $3, $4, $5, ST_SetSRID(ST_MakePoint($6, $7), 4326)::geography, $8, $9)
RETURNING id, user_id, tracking_number, sender_address_id, recipient_address_id, ST_X(current_location::geometry) as longitude, ST_Y(current_location::geometry) as latitude, status, estimated_delivery_date, created_at, updated_at;

-- name: GetShipmentByTrackingNumber :one
SELECT id, user_id, tracking_number, sender_address_id, recipient_address_id, ST_X(current_location::geometry) as longitude, ST_Y(current_location::geometry) as latitude, status, estimated_delivery_date, created_at, updated_at FROM shipments
WHERE tracking_number = $1;

-- name: UpdateShipmentStatus :exec
UPDATE shipments
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateShipmentLocation :exec
UPDATE shipments
SET current_location = ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetShipmentsByUserID :many
SELECT id, user_id, tracking_number, status, estimated_delivery_date, created_at FROM shipments
WHERE user_id = $1
ORDER BY created_at DESC;

-- Jobs & Companies
-- name: CreateCompany :one
INSERT INTO companies (id, name, logo_url, description, website, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, logo_url, description, website, user_id, created_at, updated_at;

-- name: GetCompanyByID :one
SELECT id, name, logo_url, description, website, user_id, created_at, updated_at FROM companies
WHERE id = $1;

-- name: CreateJob :one
INSERT INTO jobs (id, company_id, title, description, location, salary_min, salary_max, currency, job_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, company_id, title, description, location, salary_min, salary_max, currency, job_type, posted_at, created_at, updated_at;

-- name: GetJobByID :one
SELECT j.*, c.name as company_name, c.logo_url as company_logo_url
FROM jobs j
JOIN companies c ON j.company_id = c.id
WHERE j.id = $1;

-- name: GetJobsByCompany :many
SELECT j.*, c.name as company_name, c.logo_url as company_logo_url
FROM jobs j
JOIN companies c ON j.company_id = c.id
WHERE j.company_id = $1
ORDER BY j.posted_at DESC;

-- name: GetRecentJobs :many
SELECT j.*, c.name as company_name, c.logo_url as company_logo_url
FROM jobs j
JOIN companies c ON j.company_id = c.id
ORDER BY j.posted_at DESC
LIMIT $1;

-- B2B Quotes
-- name: CreateB2BQuote :one
INSERT INTO b2b_quotes (id, vendor_id, user_id, request_details, status, quoted_amount, currency, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, vendor_id, user_id, request_details, status, quoted_amount, currency, expires_at, created_at, updated_at;

-- name: GetB2BQuoteByID :one
SELECT bq.*, v.shop_name as vendor_shop_name, u.first_name as user_first_name, u.last_name as user_last_name
FROM b2b_quotes bq
JOIN vendors v ON bq.vendor_id = v.id
JOIN users u ON bq.user_id = u.id
WHERE bq.id = $1;

-- name: GetB2BQuotesByUserID :many
SELECT bq.*, v.shop_name as vendor_shop_name
FROM b2b_quotes bq
JOIN vendors v ON bq.vendor_id = v.id
WHERE bq.user_id = $1
ORDER BY bq.created_at DESC;

-- name: UpdateB2BQuoteStatus :exec
UPDATE b2b_quotes
SET status = $2, quoted_amount = $3, currency = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetB2BQuotesByVendorID :many
SELECT bq.*, u.first_name as user_first_name, u.last_name as user_last_name, u.email as user_email
FROM b2b_quotes bq
JOIN users u ON bq.user_id = u.id
WHERE bq.vendor_id = $1
ORDER BY bq.created_at DESC;

-- C2C Listings
-- name: CreateC2CListing :one
INSERT INTO c2c_listings (id, seller_id, category_id, brand_id, title, description, price, currency, is_negotiable, location, image_urls, condition, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id, seller_id, category_id, brand_id, title, description, price, currency, is_negotiable, location, image_urls, condition, status, created_at, updated_at;

-- name: GetC2CListingByID :one
SELECT cl.*, cs.bio as seller_bio, cs.avatar_url as seller_avatar_url, cs.reputation_score as seller_reputation
FROM c2c_listings cl
JOIN c2c_sellers cs ON cl.seller_id = cs.id
WHERE cl.id = $1;

-- name: GetC2CListingsBySellerID :many
SELECT cl.*, cs.bio as seller_bio, cs.avatar_url as seller_avatar_url, cs.reputation_score as seller_reputation
FROM c2c_listings cl
JOIN c2c_sellers cs ON cl.seller_id = cs.id
WHERE cl.seller_id = $1
ORDER BY cl.created_at DESC;

-- name: UpdateC2CListing :one
UPDATE c2c_listings
SET seller_id = $1, category_id = $2, brand_id = $3, title = $4, description = $5, price = $6, currency = $7, is_negotiable = $8, location = $9, image_urls = $10, condition = $11, status = $12, updated_at = CURRENT_TIMESTAMP
WHERE id = $13
RETURNING id, seller_id, category_id, brand_id, title, description, price, currency, is_negotiable, location, image_urls, condition, status, created_at, updated_at;

-- name: DeleteC2CListing :exec
DELETE FROM c2c_listings WHERE id = $1;

-- Messages
-- name: CreateMessage :one
INSERT INTO messages (id, sender_id, receiver_id, content)
VALUES ($1, $2, $3, $4)
RETURNING id, sender_id, receiver_id, content, sent_at;

-- name: GetMessagesBetweenUsers :many
SELECT * FROM messages
WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
ORDER BY sent_at ASC;

-- name: MarkMessagesAsRead :exec
UPDATE messages
SET read_at = CURRENT_TIMESTAMP
WHERE receiver_id = $1 AND sender_id = $2 AND read_at IS NULL;

-- name: GetUnreadMessageCount :one
SELECT COUNT(*) FROM messages
WHERE receiver_id = $1 AND read_at IS NULL;

-- Notifications
-- name: CreateNotification :one
INSERT INTO notifications (id, user_id, type, title, body)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, type, title, body, read_at, created_at;

-- name: GetNotificationsByUserID :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: MarkNotificationAsRead :exec
UPDATE notifications
SET read_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET read_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND read_at IS NULL;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id = $1 AND user_id = $2;

-- Services
-- name: CreateService :one
INSERT INTO services (id, provider_user_id, service_type, name, description, base_price, currency, location, availability_details, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, ST_SetSRID(ST_MakePoint($8, $9), 4326)::geography, $10, $11)
RETURNING id, provider_user_id, service_type, name, description, base_price, currency, location, availability_details, is_active, created_at, updated_at;

-- name: GetServiceByID :one
SELECT id, provider_user_id, service_type, name, description, base_price, currency, ST_X(location::geometry) as longitude, ST_Y(location::geometry) as latitude, availability_details, is_active, created_at, updated_at FROM services
WHERE id = $1;

-- name: GetServicesByType :many
SELECT s.*, u.first_name as provider_first_name, u.last_name as provider_last_name, v.shop_name as provider_shop_name
FROM services s
LEFT JOIN users u ON s.provider_user_id = u.id
LEFT JOIN vendors v ON u.id = v.user_id -- Assuming provider_user_id can link to a vendor
WHERE s.service_type = $1 AND s.is_active = TRUE
ORDER BY s.created_at DESC;

-- name: UpdateService :one
UPDATE services
SET name = $1, description = $2, base_price = $3, currency = $4, location = ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography, availability_details = $7, is_active = $8, updated_at = CURRENT_TIMESTAMP
WHERE id = $9
RETURNING id, provider_user_id, service_type, name, description, base_price, currency, location, availability_details, is_active, created_at, updated_at;

-- name: DeleteService :exec
DELETE FROM services WHERE id = $1;

-- Vendor Miniservice Verification
-- name: CreateVendorMiniserviceVerification :one
INSERT INTO vendor_miniservice_verification (vendor_id, miniservice_type, verification_status, verified_by_user_id, verified_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING vendor_id, miniservice_type, verification_status, verified_by_user_id, verified_at, created_at, updated_at;

-- name: GetVendorVerificationStatus :one
SELECT verification_status FROM vendor_miniservice_verification
WHERE vendor_id = $1 AND miniservice_type = $2;

-- name: UpdateVendorVerificationStatus :one
UPDATE vendor_miniservice_verification
SET verification_status = $1, verified_by_user_id = $2, verified_at = $3, updated_at = CURRENT_TIMESTAMP
WHERE vendor_id = $4 AND miniservice_type = $5
RETURNING vendor_id, miniservice_type, verification_status, verified_by_user_id, verified_at, created_at, updated_at;

-- name: ListVendorVerifications :many
SELECT * FROM vendor_miniservice_verification
WHERE vendor_id = $1;

-- name: ListMiniserviceVerifications :many
SELECT v.shop_name, vm.*
FROM vendor_miniservice_verification vm
JOIN vendors v ON vm.vendor_id = v.id
WHERE vm.miniservice_type = $1 AND vm.verification_status = 'verified';

-- name: RemoveVendorMiniserviceVerification :exec
DELETE FROM vendor_miniservice_verification WHERE vendor_id = $1 AND miniservice_type = $2;

-- Queries for Brands Table
-- name: CreateBrand :one
INSERT INTO brands (id, name, logo_url)
VALUES ($1, $2, $3)
RETURNING id, name, logo_url, created_at, updated_at;

-- name: GetBrandByID :one
SELECT id, name, logo_url, created_at, updated_at FROM brands
WHERE id = $1;

-- name: GetBrandByName :one
SELECT id, name, logo_url, created_at, updated_at FROM brands
WHERE name = $1;

-- name: ListBrands :many
SELECT id, name, logo_url, created_at, updated_at FROM brands
ORDER BY name;

-- name: UpdateBrand :one
UPDATE brands
SET name = $1, logo_url = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $3
RETURNING id, name, logo_url, created_at, updated_at;

-- name: DeleteBrand :exec
DELETE FROM brands WHERE id = $1;


-- Queries for Product Variants Table
-- name: CreateProductVariant :one
INSERT INTO product_variants (id, product_id, sku, price, stock_quantity, attributes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, product_id, sku, price, stock_quantity, attributes, created_at, updated_at;

-- name: GetProductVariantByID :one
SELECT id, product_id, sku, price, stock_quantity, attributes, created_at, updated_at FROM product_variants
WHERE id = $1;

-- name: GetProductVariantsByProductID :many
SELECT id, product_id, sku, price, stock_quantity, attributes, created_at, updated_at FROM product_variants
WHERE product_id = $1
ORDER BY created_at DESC;

-- name: UpdateProductVariant :one
UPDATE product_variants
SET sku = $1, price = $2, stock_quantity = $3, attributes = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $5
RETURNING id, product_id, sku, price, stock_quantity, attributes, created_at, updated_at;

-- name: DeleteProductVariant :exec
DELETE FROM product_variants WHERE id = $1;


-- Queries for Product Discounts Table
-- name: CreateProductDiscount :one
INSERT INTO product_discounts (id, product_id, discount_type, discount_value, start_at, end_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, product_id, discount_type, discount_value, start_at, end_at, created_at, updated_at;

-- name: GetProductDiscountByID :one
SELECT id, product_id, discount_type, discount_value, start_at, end_at, created_at, updated_at FROM product_discounts
WHERE id = $1;

-- name: GetProductDiscountsByProductID :many
SELECT id, product_id, discount_type, discount_value, start_at, end_at, created_at, updated_at FROM product_discounts
WHERE product_id = $1 AND (end_at IS NULL OR end_at > CURRENT_TIMESTAMP) -- Consider only active discounts
ORDER BY start_at DESC, created_at DESC;

-- name: UpdateProductDiscount :one
UPDATE product_discounts
SET discount_type = $1, discount_value = $2, start_at = $3, end_at = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $5
RETURNING id, product_id, discount_type, discount_value, start_at, end_at, created_at, updated_at;

-- name: DeleteProductDiscount :exec
DELETE FROM product_discounts WHERE id = $1;

-- name: GetProductByIDWithDetails :one
SELECT
    p.id, p.name, p.description, p.price, p.currency, p.stock_quantity, p.category_id, p.vendor_id, p.rating, p.review_count,
    p.is_flash_sale, p.discount_percentage, p.flash_sale_start_time, p.flash_sale_end_time,
    p.brand_id, b.name as brand_name, b.logo_url as brand_logo_url,
    p.created_at, p.updated_at
FROM products p
LEFT JOIN brands b ON p.brand_id = b.id
WHERE p.id = $1;

-- name: GetProductsWithBrandAndCategory :many
SELECT
    p.id, p.name, p.description, p.price, p.currency, p.stock_quantity, p.vendor_id, p.rating, p.review_count,
    p.is_flash_sale, p.discount_percentage, p.flash_sale_start_time, p.flash_sale_end_time,
    b.name as brand_name, b.logo_url as brand_logo_url,
    c.name as category_name
FROM products p
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN categories c ON p.category_id = c.id
ORDER BY p.created_at DESC;

-- Queries for Vehicle Types (eTaxi)
-- name: CreateVehicleType :one
INSERT INTO vehicle_types (id, name, passenger_capacity, base_price_per_km, base_price_per_minute, initial_fee, currency)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, name, passenger_capacity, base_price_per_km, base_price_per_minute, initial_fee, currency, created_at, updated_at;

-- name: GetVehicleTypeByID :one
SELECT id, name, passenger_capacity, base_price_per_km, base_price_per_minute, initial_fee, currency, created_at, updated_at FROM vehicle_types
WHERE id = $1;

-- name: ListVehicleTypes :many
SELECT id, name, passenger_capacity, base_price_per_km, base_price_per_minute, initial_fee, currency, created_at, updated_at FROM vehicle_types
ORDER BY name;

-- name: UpdateVehicleType :one
UPDATE vehicle_types
SET name = $1, passenger_capacity = $2, base_price_per_km = $3, base_price_per_minute = $4, initial_fee = $5, currency = $6, updated_at = CURRENT_TIMESTAMP
WHERE id = $7
RETURNING id, name, passenger_capacity, base_price_per_km, base_price_per_minute, initial_fee, currency, created_at, updated_at;

-- name: DeleteVehicleType :exec
DELETE FROM vehicle_types WHERE id = $1;

-- Queries for Drivers table (enhanced)
-- name: CreateDriver :one
INSERT INTO drivers (id, user_id, name, status, vehicle_make, vehicle_model, license_plate, vehicle_color, vehicle_condition, vehicle_type_id, rating, last_location)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, ST_SetSRID(ST_MakePoint($12, $13), 4326)::geography)
RETURNING id, user_id, name, status, vehicle_make, vehicle_model, license_plate, vehicle_color, vehicle_condition, vehicle_type_id, rating, last_location, updated_at, created_at;

-- name: GetDriverByID :one
SELECT
    d.id, d.user_id, d.name, d.status, d.vehicle_make, d.vehicle_model, d.license_plate, d.vehicle_color, d.vehicle_condition, d.vehicle_type_id, d.rating,
    ST_X(d.last_location::geometry) as longitude, ST_Y(d.last_location::geometry) as latitude,
    vt.name as vehicle_type_name, -- Include vehicle type name
    d.updated_at, d.created_at
FROM drivers d
LEFT JOIN vehicle_types vt ON d.vehicle_type_id = vt.id
WHERE d.id = $1;

-- name: UpdateDriver :one
UPDATE drivers
SET name = $1, status = $2, vehicle_make = $3, vehicle_model = $4, license_plate = $5, vehicle_color = $6, vehicle_condition = $7, vehicle_type_id = $8, rating = $9, last_location = ST_SetSRID(ST_MakePoint($10, $11), 4326)::geography, updated_at = CURRENT_TIMESTAMP
WHERE id = $12
RETURNING id, user_id, name, status, vehicle_make, vehicle_model, license_plate, vehicle_color, vehicle_condition, vehicle_type_id, rating, last_location, updated_at, created_at;

-- name: UpdateDriverRating :exec
UPDATE drivers
SET rating = $2, updated_at = CURRENT_TIMESTAMP -- Assuming rating is average, recalculate by app logic based on driver_ratings
WHERE id = $1;

-- Queries for Driver Ratings
-- name: CreateDriverRating :one
INSERT INTO driver_ratings (id, driver_id, user_id, rating, testimonial)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, driver_id, user_id, rating, testimonial, created_at;

-- name: GetDriverRatingByID :one
SELECT id, driver_id, user_id, rating, testimonial, created_at FROM driver_ratings
WHERE id = $1;

-- name: GetDriverRatingsByDriverID :many
SELECT dr.id, dr.driver_id, dr.user_id, dr.rating, dr.testimonial, dr.created_at, u.first_name, u.last_name, u.profile_picture_url
FROM driver_ratings dr
JOIN users u ON dr.user_id = u.id
WHERE dr.driver_id = $1
ORDER BY dr.created_at DESC;

-- name: GetAverageDriverRating :one
SELECT AVG(rating) as average_rating, COUNT(*) as total_ratings FROM driver_ratings
WHERE driver_id = $1;

-- name: DeleteDriverRating :exec
DELETE FROM driver_ratings WHERE id = $1 AND user_id = $2; -- Ensure user can only delete their own rating


-- Queries for Taxi Trips (enhanced)
-- name: CreateTaxiTrip :one
INSERT INTO taxi_trips (id, user_id, driver_id, pickup_location, dropoff_location, requested_at, accepted_at, started_at, completed_at, total_amount, currency, status, vehicle_type_id, cancellation_reason, cancelled_by)
VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography, ST_SetSRID(ST_MakePoint($6, $7), 4326)::geography, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING id, user_id, driver_id, vehicle_type_id, requested_at, accepted_at, started_at, completed_at, total_amount, currency, status, cancellation_reason, cancelled_by, created_at, updated_at;

-- name: GetTaxiTripByID :one
SELECT
    tt.id, tt.user_id, tt.driver_id, tt.vehicle_type_id, tt.requested_at, tt.accepted_at, tt.started_at, tt.completed_at,
    tt.total_amount, tt.currency, tt.status, tt.cancellation_reason, tt.cancelled_by,
    ST_X(tt.pickup_location::geometry) as pickup_longitude, ST_Y(tt.pickup_location::geometry) as pickup_latitude,
    ST_X(tt.dropoff_location::geometry) as dropoff_longitude, ST_Y(tt.dropoff_location::geometry) as dropoff_latitude,
    vt.name as vehicle_type_name -- Include vehicle type name
FROM taxi_trips tt
LEFT JOIN vehicle_types vt ON tt.vehicle_type_id = vt.id
WHERE tt.id = $1;

-- name: GetTaxiTripsByUserID :many
SELECT
    tt.id, tt.user_id, tt.driver_id, tt.vehicle_type_id, tt.requested_at, tt.accepted_at, tt.started_at, tt.completed_at,
    tt.total_amount, tt.currency, tt.status, tt.cancellation_reason, tt.cancelled_by,
    ST_X(tt.pickup_location::geometry) as pickup_longitude, ST_Y(tt.pickup_location::geometry) as pickup_latitude,
    ST_X(tt.dropoff_location::geometry) as dropoff_longitude, ST_Y(tt.dropoff_location::geometry) as dropoff_latitude,
    vt.name as vehicle_type_name
FROM taxi_trips tt
LEFT JOIN vehicle_types vt ON tt.vehicle_type_id = vt.id
WHERE tt.user_id = $1
ORDER BY tt.requested_at DESC;

-- name: GetTaxiTripsByDriverID :many
SELECT
    tt.id, tt.user_id, tt.driver_id, tt.vehicle_type_id, tt.requested_at, tt.accepted_at, tt.started_at, tt.completed_at,
    tt.total_amount, tt.currency, tt.status, tt.cancellation_reason, tt.cancelled_by,
    ST_X(tt.pickup_location::geometry) as pickup_longitude, ST_Y(tt.pickup_location::geometry) as pickup_latitude,
    ST_X(tt.dropoff_location::geometry) as dropoff_longitude, ST_Y(tt.dropoff_location::geometry) as dropoff_latitude,
    vt.name as vehicle_type_name
FROM taxi_trips tt
LEFT JOIN vehicle_types vt ON tt.vehicle_type_id = vt.id
WHERE tt.driver_id = $1
ORDER BY tt.requested_at DESC;

-- name: UpdateTaxiTripStatus :exec
UPDATE taxi_trips
SET status = $2,
    accepted_at = CASE WHEN $2 = 'accepted' THEN CURRENT_TIMESTAMP ELSE accepted_at END,
    started_at = CASE WHEN $2 = 'in_progress' THEN CURRENT_TIMESTAMP ELSE started_at END,
    completed_at = CASE WHEN $2 = 'completed' THEN CURRENT_TIMESTAMP ELSE completed_at END,
    cancellation_reason = CASE WHEN $2 = 'cancelled' THEN $3 ELSE cancellation_reason END, -- $3 is cancellation_reason
    cancelled_by = CASE WHEN $2 = 'cancelled' THEN $4 ELSE cancelled_by END, -- $4 is cancelled_by
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CancelTaxiTrip :exec
UPDATE taxi_trips
SET status = 'cancelled',
    cancellation_reason = $2,
    cancelled_by = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateTransactionsForOrderPayment :exec
-- This is a placeholder. Actual wallet logic will be complex and handled by application logic within transactions.
UPDATE transactions SET status = $1, payment_gateway_ref = $2 WHERE id = $3;