SELECT
    c.id,
    c.name,
    COUNT(*) OVER () AS total_count
FROM
    go_categories AS c
LIMIT ?
OFFSET ?
