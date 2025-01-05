SELECT
    c.id,
    c.name,
    IF(c.organisation_id IS NULL, false, true) AS can_edit,
    COUNT(*) OVER () AS total_count
FROM
    categories AS c
WHERE c.organisation_id IS NULL
   OR c.organisation_id = get_current_organisation(?)
ORDER BY c.name
LIMIT ?
OFFSET ?
