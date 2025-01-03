SELECT
    c.id,
    c.name,
    IF(c.organisation_id IS NULL, false, true) AS can_edit
FROM
    categories AS c
WHERE
    c.id = ?
    AND
    (
        c.organisation_id IS NULL
        OR c.organisation_id = get_current_organisation(?)
    )