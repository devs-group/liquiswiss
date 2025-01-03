SELECT
    COUNT(*) OVER () AS total_count
FROM
    employees AS e
WHERE
    e.organisation_id = get_current_organisation(?)
LIMIT ?
OFFSET ?
