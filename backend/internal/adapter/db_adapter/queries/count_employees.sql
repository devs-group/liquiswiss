SELECT
    COUNT(*) OVER () AS total_count
FROM
    employees AS e
WHERE
    e.organisation_id = get_current_user_organisation_id(?)
LIMIT ?
OFFSET ?
