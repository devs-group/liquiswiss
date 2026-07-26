SELECT EXISTS (
    SELECT 1
    FROM oauth_refresh_tokens
    WHERE user_id = ? AND client_id = ? AND revoked_at IS NULL AND expires_at > NOW()
)
