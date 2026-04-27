-- Users
-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name, phone_number, profile_picture_url)
VALUES (@email, @password_hash, @first_name, @last_name, @phone_number, @profile_picture_url)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = @id;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = @email;

-- User Roles
-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role) VALUES (@user_id, @role);

-- name: GetUserRoles :many
SELECT role FROM user_roles WHERE user_id = @user_id;

-- Businesses (The "Stalls" in the Mall)
-- name: CreateBusiness :one
INSERT INTO businesses (id, owner_id, name, description, logo_url, banner_url, miniservice_type, address_id, phone_number, email)
VALUES (@id, @owner_id, @name, @description, @logo_url, @banner_url, @miniservice_type, @address_id, @phone_number, @email)
RETURNING *;

-- name: GetBusinessByID :one
SELECT * FROM businesses WHERE id = @id;

-- name: GetBusinessesByOwnerID :many
SELECT * FROM businesses WHERE owner_id = @owner_id;

-- name: UpdateBusinessStatus :one
UPDATE businesses 
SET verification_status = @status, updated_at = CURRENT_TIMESTAMP
WHERE id = @id
RETURNING *;

-- name: ListBusinessesByType :many
SELECT * FROM businesses WHERE miniservice_type = @miniservice_type AND verification_status = 'approved';

-- Products
-- name: CreateProduct :one
INSERT INTO products (id, business_id, name, description, price, currency, stock_quantity, category_id, brand_id, is_flash_sale, discount_percentage, rating)
VALUES (@id, @business_id, @name, @description, @price, @currency, @stock_quantity, @category_id, @brand_id, @is_flash_sale, @discount_percentage, @rating)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET name = @name, description = @description, price = @price, stock_quantity = @stock_quantity, category_id = @category_id, image_urls = @image_urls, updated_at = CURRENT_TIMESTAMP
WHERE id = @id
RETURNING *;

-- name: GetProductsByBusiness :many
SELECT * FROM products WHERE business_id = @business_id;

-- name: GetProductByIDWithDetails :one
SELECT p.*, b.name as brand_name, c.name as category_name
FROM products p
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.id = @id;

-- name: GetFeaturedProducts :many
SELECT p.*, b.name as brand_name, c.name as category_name
FROM products p
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.rating >= 4.0
LIMIT @limit_count OFFSET @offset_count;

-- name: GetFlashSaleProducts :many
SELECT * FROM products WHERE is_flash_sale = TRUE AND (flash_sale_end_time IS NULL OR flash_sale_end_time > CURRENT_TIMESTAMP)
LIMIT @limit_count OFFSET @offset_count;

-- name: GetProducts :many
SELECT p.*, b.name as brand_name, c.name as category_name
FROM products p
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN categories c ON p.category_id = c.id
LIMIT @limit_count OFFSET @offset_count;

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
VALUES (@id, @name, @logo_url)
RETURNING *;

-- name: GetBrandByID :one
SELECT * FROM brands WHERE id = @id;

-- name: ListBrands :many
SELECT * FROM brands ORDER BY name;

-- name: UpdateBrand :one
UPDATE brands
SET name = @name, logo_url = @logo_url, updated_at = CURRENT_TIMESTAMP
WHERE id = @id
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = @id;

-- Product Variants
-- name: CreateProductVariant :one
INSERT INTO product_variants (id, product_id, sku, price, stock_quantity, attributes)
VALUES (@id, @product_id, @sku, @price, @stock_quantity, @attributes)
RETURNING *;

-- name: GetProductVariantsByProductID :many
SELECT * FROM product_variants WHERE product_id = @product_id;

-- name: GetProductVariantByID :one
SELECT * FROM product_variants WHERE id = @id;

-- name: UpdateProductVariant :one
UPDATE product_variants
SET sku = @sku, price = @price, stock_quantity = @stock_quantity, attributes = @attributes, updated_at = CURRENT_TIMESTAMP
WHERE id = @id
RETURNING *;

-- name: DeleteProductVariant :exec
DELETE FROM product_variants WHERE id = @id;

-- Product Discounts
-- name: CreateProductDiscount :one
INSERT INTO product_discounts (id, product_id, discount_type, discount_value, start_at, end_at)
VALUES (@id, @product_id, @discount_type, @discount_value, @start_at, @end_at)
RETURNING *;

-- name: GetProductDiscountsByProductID :many
SELECT * FROM product_discounts WHERE product_id = @product_id;

-- name: GetProductDiscountByID :one
SELECT * FROM product_discountS WHERE id = @id;

-- name: UpdateProductDiscount :one
UPDATE product_discounts
SET discount_type = @discount_type, discount_value = @discount_value, start_at = @start_at, end_at = @end_at, updated_at = CURRENT_TIMESTAMP
WHERE id = @id
RETURNING *;

-- name: DeleteProductDiscount :exec
DELETE FROM product_discounts WHERE id = @id;

-- Services
-- name: CreateService :one
INSERT INTO services (id, business_id, service_type, name, description, base_price, currency, location)
VALUES (@id, @business_id, @service_type, @name, @description, @base_price, @currency, @location)
RETURNING *;

