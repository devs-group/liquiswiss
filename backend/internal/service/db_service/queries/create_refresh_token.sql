INSERT INTO refresh_tokens (user_id, token_id, expires_at, device_name)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE token_id = VALUES(token_id), expires_at = VALUES(expires_at), device_name = VALUES(device_name)