UPDATE oauth_refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND client_id = ? AND revoked_at IS NULL;
