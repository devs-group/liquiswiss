SELECT
    o.id,
    o.name,
    member_counts.member_count AS member_count,
    u2o.role
FROM users_2_organisations AS u2o
INNER JOIN organisations o ON o.id = u2o.organisation_id
LEFT JOIN (
    SELECT organisation_id, COUNT(*) AS member_count
    FROM users_2_organisations
    GROUP BY organisation_id
) AS member_counts ON member_counts.organisation_id = o.id
WHERE
    u2o.organisation_id = ?
    AND u2o.user_id = ?
