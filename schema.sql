-- Enable PostGIS extension for geospatial data
CREATE EXTENSION IF NOT EXISTS postgis;

-- User Roles Enum
CREATE TYPE user_role_type AS ENUM (
    'executive_admin', 
    'admin', 
    'staff', 
    'vendor', 
    'customer', 
    'driver', 
    'host', 
    'c2c_seller'
);

-- Business Verification Status
CREATE TYPE business_verification_status AS ENUM (
    'pending', 
    'document_verified', 
    'approved', 
    'rejected', 
    'suspended'
);

-- Document Type for Verification
CREATE TYPE document_type AS ENUM (
    'id_proof', 
    'business_license', 
    'tax_certificate', 
    'permit', 
    'other'
);

-- Property types Enum
CREATE TYPE property_type_enum AS ENUM (
    'apartment', 'house', 'condo', 'villa', 'cabin', 'studio', 'townhouse'
);

-- RLS Helper Functions
CREATE OR REPLACE FUNCTION get_app_user_id() RETURNS TEXT AS $$
    SELECT current_setting('app.current_user_id', true);
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION get_app_user_roles() RETURNS TEXT AS $$
    SELECT current_setting('app.current_user_roles', true);
$$ LANGUAGE sql STABLE;

-- Check if user has a specific role
CREATE OR REPLACE FUNCTION has_role(required_role TEXT) RETURNS BOOLEAN AS $$
    SELECT get_app_user_roles() LIKE '%' || required_role || '%';
$$ LANGUAGE sql STABLE;

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

CREATE TABLE user_roles (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role user_role_type NOT NULL,
    PRIMARY KEY (user_id, role)
);

