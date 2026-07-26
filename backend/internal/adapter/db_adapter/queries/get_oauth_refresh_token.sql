SELECT token_hash, client_id, user_id, expires_at, revoked_at, rotated_from
FROM oauth_refresh_tokens
WHERE token_hash = ?;
