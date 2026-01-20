SELECT
    u.id AS user_id,
    u.name,
    u.email,
    u2o.role,
    u2o.is_default
FROM users_2_organisations u2o
INNER JOIN users u ON u.id = u2o.user_id
WHERE u2o.organisation_id = ?
ORDER BY u2o.role = 'owner' DESC, u.name ASC