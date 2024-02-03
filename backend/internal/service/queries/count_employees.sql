SELECT
    COUNT(*) OVER () AS total_count
FROM
    go_employees AS e
WHERE
    e.owner = ?
LIMIT ?
OFFSET ?
