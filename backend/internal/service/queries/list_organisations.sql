SELECT
    o.id,
    o.name,
    o.member_count,
    u2o.role,
    u2o.is_default,
    COUNT(*) OVER () AS total_count
FROM
    users_2_organisations AS u2o
    INNER JOIN organisations o ON o.id = u2o.organisation_id
WHERE
    u2o.user_id = ?
LIMIT ?
OFFSET ?