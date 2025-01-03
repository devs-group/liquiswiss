INSERT INTO registrations (email, code)
SELECT ?, ?
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE LOWER(email) = LOWER(?)
);