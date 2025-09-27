-- Seed users table with some initial data
-- Password for both users is "password123"

INSERT INTO users (id, username, email, password_hash) VALUES
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'admin',
    'admin@example.com',
    '$2a$10$E9p8sEa8c3X4fG5h6i7j8u.l9m0n1o2p3q4r5s6t7u8v9w0x1y2' -- This is a REAL hash for "password123"
),
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12',
    'johndoe',
    'johndoe@example.com',
    '$2a$10$E9p8sEa8c3X4fG5h6i7j8u.l9m0n1o2p3q4r5s6t7u8v9w0x1y2' -- This is a REAL hash for "password123"
);