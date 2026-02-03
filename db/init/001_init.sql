-- 1. Таблица пользователей (используется Auth Service)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL, -- Та самая колонка, которой не хватало
    role TEXT NOT NULL DEFAULT 'client' -- client, courier, admin
);

-- 2. Таблица товаров (используется Catalog Service)
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    image_url TEXT
);

-- 3. Таблица курьеров (используется Courier Service)
CREATE TABLE IF NOT EXISTS couriers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    is_available BOOLEAN DEFAULT TRUE,
    current_lat DOUBLE PRECISION,
    current_lon DOUBLE PRECISION
);

-- 4. Таблица заказов (используется Order Service)
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    courier_id INTEGER REFERENCES couriers(id), -- Заполняется после 'accept'
    status TEXT NOT NULL DEFAULT 'new', -- new, accepted, cooking, delivering, completed
    total_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Детали заказа (связывает товары с заказами)
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price_at_purchase DECIMAL(10, 2) NOT NULL -- Цена на момент заказа
);

-- 6. Индексы для ускорения поиска (важно по мере роста базы)
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
