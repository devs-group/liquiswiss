SELECT EXISTS(
    SELECT 1 FROM go_refresh_tokens
    WHERE token_id = ? AND user_id = ? AND expires_at > NOW()
)