SELECT
    c.id,
    c.name,
    IF(c.organisation_id IS NULL, false, true) AS can_edit,
    COUNT(*) OVER () AS total_count
FROM
    categories AS c
WHERE c.organisation_id IS NULL
   OR c.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)
ORDER BY c.name
LIMIT ?
OFFSET ?
