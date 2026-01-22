-- Stage 0 schema bootstrap (MVP). Later stages will add constraints and reference data.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Users: both customers and couriers. Role: user|courier|admin
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  role text NOT NULL CHECK (role IN ('user','courier','admin')),
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Courier profile/location data
CREATE TABLE IF NOT EXISTS couriers (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  is_available boolean NOT NULL DEFAULT true,
  current_lat double precision,
  current_lng double precision,
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_couriers_availability ON couriers (is_available);

-- Product catalog
CREATE TABLE IF NOT EXISTS products (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  price_cents integer NOT NULL CHECK (price_cents >= 0),
  image_url text,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_products_active ON products (is_active);

-- Orders
CREATE TABLE IF NOT EXISTS orders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id),
  courier_id uuid REFERENCES users(id),
  status text NOT NULL CHECK (status IN ('created','accepted','on_the_way','delivered','paid')),
  total_cents integer NOT NULL CHECK (total_cents >= 0),
  delivery_address text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_courier_id_status ON orders(courier_id, status);

-- Order items
CREATE TABLE IF NOT EXISTS order_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id uuid NOT NULL REFERENCES products(id),
  quantity integer NOT NULL CHECK (quantity > 0),
  options jsonb NOT NULL DEFAULT '{}'::jsonb,
  price_cents integer NOT NULL CHECK (price_cents >= 0)
);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- Order status history (audit log)
CREATE TABLE IF NOT EXISTS order_status_history (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  status text NOT NULL,
  changed_by uuid,
  changed_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history(order_id);
