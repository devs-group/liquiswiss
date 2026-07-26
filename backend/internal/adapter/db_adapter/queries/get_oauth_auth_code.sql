SELECT code_hash, client_id, user_id, code_challenge, redirect_uri, COALESCE(resource, ''), expires_at, used_at
FROM oauth_auth_codes
WHERE code_hash = ?;