-- Addresses
CREATE TABLE addresses (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label TEXT, 
    recipient_name TEXT,
    recipient_phone TEXT,
    address_line1 TEXT NOT NULL,
    address_line2 TEXT,
    city TEXT NOT NULL,
    state TEXT,
    postal_code TEXT,
    country TEXT NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    latitude NUMERIC(10, 8),
    longitude NUMERIC(11, 8),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- General Categories
CREATE TABLE categories (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    image_url TEXT,
    parent_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
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

-- Businesses (The "Stalls" in the Mall)
CREATE TABLE businesses (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    logo_url TEXT,
    banner_url TEXT,
    miniservice_type TEXT NOT NULL, -- 'liquor', 'hotel', 'laundry', 'grocery', etc.
    address_id TEXT REFERENCES addresses(id) ON DELETE SET NULL,
    phone_number TEXT,
    email TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    verification_status business_verification_status DEFAULT 'pending',
    rating NUMERIC(2, 1) DEFAULT 0.0,
    review_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Business Locations (Multiple branches/warehouses)
CREATE TABLE business_locations (
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    address_id TEXT NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
    PRIMARY KEY (business_id, address_id)
);

-- Vendor Documents for Verification (Linked to Business)
CREATE TABLE vendor_documents (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    type document_type NOT NULL,
    url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    review_notes TEXT,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Business Verification Logs
CREATE TABLE business_verification_logs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    actor_id TEXT NOT NULL REFERENCES users(id),
    old_status business_verification_status,
    new_status business_verification_status NOT NULL,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Products (Linked to Business)
CREATE TABLE products (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    category_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
    brand_id TEXT REFERENCES brands(id) ON DELETE SET NULL,
    image_urls TEXT[],
    rating NUMERIC(2, 1) DEFAULT 0.0,
    review_count INTEGER DEFAULT 0,
    is_flash_sale BOOLEAN DEFAULT FALSE,
    discount_percentage NUMERIC(5, 2) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Product Variants
CREATE TABLE product_variants (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku TEXT UNIQUE,
    price NUMERIC(10, 2) NOT NULL,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    attributes JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Product Discounts
CREATE TABLE product_discounts (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    discount_type TEXT NOT NULL,
    discount_value NUMERIC(11, 2) NOT NULL,
    start_at TIMESTAMP WITH TIME ZONE,
    end_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Properties (Stays/Hotels - Linked to Business)
CREATE TABLE properties (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    address_id TEXT REFERENCES addresses(id) ON DELETE SET NULL,
    price_per_night NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    number_of_guests INTEGER DEFAULT 1,
    number_of_bedrooms INTEGER DEFAULT 1,
    type property_type_enum NOT NULL DEFAULT 'house',
    image_urls TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Property Bookings
CREATE TABLE property_bookings (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    property_id TEXT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    total_amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Food Items (Linked to Business)
CREATE TABLE food_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    image_url TEXT,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Grocery Items (eGrocery - Linked to Business)
CREATE TABLE grocery_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    image_url TEXT,
    unit TEXT, -- 'kg', 'bundle', 'pcs', etc.
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    category TEXT,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Liquor Items (eLiquor - Linked to Business)
CREATE TABLE liquor_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    image_url TEXT,
    volume TEXT, -- '750ml', '1L', etc.
    abv NUMERIC(4, 2), -- Alcohol by volume
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    category TEXT,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Pharmacy Items (eHealth - Linked to Business)
CREATE TABLE pharmacy_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    image_url TEXT,
    requires_prescription BOOLEAN DEFAULT FALSE,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    category TEXT,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Wholesale Items (B2B - Linked to Business)
CREATE TABLE wholesale_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    image_urls TEXT[],
    unit_price NUMERIC(10, 2),
    bulk_price NUMERIC(10, 2),
    bulk_quantity INTEGER,
    category TEXT,
    is_available BOOLEAN DEFAULT TRUE,
    rating NUMERIC(2, 1) DEFAULT 0.0,
    review_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Bus Routes (eBus)
CREATE TABLE bus_routes (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    origin TEXT NOT NULL,
    destination TEXT NOT NULL,
    departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrival_time TIMESTAMP WITH TIME ZONE,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    available_seats INTEGER NOT NULL DEFAULT 0,
    bus_type TEXT, -- 'Executive', 'Economy', etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Movies (eCinema)
CREATE TABLE movies (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    genre TEXT,
    rating NUMERIC(2, 1),
    duration_minutes INTEGER,
    poster_url TEXT,
    is_now_playing BOOLEAN DEFAULT FALSE,
    release_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Movie Showtimes
CREATE TABLE movie_showtimes (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    show_time TIMESTAMP WITH TIME ZONE NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    room_number TEXT,
    available_seats INTEGER NOT NULL DEFAULT 0
);

-- Flights (eFlights)
CREATE TABLE flights (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    airline_name TEXT NOT NULL,
    flight_number TEXT NOT NULL,
    origin TEXT NOT NULL,
    destination TEXT NOT NULL,
    departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    class_type TEXT, -- 'Economy', 'Business', 'First'
    available_seats INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tours (eTravel)
CREATE TABLE tours (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    location TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    rating NUMERIC(2, 1) DEFAULT 0.0,
    image_url TEXT,
    duration TEXT, -- '3 Days', '1 Week', etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Jobs (eJobs)
CREATE TABLE jobs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT,
    job_type TEXT, -- 'Full-time', 'Contract', 'Part-time'
    location TEXT,
    salary_range TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Bills (eBills)
CREATE TABLE bills (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    biller_name TEXT NOT NULL,
    account_number TEXT NOT NULL,
    amount_due NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    due_date DATE,
    status TEXT DEFAULT 'unpaid', -- 'unpaid', 'paid', 'overdue'
    category TEXT, -- 'Electricity', 'Water', etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Messaging
CREATE TABLE messages (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Notifications
CREATE TABLE notifications (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    type TEXT, -- 'order', 'promo', 'system'
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Services (Laundry, Repair, etc. - Linked to Business)
CREATE TABLE services (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    service_type TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    base_price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    location GEOGRAPHY(POINT, 4326),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Service Bookings
CREATE TABLE service_bookings (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_type TEXT NOT NULL,
    service_item_id TEXT NOT NULL,
    provider_id TEXT,
    provider_type TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    total_amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Vehicle Types
CREATE TABLE vehicle_types (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    passenger_capacity INTEGER NOT NULL,
    base_price_per_km NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    initial_fee NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Drivers
CREATE TABLE drivers (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'offline',
    vehicle_type_id TEXT REFERENCES vehicle_types(id) ON DELETE SET NULL,
    rating NUMERIC(2, 1) DEFAULT 0.0,
    last_location GEOGRAPHY(POINT, 4326),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Hubs & Tasks (Global)
CREATE TABLE hubs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE tasks (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 1
);

-- Orders
CREATE TABLE orders (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'pending',
    shipping_address_id TEXT REFERENCES addresses(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Order Items
CREATE TABLE order_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    business_id TEXT NOT NULL REFERENCES businesses(id),
    item_id TEXT NOT NULL,
    item_type TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Cart Items
CREATE TABLE cart_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    business_id TEXT NOT NULL REFERENCES businesses(id),
    item_id TEXT NOT NULL,
    item_type TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, item_id, item_type)
);

-- Taxi Trips
CREATE TABLE taxi_trips (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    driver_id TEXT REFERENCES drivers(id) ON DELETE SET NULL,
    pickup_location GEOGRAPHY(POINT, 4326) NOT NULL,
    dropoff_location GEOGRAPHY(POINT, 4326) NOT NULL,
    total_amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'requested',
    accepted_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- C2C Marketplace
CREATE TABLE c2c_sellers (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE c2c_listings (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id TEXT NOT NULL REFERENCES c2c_sellers(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    image_urls TEXT[],
    is_negotiable BOOLEAN DEFAULT FALSE,
    location TEXT,
    condition TEXT DEFAULT 'used',
    status TEXT NOT NULL DEFAULT 'available',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Transactions & Wallets
CREATE TABLE user_wallet (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id TEXT REFERENCES orders(id) ON DELETE SET NULL,
    amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    status TEXT NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Review System
CREATE TABLE reviews (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    target_id TEXT NOT NULL, -- can be business_id, driver_id, etc.
    target_type TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id),
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Search Index
CREATE INDEX idx_products_search ON products USING GIN (to_tsvector('english', name || ' ' || description));

-- ==========================================
-- ROW LEVEL SECURITY (RLS) POLICIES
-- ==========================================

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE businesses ENABLE ROW LEVEL SECURITY;
ALTER TABLE business_locations ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE food_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE grocery_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE liquor_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE pharmacy_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE wholesale_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE bus_routes ENABLE ROW LEVEL SECURITY;
ALTER TABLE movies ENABLE ROW LEVEL SECURITY;
ALTER TABLE movie_showtimes ENABLE ROW LEVEL SECURITY;
ALTER TABLE flights ENABLE ROW LEVEL SECURITY;
ALTER TABLE jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE bills ENABLE ROW LEVEL SECURITY;
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE services ENABLE ROW LEVEL SECURITY;
ALTER TABLE properties ENABLE ROW LEVEL SECURITY;
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE vendor_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE c2c_listings ENABLE ROW LEVEL SECURITY;
ALTER TABLE taxi_trips ENABLE ROW LEVEL SECURITY;
ALTER TABLE service_bookings ENABLE ROW LEVEL SECURITY;
ALTER TABLE property_bookings ENABLE ROW LEVEL SECURITY;

ALTER TABLE users FORCE ROW LEVEL SECURITY;
ALTER TABLE businesses FORCE ROW LEVEL SECURITY;
ALTER TABLE business_locations FORCE ROW LEVEL SECURITY;
ALTER TABLE products FORCE ROW LEVEL SECURITY;
ALTER TABLE food_items FORCE ROW LEVEL SECURITY;
ALTER TABLE grocery_items FORCE ROW LEVEL SECURITY;
ALTER TABLE liquor_items FORCE ROW LEVEL SECURITY;
ALTER TABLE pharmacy_items FORCE ROW LEVEL SECURITY;
ALTER TABLE wholesale_items FORCE ROW LEVEL SECURITY;
ALTER TABLE bus_routes FORCE ROW LEVEL SECURITY;
ALTER TABLE movies FORCE ROW LEVEL SECURITY;
ALTER TABLE movie_showtimes FORCE ROW LEVEL SECURITY;
ALTER TABLE flights FORCE ROW LEVEL SECURITY;
ALTER TABLE jobs FORCE ROW LEVEL SECURITY;
ALTER TABLE bills FORCE ROW LEVEL SECURITY;
ALTER TABLE messages FORCE ROW LEVEL SECURITY;
ALTER TABLE notifications FORCE ROW LEVEL SECURITY;
ALTER TABLE services FORCE ROW LEVEL SECURITY;
ALTER TABLE properties FORCE ROW LEVEL SECURITY;
ALTER TABLE orders FORCE ROW LEVEL SECURITY;
ALTER TABLE vendor_documents FORCE ROW LEVEL SECURITY;
ALTER TABLE c2c_listings FORCE ROW LEVEL SECURITY;
ALTER TABLE taxi_trips FORCE ROW LEVEL SECURITY;
ALTER TABLE service_bookings FORCE ROW LEVEL SECURITY;
ALTER TABLE property_bookings FORCE ROW LEVEL SECURITY;

-- 1. USERS Policy
CREATE POLICY users_access_policy ON users
    USING (id = get_app_user_id() OR has_role('staff') OR has_role('admin'));

-- 2. BUSINESSES Policy
CREATE POLICY businesses_view_policy ON businesses
    FOR SELECT USING (verification_status = 'approved' OR owner_id = get_app_user_id() OR has_role('staff'));

CREATE POLICY businesses_manage_policy ON businesses
    FOR ALL USING (owner_id = get_app_user_id() OR has_role('admin'));

CREATE POLICY business_locations_policy ON business_locations
    FOR ALL USING (business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff') OR has_role('admin'));

-- 3. PRODUCTS/SERVICES/PROPERTIES Policy
CREATE POLICY products_manage_policy ON products
    FOR ALL USING (
        business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
    );

CREATE POLICY grocery_manage_policy ON grocery_items
    FOR ALL USING (
        business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
    );

CREATE POLICY liquor_manage_policy ON liquor_items
    FOR ALL USING (
        business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
    );

CREATE POLICY pharmacy_manage_policy ON pharmacy_items
    FOR ALL USING (
        business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
    );

CREATE POLICY wholesale_manage_policy ON wholesale_items
    FOR ALL USING (
        business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
    );

CREATE POLICY services_manage_policy ON services
    FOR ALL USING (
        business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
    );

CREATE POLICY properties_manage_policy ON properties
    FOR ALL USING (
        business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
    );

CREATE POLICY tours_manage_policy ON tours
    FOR ALL USING (
        business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
    );

CREATE POLICY tours_view_policy ON tours
    FOR SELECT USING (true);

-- 4. VENDOR DOCUMENTS Policy
CREATE POLICY vendor_docs_access_policy ON vendor_documents
    USING (
        business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
    );

-- 5. ORDERS Policy
CREATE POLICY orders_access_policy ON orders
    USING (
        user_id = get_app_user_id() OR 
        id IN (SELECT order_id FROM order_items WHERE business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id())) OR
        has_role('staff')
    );

-- 6. TAXI TRIPS Policy
CREATE POLICY taxi_trips_access_policy ON taxi_trips
    USING (
        user_id = get_app_user_id() OR 
        driver_id IN (SELECT id FROM drivers WHERE user_id = get_app_user_id()) OR
        has_role('staff')
    );

-- 7. SERVICE/PROPERTY BOOKINGS Policy
CREATE POLICY service_bookings_policy ON service_bookings
    USING (
        user_id = get_app_user_id() OR 
        provider_id = get_app_user_id() OR
        has_role('staff')
    );

CREATE POLICY property_bookings_policy ON property_bookings
    USING (
        user_id = get_app_user_id() OR 
        property_id IN (SELECT id FROM properties WHERE business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id())) OR
        has_role('staff')
    );

-- 8. Specialized Services Policy
CREATE POLICY bus_routes_view_policy ON bus_routes FOR SELECT USING (true);
CREATE POLICY bus_routes_manage_policy ON bus_routes FOR ALL USING (has_role('admin') OR has_role('staff'));

CREATE POLICY movies_view_policy ON movies FOR SELECT USING (true);
CREATE POLICY movie_showtimes_view_policy ON movie_showtimes FOR SELECT USING (true);

CREATE POLICY flights_view_policy ON flights FOR SELECT USING (true);

CREATE POLICY jobs_view_policy ON jobs FOR SELECT USING (is_active = true OR has_role('staff'));
CREATE POLICY jobs_manage_policy ON jobs FOR ALL USING (
    business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')
);

CREATE POLICY bills_access_policy ON bills USING (user_id = get_app_user_id() OR has_role('staff'));

CREATE POLICY messages_access_policy ON messages USING (sender_id = get_app_user_id() OR receiver_id = get_app_user_id());

CREATE POLICY notifications_access_policy ON notifications USING (user_id = get_app_user_id());

-- B2B / Wholesale Extensions
CREATE TABLE rfqs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'responded', 'converted', 'closed'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE rfq_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    rfq_id TEXT NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
    product_id TEXT REFERENCES products(id) ON DELETE SET NULL,
    item_name TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit TEXT,
    target_price NUMERIC(10, 2)
);

CREATE TABLE b2b_quotes (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    rfq_id TEXT NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
    business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    total_amount NUMERIC(10, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'Ksh',
    valid_until TIMESTAMP WITH TIME ZONE,
    status TEXT NOT NULL DEFAULT 'offered', -- 'offered', 'accepted', 'declined'
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE rfqs ENABLE ROW LEVEL SECURITY;
ALTER TABLE rfq_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE b2b_quotes ENABLE ROW LEVEL SECURITY;

ALTER TABLE rfqs FORCE ROW LEVEL SECURITY;
ALTER TABLE rfq_items FORCE ROW LEVEL SECURITY;
ALTER TABLE b2b_quotes FORCE ROW LEVEL SECURITY;

CREATE POLICY rfqs_access_policy ON rfqs
    USING (buyer_id = get_app_user_id() OR business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff'));

CREATE POLICY rfq_items_access_policy ON rfq_items
    USING (rfq_id IN (SELECT id FROM rfqs WHERE buyer_id = get_app_user_id() OR business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR has_role('staff')));

CREATE POLICY b2b_quotes_access_policy ON b2b_quotes
    USING (business_id IN (SELECT id FROM businesses WHERE owner_id = get_app_user_id()) OR rfq_id IN (SELECT id FROM rfqs WHERE buyer_id = get_app_user_id()) OR has_role('staff'));
