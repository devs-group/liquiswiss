SELECT
    c.id,
    c.code,
    c.description,
    c.locale_code,
    COUNT(*) OVER () AS total_count
FROM
    go_currencies AS c
LIMIT ?
OFFSET ?
