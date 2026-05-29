-- Ecommerce Filtering Schema Extension

-- Attributes Table (e.g., 'Color', 'Size', 'Voltage')
CREATE TABLE attributes (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Attribute Values (e.g., 'Red', 'XL', '240V')
CREATE TABLE attribute_values (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    attribute_id TEXT NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(attribute_id, value)
);

-- Linking Products to Attributes (using JSONB for dynamic filtering)
-- This allows products to store dynamic sets of attributes (e.g., {"color": "red", "size": "XL"})
ALTER TABLE products ADD COLUMN attribute_data JSONB DEFAULT '{}';

-- Create GIN index for fast filtering on attribute data
CREATE INDEX idx_products_attribute_data ON products USING GIN (attribute_data);

-- Helper to make attribute values easily searchable for filter UI
CREATE INDEX idx_products_attribute_data_gin_ops ON products USING GIN (attribute_data jsonb_path_ops);

-- Update RLS for new tables
ALTER TABLE attributes ENABLE ROW LEVEL SECURITY;
ALTER TABLE attribute_values ENABLE ROW LEVEL SECURITY;

CREATE POLICY attributes_view_policy ON attributes FOR SELECT USING (true);
CREATE POLICY attribute_values_view_policy ON attribute_values FOR SELECT USING (true);
