SELECT
    o.id,
    o.name,
    o.member_count,
    u2o.role
FROM
    go_users_2_organisations AS u2o
    INNER JOIN go_organisations o ON o.id = u2o.organisation_id
WHERE
    u2o.organisation_id = ?
    AND u2o.user_id = ?
