SELECT
    c.id,
    c.code,
    c.description,
    c.locale_code
FROM
    go_currencies AS c
WHERE
    c.id = ?