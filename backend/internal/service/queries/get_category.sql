SELECT
    c.id,
    c.name,
    IF(c.organisation_id IS NULL, false, true) AS can_edit
FROM
    categories AS c
WHERE
    c.id = ?
    AND (c.organisation_id IS NULL
        OR c.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)
    )