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
    CASE
        WHEN r.type = 'single' THEN
            IF(r.end_date IS NULL OR r.start_date <= r.end_date,
               r.start_date,
               NULL)
        WHEN r.type = 'repeating' AND r.cycle = 'daily' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL DATEDIFF(CURDATE(), r.start_date) DIV 1 + 1 DAY) <= r.end_date,
               DATE_ADD(r.start_date, INTERVAL DATEDIFF(CURDATE(), r.start_date) DIV 1 + 1 DAY),
               NULL)
        WHEN r.type = 'repeating' AND r.cycle = 'weekly' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL DATEDIFF(CURDATE(), r.start_date) DIV 7 + 1 WEEK) <= r.end_date,
               DATE_ADD(r.start_date, INTERVAL DATEDIFF(CURDATE(), r.start_date) DIV 7 + 1 WEEK),
               NULL)
        WHEN r.type = 'repeating' AND r.cycle = 'monthly' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) + 1 MONTH) <= r.end_date,
               IF(TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) < 0,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) + 1 MONTH)),
               NULL)
        WHEN r.type = 'repeating' AND r.cycle = 'quarterly' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL (FLOOR(TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) / 3) + 1) * 3 MONTH) <= r.end_date,
               IF(TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) < 0,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL (FLOOR(TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) / 3) + 1) * 3 MONTH)),
               NULL)
        WHEN r.type = 'repeating' AND r.cycle = 'biannually' THEN
            IF(r.end_date IS NULL OR
               DATE_ADD(r.start_date, INTERVAL (FLOOR(TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) / 6) + 1) * 6 MONTH) <= r.end_date,
               IF(TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) < 0,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL (FLOOR(TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) / 6) + 1) * 6 MONTH)),
               NULL)
        WHEN r.type = 'repeating' AND r.cycle = 'yearly' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL TIMESTAMPDIFF(YEAR, r.start_date, CURDATE()) + 1 YEAR) <= r.end_date,
               IF(TIMESTAMPDIFF(YEAR, r.start_date, CURDATE()) < 0,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL TIMESTAMPDIFF(YEAR, r.start_date, CURDATE()) + 1 YEAR)),
               NULL)
        END AS next_execution_date,
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