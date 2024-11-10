SELECT
    r.id,
    r.name,
    r.amount,
    r.cycle,
    r.type,
    r.start_date,
    r.end_date,
    c.id,
    c.name,
    cur.id,
    cur.code,
    cur.description,
    cur.locale_code,
    emp.id,
    emp.name,
    COUNT(*) OVER () AS total_count
FROM
    go_transactions AS r
    INNER JOIN go_categories c ON r.category = c.id
    INNER JOIN go_currencies cur ON r.currency = cur.id
    LEFT JOIN go_employees emp ON r.employee = emp.id
WHERE
    r.owner = ?
ORDER BY
    %s IS NULL,
    %s %s,
    r.name ASC
LIMIT ?
OFFSET
    ?