-- name: ListServicesByBusiness :many
SELECT * FROM services WHERE business_id = @business_id;

-- name: ListServicesByType :many
SELECT * FROM services WHERE service_type = @service_type AND is_active = TRUE;

-- name: GetServiceByID :one
SELECT * FROM services WHERE id = @id;

-- name: DeleteService :exec
DELETE FROM services WHERE id = @id;

-- Service Bookings
-- name: CreateServiceBooking :one
INSERT INTO service_bookings (id, user_id, service_type, service_item_id, provider_id, provider_type, start_time, end_time, total_amount, currency, status)
VALUES (@id, @user_id, @service_type, @service_item_id, @provider_id, @provider_type, @start_time, @end_time, @total_amount, @currency, @status)
RETURNING *;

-- name: GetServiceBookingsByUserID :many
SELECT * FROM service_bookings WHERE user_id = @user_id ORDER BY created_at DESC;

-- name: GetServiceBookingsByProviderID :many
SELECT * FROM service_bookings WHERE provider_id = @provider_id ORDER BY created_at DESC;

-- name: UpdateServiceBookingStatus :one
UPDATE service_bookings
SET status = @status, updated_at = CURRENT_TIMESTAMP
WHERE id = @id
RETURNING *;

-- Properties
-- name: CreateProperty :one
INSERT INTO properties (id, business_id, title, description, address_id, price_per_night, currency, number_of_guests, number_of_bedrooms, type, image_urls)
VALUES (@id, @business_id, @title, @description, @address_id, @price_per_night, @currency, @number_of_guests, @number_of_bedrooms, @type, @image_urls)
RETURNING *;

-- name: GetPropertyByID :one
SELECT * FROM properties WHERE id = @id;

-- name: ListPropertiesByBusiness :many
SELECT * FROM properties WHERE business_id = @business_id;

-- name: ListProperties :many
SELECT * FROM properties ORDER BY created_at DESC;

-- name: SearchPropertiesByLocation :many
SELECT p.*, b.name as business_name,
       ST_Distance(ST_SetSRID(ST_MakePoint(a.longitude, a.latitude), 4326)::geography, ST_SetSRID(ST_MakePoint(@lng, @lat), 4326)::geography) as distance_meters
FROM properties p
JOIN businesses b ON p.business_id = b.id
JOIN addresses a ON p.address_id = a.id
WHERE ST_DWithin(ST_SetSRID(ST_MakePoint(a.longitude, a.latitude), 4326)::geography, ST_SetSRID(ST_MakePoint(@lng, @lat), 4326)::geography, @radius)
ORDER BY distance_meters ASC;

-- Property Bookings
-- name: CreatePropertyBooking :one
INSERT INTO property_bookings (id, user_id, property_id, check_in_date, check_out_date, total_amount, currency, status)
VALUES (@id, @user_id, @property_id, @check_in_date, @check_out_date, @total_amount, @currency, @status)
RETURNING *;

-- name: GetPropertyBookingsByUserID :many
SELECT * FROM property_bookings WHERE user_id = @user_id ORDER BY created_at DESC;

-- Food Items
-- name: CreateFoodItem :one
INSERT INTO food_items (id, business_id, name, description, price, currency, image_url, is_available)
VALUES (@id, @business_id, @name, @description, @price, @currency, @image_url, @is_available)
RETURNING *;

-- name: ListFoodItemsByBusiness :many
SELECT * FROM food_items WHERE business_id = @business_id;

-- name: ListAllFoodItems :many
SELECT f.*, b.name as restaurant_name 
FROM food_items f
JOIN businesses b ON f.business_id = b.id
ORDER BY f.created_at DESC;

-- Grocery Items
-- name: CreateGroceryItem :one
INSERT INTO grocery_items (id, business_id, name, description, price, currency, image_url, unit, stock_quantity, category, is_available)
VALUES (@id, @business_id, @name, @description, @price, @currency, @image_url, @unit, @stock_quantity, @category, @is_available)
RETURNING *;

-- name: ListGroceryItems :many
SELECT * FROM grocery_items WHERE is_available = TRUE ORDER BY created_at DESC;

-- name: ListGroceryItemsByBusiness :many
SELECT * FROM grocery_items WHERE business_id = @business_id ORDER BY created_at DESC;

-- Liquor Items
-- name: CreateLiquorItem :one
INSERT INTO liquor_items (id, business_id, name, description, price, currency, image_url, volume, abv, stock_quantity, category, is_available)
VALUES (@id, @business_id, @name, @description, @price, @currency, @image_url, @volume, @abv, @stock_quantity, @category, @is_available)
RETURNING *;

-- name: ListLiquorItems :many
SELECT * FROM liquor_items WHERE is_available = TRUE ORDER BY created_at DESC;

-- name: ListLiquorItemsByBusiness :many
SELECT * FROM liquor_items WHERE business_id = @business_id ORDER BY created_at DESC;

