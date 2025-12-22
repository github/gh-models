-- BrainSait SQL Analyst - Sample Analytics Database

-- Products table
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    price DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Customers table
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    country VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders table
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customers(id),
    total_amount DECIMAL(10,2),
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order items table
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id),
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER,
    unit_price DECIMAL(10,2)
);

-- Sample data: Products
INSERT INTO products (name, category, price) VALUES
    ('Laptop Pro 15', 'Electronics', 1299.99),
    ('Wireless Mouse', 'Electronics', 49.99),
    ('USB-C Hub', 'Electronics', 79.99),
    ('Standing Desk', 'Furniture', 599.99),
    ('Ergonomic Chair', 'Furniture', 449.99),
    ('Monitor 27"', 'Electronics', 399.99),
    ('Keyboard Mechanical', 'Electronics', 149.99),
    ('Desk Lamp', 'Furniture', 89.99),
    ('Webcam HD', 'Electronics', 129.99),
    ('Headphones', 'Electronics', 199.99);

-- Sample data: Customers
INSERT INTO customers (name, email, country) VALUES
    ('John Smith', 'john@example.com', 'USA'),
    ('Maria Garcia', 'maria@example.com', 'Spain'),
    ('Ahmed Hassan', 'ahmed@example.com', 'Egypt'),
    ('Yuki Tanaka', 'yuki@example.com', 'Japan'),
    ('Emma Wilson', 'emma@example.com', 'UK'),
    ('Mohammed Ali', 'mo@example.com', 'UAE'),
    ('Lisa Chen', 'lisa@example.com', 'China'),
    ('Hans Mueller', 'hans@example.com', 'Germany'),
    ('Sarah Johnson', 'sarah@example.com', 'USA'),
    ('Carlos Rodriguez', 'carlos@example.com', 'Mexico');

-- Sample data: Orders (last 3 months)
INSERT INTO orders (customer_id, total_amount, status, created_at) VALUES
    (1, 1349.98, 'completed', NOW() - INTERVAL '5 days'),
    (2, 599.99, 'completed', NOW() - INTERVAL '10 days'),
    (3, 249.98, 'completed', NOW() - INTERVAL '15 days'),
    (4, 1749.97, 'completed', NOW() - INTERVAL '20 days'),
    (5, 449.99, 'completed', NOW() - INTERVAL '25 days'),
    (6, 879.98, 'completed', NOW() - INTERVAL '30 days'),
    (7, 1299.99, 'completed', NOW() - INTERVAL '35 days'),
    (8, 329.98, 'shipped', NOW() - INTERVAL '3 days'),
    (9, 649.98, 'shipped', NOW() - INTERVAL '2 days'),
    (10, 199.99, 'pending', NOW() - INTERVAL '1 day'),
    (1, 449.99, 'completed', NOW() - INTERVAL '45 days'),
    (3, 1299.99, 'completed', NOW() - INTERVAL '50 days'),
    (5, 279.98, 'completed', NOW() - INTERVAL '55 days'),
    (7, 599.99, 'completed', NOW() - INTERVAL '60 days'),
    (9, 849.98, 'completed', NOW() - INTERVAL '65 days');

-- Sample data: Order items
INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
    (1, 1, 1, 1299.99),
    (1, 2, 1, 49.99),
    (2, 4, 1, 599.99),
    (3, 2, 2, 49.99),
    (3, 7, 1, 149.99),
    (4, 1, 1, 1299.99),
    (4, 5, 1, 449.99),
    (5, 5, 1, 449.99),
    (6, 6, 1, 399.99),
    (6, 3, 1, 79.99),
    (6, 10, 2, 199.99),
    (7, 1, 1, 1299.99),
    (8, 7, 1, 149.99),
    (8, 9, 1, 129.99),
    (8, 2, 1, 49.99),
    (9, 4, 1, 599.99),
    (9, 2, 1, 49.99),
    (10, 10, 1, 199.99),
    (11, 5, 1, 449.99),
    (12, 1, 1, 1299.99),
    (13, 3, 2, 79.99),
    (13, 8, 1, 89.99),
    (13, 2, 1, 49.99),
    (14, 4, 1, 599.99),
    (15, 6, 1, 399.99),
    (15, 5, 1, 449.99);

-- Create useful views
CREATE VIEW monthly_sales AS
SELECT 
    DATE_TRUNC('month', o.created_at) as month,
    COUNT(DISTINCT o.id) as order_count,
    SUM(o.total_amount) as total_sales,
    COUNT(DISTINCT o.customer_id) as unique_customers
FROM orders o
WHERE o.status IN ('completed', 'shipped')
GROUP BY DATE_TRUNC('month', o.created_at)
ORDER BY month DESC;

CREATE VIEW product_performance AS
SELECT 
    p.id,
    p.name,
    p.category,
    COUNT(oi.id) as times_ordered,
    SUM(oi.quantity) as total_quantity,
    SUM(oi.quantity * oi.unit_price) as total_revenue
FROM products p
LEFT JOIN order_items oi ON p.id = oi.product_id
GROUP BY p.id, p.name, p.category
ORDER BY total_revenue DESC;

CREATE VIEW customer_segments AS
SELECT 
    c.id,
    c.name,
    c.country,
    COUNT(o.id) as order_count,
    SUM(o.total_amount) as total_spent,
    MAX(o.created_at) as last_order
FROM customers c
LEFT JOIN orders o ON c.id = o.customer_id
GROUP BY c.id, c.name, c.country
ORDER BY total_spent DESC;
