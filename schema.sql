-- Enable PostGIS extension for geospatial data
CREATE EXTENSION IF NOT EXISTS postgis;

-- Core User Management
CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT,
    phone_number TEXT UNIQUE,
    profile_picture_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User Roles: allows a user to have multiple roles on the platform
CREATE TYPE user_role_type AS ENUM ('customer', 'driver', 'vendor', 'host', 'c2c_seller', 'admin');

CREATE TABLE user_roles (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role user_role_type NOT NULL,
    PRIMARY KEY (user_id, role)
);

-- Addresses (Expanded)
CREATE TABLE addresses (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label TEXT, -- e.a., "Home", "Work", "Shop Address"
    recipient_name TEXT,
    recipient_phone TEXT,
    address_line1 TEXT NOT NULL,
    address_line2 TEXT,
    city TEXT NOT NULL,
    state TEXT,
    postal_code TEXT,
    country TEXT NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    latitude NUMERIC(10, 8), -- Optional: for geo-coding addresses
    longitude NUMERIC(11, 8), -- Optional: for geo-coding addresses
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- General Categories (can be used for products, services, etc.)
CREATE TABLE categories (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    image_url TEXT,
    parent_id TEXT REFERENCES categories(id) ON DELETE SET NULL, -- for subcategories
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Brands Table
CREATE TABLE brands (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    logo_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Generic Product Table (for e-commerce, marketplace, etc.)
CREATE TABLE products (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    category_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
    vendor_id TEXT REFERENCES vendors(id) ON DELETE SET NULL, -- Link product to a vendor (vendors table)
    rating NUMERIC(2, 1) DEFAULT 0.0, -- Average rating
    review_count INTEGER DEFAULT 0,
    -- Flash Sale fields
    is_flash_sale BOOLEAN DEFAULT FALSE,
    discount_percentage NUMERIC(5, 2) DEFAULT 0.0,
    flash_sale_start_time TIMESTAMP WITH TIME ZONE,
    flash_sale_end_time TIMESTAMP WITH TIME ZONE,
    -- New fields for product details
    brand_id TEXT REFERENCES brands(id) ON DELETE SET NULL, -- Link to brands table
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_images (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    is_thumbnail BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_videos (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    video_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Universal Cart Item Table
-- This table is designed to be generic and can link to different types of items
CREATE TABLE cart_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id TEXT NOT NULL, -- ID of the item (product, food_item, service, etc.)
    item_type TEXT NOT NULL, -- Type of item (e.g., 'product', 'food_item', 'service')
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, item_id, item_type) -- A user can only have one entry for a specific item type
);

-- Orders and Order Items (These would typically reference a generic item, but for now, let's link to products for simplicity)
-- NOTE: For a truly universal cart and order system, `order_items` might also need an `item_type` column.
-- For now, we'll keep it linked to `products` assuming the primary e-commerce flow.
CREATE TABLE orders (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    total_amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'pending', -- e.g., pending, processing, shipped, delivered, cancelled
    shipping_address_id TEXT REFERENCES addresses(id) ON DELETE SET NULL, -- Address used for this specific order
    -- payment_method TEXT, -- Could be a transaction ID or reference
    delivery_fee NUMERIC(10, 2) DEFAULT 0.00,
    discount_amount NUMERIC(10, 2) DEFAULT 0.00,
    estimated_delivery_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- eFood: Food Order Specifics
CREATE TABLE food_order_details (
    order_id TEXT PRIMARY KEY REFERENCES orders(id) ON DELETE CASCADE,
    vendor_id TEXT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    driver_id TEXT REFERENCES drivers(id) ON DELETE SET NULL,
    delivery_instructions TEXT,
    preparation_status TEXT NOT NULL DEFAULT 'received', -- received, preparing, ready, picked_up, out_for_delivery, delivered
    estimated_preparation_time TIMESTAMP WITH TIME ZONE,
    actual_preparation_time TIMESTAMP WITH TIME ZONE,
    actual_pickup_time TIMESTAMP WITH TIME ZONE,
    actual_delivery_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id TEXT NOT NULL, -- ID of the item (product, food_item, service, etc.)
    item_type TEXT NOT NULL, -- Type of item (e.g., 'product', 'food_item', 'service')
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(10, 2) NOT NULL, -- Price at the time of order
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- eFood: Order Customization Tracking
CREATE TABLE order_item_options (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    order_item_id TEXT NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    option_id TEXT NOT NULL, -- ID of the food_item_option
    price_at_order NUMERIC(10, 2) NOT NULL, -- Store price at time of order
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Wallet and Payment Transactions
CREATE TABLE transactions (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id TEXT REFERENCES orders(id) ON DELETE SET NULL, -- Link to order if it's a payment for an order
    transaction_type TEXT NOT NULL, -- e.g., "credit", "debit", "top_up", "payment", "refund"
    amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'completed', -- e.g., pending, completed, failed
    payment_gateway_ref TEXT, -- Reference from payment gateway
    description TEXT,
    -- Enhanced transaction fields
    recipient_user_id TEXT REFERENCES users(id) ON DELETE SET NULL, -- For P2P transfers
    metadata JSONB, -- To store additional details like bill information, recipient name, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_wallet (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Role-specific tables linked to Users

-- Hubs (Existing table)
CREATE TABLE hubs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL
);

-- Tasks (Existing table)
CREATE TABLE tasks (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 1
);

-- Drivers table
CREATE TABLE drivers (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(), -- Driver ID, distinct from user ID
    user_id TEXT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL, -- Driver's name
    status TEXT NOT NULL DEFAULT 'offline', -- online, offline, busy
    is_verified BOOLEAN DEFAULT FALSE, -- New: To track driver verification status
    current_trip_id TEXT, -- New: To store the ID of the trip the driver is currently on
    vehicle_make TEXT,
    vehicle_model TEXT,
    license_plate TEXT,
    vehicle_color TEXT,
    vehicle_condition TEXT DEFAULT 'used', -- e.g., 'new', 'good', 'fair', 'poor'
    vehicle_type_id TEXT REFERENCES vehicle_types(id) ON DELETE SET NULL, -- Link to vehicle types
    rating NUMERIC(2, 1) DEFAULT 0.0,
    last_location GEOGRAPHY(POINT, 4326),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_drivers_location ON drivers USING GIST (last_location);

-- Vendors table (for e-commerce sellers, food vendors, grocery stores, etc.)
CREATE TABLE vendors (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shop_name TEXT NOT NULL,
    description TEXT,
    logo_url TEXT,
    address_id TEXT REFERENCES addresses(id) ON DELETE SET NULL, -- Physical address of the vendor
    phone_number TEXT,
    email TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    rating NUMERIC(2, 1) DEFAULT 0.0,
    review_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- eFood: Vendor Operating Hours
CREATE TABLE vendor_operating_hours (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id TEXT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday, 6=Saturday
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    is_closed BOOLEAN DEFAULT FALSE, -- Allows marking a whole day as closed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(vendor_id, day_of_week)
);

-- C2C Sellers table
CREATE TABLE c2c_sellers (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url TEXT,
    reputation_score INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Hosts table (for eHost/rental services)
CREATE TABLE hosts (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    response_rate NUMERIC(3,2) DEFAULT 1.0, -- e.g., 0.95 for 95%
    is_superhost BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Define property types enum
CREATE TYPE property_type_enum AS ENUM ('apartment', 'house', 'condo', 'villa', 'cabin', 'studio', 'townhouse');

-- Properties (for eHost/rental services)
CREATE TABLE properties (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id TEXT NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    address_id TEXT REFERENCES addresses(id) ON DELETE SET NULL, -- Specific property address
    latitude NUMERIC(10, 8),
    longitude NUMERIC(11, 8),
    price_per_night NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    number_of_guests INTEGER DEFAULT 1,
    number_of_bedrooms INTEGER DEFAULT 1,
    number_of_beds INTEGER DEFAULT 1,
    number_of_bathrooms NUMERIC(3,1) DEFAULT 1.0,
    type property_type_enum NOT NULL DEFAULT 'house',
    total_rooms INTEGER,
    has_wifi BOOLEAN DEFAULT FALSE,
    has_kitchen BOOLEAN DEFAULT FALSE,
    has_parking BOOLEAN DEFAULT FALSE,
    has_pool BOOLEAN DEFAULT FALSE,
    has_ac BOOLEAN DEFAULT FALSE,
    image_urls TEXT[], -- Multiple image URLs for the property
    video_url TEXT, -- One video URL for the property
    available_from DATE,
    available_to DATE,
    check_in_time TIME,
    check_out_time TIME,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- eFood: Menu Categories
CREATE TABLE menu_categories (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id TEXT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Food Items (if separate from generic products)
CREATE TABLE food_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id TEXT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    menu_category_id TEXT REFERENCES menu_categories(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    image_url TEXT,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- eFood: Item Customization (Modifiers)
CREATE TABLE food_item_option_groups (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    food_item_id TEXT NOT NULL REFERENCES food_items(id) ON DELETE CASCADE,
    name TEXT NOT NULL, -- e.g., "Choice of Soda", "Add-ons"
    min_selection INTEGER DEFAULT 0, -- 0 for optional, 1 for required
    max_selection INTEGER DEFAULT 1,
    is_required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE food_item_options (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id TEXT NOT NULL REFERENCES food_item_option_groups(id) ON DELETE CASCADE,
    name TEXT NOT NULL, -- e.g., "Coke", "Bacon"
    price NUMERIC(10, 2) DEFAULT 0.00,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE surge_zones (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT,
    boundary GEOMETRY(POLYGON, 4326),
    multiplier NUMERIC(3, 2) DEFAULT 1.0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Taxi Trips (for eTaxi mini-service)
CREATE TABLE taxi_trips (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    driver_id TEXT NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    pickup_location GEOGRAPHY(POINT, 4326) NOT NULL,
    dropoff_location GEOGRAPHY(POINT, 4326) NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    total_amount NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'requested', -- e.g., requested, accepted, in_progress, completed, cancelled
    cancellation_reason TEXT,
    cancelled_by TEXT, -- e.g., 'user', 'driver', 'admin'
    vehicle_type_id TEXT REFERENCES vehicle_types(id) ON DELETE SET NULL, -- Type of vehicle booked
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_taxi_trips_user_id ON taxi_trips(user_id);
CREATE INDEX idx_taxi_trips_driver_id ON taxi_trips(driver_id);
CREATE INDEX idx_taxi_trips_requested_at ON taxi_trips(requested_at);

-- eTaxi: Trip Location Tracking
CREATE TABLE taxi_trip_locations (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id TEXT NOT NULL REFERENCES taxi_trips(id) ON DELETE CASCADE,
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    captured_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_taxi_trip_locations_spatial ON taxi_trip_locations USING GIST (location);

-- Property Bookings (for eHost mini-service)
CREATE TABLE property_bookings (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    property_id TEXT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    total_amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'pending', -- e.g., pending, confirmed, cancelled, completed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_property_bookings_user_id ON property_bookings(user_id);
CREATE INDEX idx_property_bookings_property_id ON property_bookings(property_id);

-- Generic Bookings for other services (Cleaning, Laundry, Repair, Flights, Tours)
CREATE TABLE service_bookings (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_type TEXT NOT NULL, -- e.g., 'cleaning', 'laundry', 'repair', 'flight', 'tour'
    service_item_id TEXT NOT NULL, -- ID of the specific service or item being booked (e.g., repair_service_id, flight_id)
    provider_id TEXT, -- ID of the provider (e.g., user_id of a repair technician, or a vendor_id)
    provider_type TEXT, -- e.g., 'user', 'vendor'
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE, -- Optional for services with duration
    total_amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'pending', -- e.g., pending, confirmed, cancelled, completed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_service_bookings_user_id ON service_bookings(user_id);
CREATE INDEX idx_service_bookings_service_type ON service_bookings(service_type);
CREATE INDEX idx_service_bookings_provider_id ON service_bookings(provider_id);

-- Shipments (for eDelivery mini-service)
CREATE TABLE shipments (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tracking_number TEXT UNIQUE NOT NULL,
    sender_address_id TEXT REFERENCES addresses(id) ON DELETE SET NULL,
    recipient_address_id TEXT REFERENCES addresses(id) ON DELETE SET NULL,
    current_location GEOGRAPHY(POINT, 4326),
    status TEXT NOT NULL DEFAULT 'pending', -- e.g., pending, in_transit, delivered, cancelled
    estimated_delivery_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_shipments_user_id ON shipments(user_id);
CREATE INDEX idx_shipments_tracking_number ON shipments(tracking_number);
CREATE INDEX idx_shipments_status ON shipments(status);

-- Jobs (for eJobs mini-service)
CREATE TABLE companies (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    logo_url TEXT,
    description TEXT,
    website TEXT,
    user_id TEXT UNIQUE REFERENCES users(id) ON DELETE SET NULL, -- Optional: link to a user who manages the company profile
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_companies_name ON companies(name);

CREATE TABLE jobs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    location TEXT,
    salary_min NUMERIC(10, 2),
    salary_max NUMERIC(10, 2),
    currency TEXT DEFAULT 'Ksh',
    job_type TEXT NOT NULL, -- e.g., Full-time, Contract, Part-time
    posted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_jobs_company_id ON jobs(company_id);
CREATE INDEX idx_jobs_title ON jobs(title);
CREATE INDEX idx_jobs_location ON jobs(location);

-- B2B Quotes (for B2B mini-service)
CREATE TABLE b2b_quotes (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id TEXT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_details TEXT NOT NULL, -- Details of what was requested
    status TEXT NOT NULL DEFAULT 'pending', -- e.g., pending, quoted, accepted, rejected
    quoted_amount NUMERIC(10, 2),
    currency TEXT DEFAULT 'Ksh',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_b2b_quotes_vendor_id ON b2b_quotes(vendor_id);
CREATE INDEX idx_b2b_quotes_user_id ON b2b_quotes(user_id);
CREATE INDEX idx_b2b_quotes_status ON b2b_quotes(status);

-- C2C Listings (for C2C mini-service)
CREATE TABLE c2c_listings (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id TEXT NOT NULL REFERENCES c2c_sellers(id) ON DELETE CASCADE,
    category_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
    brand_id TEXT REFERENCES brands(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    is_negotiable BOOLEAN DEFAULT FALSE,
    location TEXT,
    image_urls TEXT[], -- Multiple image URLs for the listing
    condition TEXT NOT NULL DEFAULT 'used', -- New field for product condition
    status TEXT NOT NULL DEFAULT 'available', -- e.g., 'available', 'sold', 'pending', 'inactive'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_c2c_listings_seller_id ON c2c_listings(seller_id);
CREATE INDEX idx_c2c_listings_created_at ON c2c_listings(created_at);

-- Messages (for Social/Inbox functionality)
CREATE TABLE messages (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_receiver_id ON messages(receiver_id);
CREATE INDEX idx_messages_sent_at ON messages(sent_at);

-- Notifications (for alerts/notifications)
CREATE TABLE notifications (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL, -- e.g., 'order_update', 'system', 'promo', 'service_alert'
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

-- Services Table (for bookable services like cleaning, laundry, repair, flights, tours)
CREATE TABLE services (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- The user providing the service
    service_type TEXT NOT NULL, -- e.g., 'cleaning', 'laundry', 'repair', 'flight', 'tour', 'hotel'
    name TEXT NOT NULL,
    description TEXT,
    base_price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    location GEOGRAPHY(POINT, 4326), -- For services like cleaning, repair, taxi, hotel locations
    availability_details TEXT, -- e.g., opening hours, service area, flight schedule, hotel availability
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_services_provider_user_id ON services(provider_user_id);
CREATE INDEX idx_services_service_type ON services(service_type);

-- Vendor Miniservice Verification
CREATE TABLE vendor_miniservice_verification (
    vendor_id TEXT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    miniservice_type TEXT NOT NULL, -- e.g., 'liquor', 'pharmacy', 'taxi', 'flights'
    verification_status TEXT NOT NULL DEFAULT 'pending', -- e.g., 'pending', 'verified', 'rejected'
    verified_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL, -- Admin user who verified
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (vendor_id, miniservice_type) -- A vendor is verified once per miniservice type
);
CREATE INDEX idx_vendor_miniservice_verification_vendor_id ON vendor_miniservice_verification(vendor_id);
CREATE INDEX idx_vendor_miniservice_verification_miniservice_type ON vendor_miniservice_verification(miniservice_type);


-- Product Variants Table
CREATE TABLE product_variants (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku TEXT UNIQUE, -- Stock Keeping Unit
    price NUMERIC(10, 2) NOT NULL, -- Price for this specific variant
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    attributes JSONB, -- e.g., {"color": "Red", "storage": "128GB"}
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);

-- Product Discounts Table
CREATE TABLE product_discounts (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    discount_type TEXT NOT NULL, -- e.g., 'percentage', 'fixed_amount'
    discount_value NUMERIC(11, 2) NOT NULL, -- e.g., 10 for 10%, or 5.00 for $5 off. Increased precision for value.
    start_at TIMESTAMP WITH TIME ZONE,
    end_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_product_discounts_product_id ON product_discounts(product_id);
CREATE INDEX idx_product_discounts_end_at ON product_discounts(end_at);

-- Vehicle Types for eTaxi
CREATE TABLE vehicle_types (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL, -- e.g., 'Small Car', 'Large Car', 'Motorbike', 'Normal Car'
    passenger_capacity INTEGER NOT NULL,
    base_price_per_km NUMERIC(10, 2) NOT NULL DEFAULT 0.00, -- Example pricing structure
    base_price_per_minute NUMERIC(10, 2) NOT NULL DEFAULT 0.00, -- Example pricing structure
    initial_fee NUMERIC(10, 2) NOT NULL DEFAULT 0.00, -- Example pricing structure
    currency TEXT NOT NULL DEFAULT 'Ksh',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- Driver Ratings Table
CREATE TABLE driver_ratings (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id TEXT NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- The user who rated
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    testimonial TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_driver_ratings_driver_id ON driver_ratings(driver_id);
CREATE INDEX idx_driver_ratings_user_id ON driver_ratings(user_id);