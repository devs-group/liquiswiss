UPDATE oauth_auth_codes
SET used_at = CURRENT_TIMESTAMP
WHERE code_hash = ? AND used_at IS NULL;
