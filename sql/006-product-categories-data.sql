-- Insert categories
INSERT INTO categories (code, name) VALUES
    ('CLOTHING', 'Clothing'),
    ('SHOES', 'Shoes'),
    ('ACCESSORIES', 'Accessories')
ON CONFLICT (code) DO NOTHING;

-- PROD001, PROD004, PROD007 belongs to "Clothing"
UPDATE products SET category_id = (SELECT id FROM categories WHERE code = 'CLOTHING') WHERE code IN ('PROD001', 'PROD004', 'PROD007');

-- PROD002, PROD006 belongs to "Shoes"
UPDATE products SET category_id = (SELECT id FROM categories WHERE code = 'SHOES') WHERE code IN ('PROD002', 'PROD006');

-- PROD003, PROD005, PROD008 belongs to "Accessories"
UPDATE products SET category_id = (SELECT id FROM categories WHERE code = 'ACCESSORIES') WHERE code IN ('PROD003', 'PROD005', 'PROD008');

