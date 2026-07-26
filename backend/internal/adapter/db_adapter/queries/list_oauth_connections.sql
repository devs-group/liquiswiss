SELECT rt.client_id, c.client_name, MIN(rt.created_at) AS created_at, MAX(rt.created_at) AS last_used_at
FROM oauth_refresh_tokens rt
JOIN oauth_clients c ON c.client_id = rt.client_id
WHERE rt.user_id = ? AND rt.revoked_at IS NULL AND rt.expires_at > CURRENT_TIMESTAMP
GROUP BY rt.client_id, c.client_name
ORDER BY last_used_at DESC;
