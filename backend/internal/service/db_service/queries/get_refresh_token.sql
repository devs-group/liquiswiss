SELECT EXISTS(
    SELECT 1 FROM refresh_tokens
    WHERE token_id = ? AND user_id = ? AND expires_at > NOW()
)