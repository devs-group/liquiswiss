SELECT
    c.id,
    c.name,
    COUNT(*) OVER () AS total_count
FROM
    go_categories AS c
ORDER BY c.name
LIMIT ?
OFFSET ?
