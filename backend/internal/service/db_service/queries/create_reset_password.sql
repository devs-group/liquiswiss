INSERT INTO reset_password (email, code, created_at)
SELECT ?, ?, NOW()
WHERE EXISTS (
    SELECT 1
    FROM users
    WHERE email = ?
)
AND NOT EXISTS (
    SELECT 1
    FROM reset_password
    WHERE email = ?
      AND TIMESTAMPDIFF(MINUTE, created_at, NOW()) <= ?
)
ON DUPLICATE KEY UPDATE
    code = VALUES(code),
    created_at = VALUES(created_at);