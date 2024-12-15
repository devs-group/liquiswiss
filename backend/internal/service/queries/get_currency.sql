SELECT
    c.id,
    c.code,
    c.description,
    c.locale_code
FROM
    currencies AS c
WHERE
    c.id = ?