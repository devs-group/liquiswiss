SELECT
    c.id,
    c.code,
    c.description,
    c.locale_code,
    COUNT(*) OVER () AS total_count
FROM
    currencies AS c
LIMIT ?
OFFSET ?
