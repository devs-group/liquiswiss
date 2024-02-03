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
    cur.locale_code
FROM
    go_transactions r
    INNER JOIN go_categories c ON r.category = c.id
    INNER JOIN go_currencies cur ON r.currency = cur.id
WHERE
    r.id = ?
    AND r.owner = ?