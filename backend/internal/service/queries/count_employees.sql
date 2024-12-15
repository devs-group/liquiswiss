SELECT
    COUNT(*) OVER () AS total_count
FROM
    employees AS e
WHERE
    e.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)
LIMIT ?
OFFSET ?
