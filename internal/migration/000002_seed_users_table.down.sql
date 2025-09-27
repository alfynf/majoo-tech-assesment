-- Remove the seeded users
DELETE FROM users WHERE email IN ('admin@example.com', 'johndoe@example.com');