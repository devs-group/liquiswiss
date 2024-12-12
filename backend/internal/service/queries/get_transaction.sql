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
    v.id,
    v.value,
    CONCAT(FORMAT(v.value / 100, IF(v.value % 10 = 0, 1, 2)), '%') AS formatted_value,
    IF(v.owner IS NULL, false, true) AS can_edit
FROM
    go_transactions r
    INNER JOIN go_categories c ON r.category = c.id
    INNER JOIN go_currencies cur ON r.currency = cur.id
    LEFT JOIN vats v ON r.vat = v.id
    LEFT JOIN go_employees emp ON r.employee = emp.id
WHERE
    r.id = ?
    AND r.owner = ?