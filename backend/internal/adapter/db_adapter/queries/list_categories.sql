SELECT
    c.id,
    c.name,
    IF(c.organisation_id IS NULL, false, true) AS can_edit,
    EXISTS(
        SELECT 1 FROM transactions AS t
        WHERE t.category_id = c.id
          AND t.organisation_id = get_current_user_organisation_id(?)
    ) AS in_use,
    COUNT(*) OVER () AS total_count
FROM
    categories AS c
WHERE c.organisation_id IS NULL
   OR c.organisation_id = get_current_user_organisation_id(?)
ORDER BY c.name
LIMIT ?
OFFSET ?