-- Pharmacy Items
-- name: CreatePharmacyItem :one
INSERT INTO pharmacy_items (id, business_id, name, description, price, currency, image_url, requires_prescription, stock_quantity, category, is_available)
VALUES (@id, @business_id, @name, @description, @price, @currency, @image_url, @requires_prescription, @stock_quantity, @category, @is_available)
RETURNING *;

-- name: ListPharmacyItems :many
SELECT * FROM pharmacy_items WHERE is_available = TRUE ORDER BY created_at DESC;

-- name: ListPharmacyItemsByBusiness :many
SELECT * FROM pharmacy_items WHERE business_id = @business_id ORDER BY created_at DESC;

-- Bus Routes
-- name: CreateBusRoute :one
INSERT INTO bus_routes (id, business_id, origin, destination, departure_time, arrival_time, price, currency, available_seats, bus_type)
VALUES (@id, @business_id, @origin, @destination, @departure_time, @arrival_time, @price, @currency, @available_seats, @bus_type)
RETURNING *;

-- name: ListBusRoutes :many
SELECT * FROM bus_routes WHERE departure_time > CURRENT_TIMESTAMP ORDER BY departure_time ASC;

-- Movies
-- name: CreateMovie :one
INSERT INTO movies (id, title, description, genre, rating, duration_minutes, poster_url, is_now_playing, release_date)
VALUES (@id, @title, @description, @genre, @rating, @duration_minutes, @poster_url, @is_now_playing, @release_date)
RETURNING *;

-- name: ListNowPlayingMovies :many
SELECT * FROM movies WHERE is_now_playing = TRUE ORDER BY rating DESC;

-- name: ListComingSoonMovies :many
SELECT * FROM movies WHERE is_now_playing = FALSE AND release_date > CURRENT_DATE ORDER BY release_date ASC;

-- Movie Showtimes
-- name: CreateMovieShowtime :one
INSERT INTO movie_showtimes (id, movie_id, business_id, show_time, price, currency, room_number, available_seats)
VALUES (@id, @movie_id, @business_id, @show_time, @price, @currency, @room_number, @available_seats)
RETURNING *;

-- name: ListMovieShowtimesByMovie :many
SELECT * FROM movie_showtimes WHERE movie_id = @movie_id AND show_time > CURRENT_TIMESTAMP ORDER BY show_time ASC;

-- Flights
-- name: CreateFlight :one
INSERT INTO flights (id, airline_name, flight_number, origin, destination, departure_time, arrival_time, price, currency, class_type, available_seats)
VALUES (@id, @airline_name, @flight_number, @origin, @destination, @departure_time, @arrival_time, @price, @currency, @class_type, @available_seats)
RETURNING *;

-- name: ListFlights :many
SELECT * FROM flights WHERE departure_time > CURRENT_TIMESTAMP ORDER BY departure_time ASC;

-- Jobs
-- name: CreateJob :one
INSERT INTO jobs (id, business_id, title, description, category, job_type, location, salary_range, expires_at)
VALUES (@id, @business_id, @title, @description, @category, @job_type, @location, @salary_range, @expires_at)
RETURNING *;

-- name: ListActiveJobs :many
SELECT * FROM jobs WHERE is_active = TRUE AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP) ORDER BY created_at DESC;

-- Bills
-- name: CreateBill :one
INSERT INTO bills (id, user_id, biller_name, account_number, amount_due, currency, due_date, category)
VALUES (@id, @user_id, @biller_name, @account_number, @amount_due, @currency, @due_date, @category)
RETURNING *;

-- name: ListUserBills :many
SELECT * FROM bills WHERE user_id = @user_id ORDER BY due_date ASC;

-- Wallet
-- name: GetWalletBalance :one
SELECT * FROM user_wallet WHERE user_id = @user_id;

-- name: UpdateWalletBalance :one
UPDATE user_wallet SET balance = @balance, last_updated = CURRENT_TIMESTAMP WHERE user_id = @user_id RETURNING *;

-- Messages
-- name: CreateMessage :one
INSERT INTO messages (id, sender_id, receiver_id, content)
VALUES (@id, @sender_id, @receiver_id, @content)
RETURNING *;

-- name: ListUserMessages :many
SELECT * FROM messages WHERE sender_id = @id OR receiver_id = @id ORDER BY created_at DESC;

-- Notifications
-- name: CreateNotification :one
INSERT INTO notifications (id, user_id, title, body, type)
VALUES (@id, @user_id, @title, @body, @type)
RETURNING *;

-- name: ListUserNotifications :many
SELECT * FROM notifications WHERE user_id = @user_id ORDER BY created_at DESC;

-- name: MarkNotificationAsRead :exec
UPDATE notifications SET is_read = TRUE WHERE id = @id AND user_id = @user_id;

-- Tours
-- name: CreateTour :one
INSERT INTO tours (id, business_id, title, description, location, price, currency, image_url, duration)
VALUES (@id, @business_id, @title, @description, @location, @price, @currency, @image_url, @duration)
RETURNING *;

-- name: ListTours :many
SELECT * FROM tours ORDER BY rating DESC;
