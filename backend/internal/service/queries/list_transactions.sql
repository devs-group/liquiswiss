SELECT
    r.id,
    r.name,
    r.amount,
    IF(v.id,
       IF(r.vat_included,
          (r.amount * v.value) DIV (10000 + v.value),
          (r.amount * v.value) DIV 10000
       ),
    0) AS vat_amount,
    r.vat_included,
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
    IF(v.organisation_id IS NULL, false, true) AS can_edit,
    CASE
        WHEN r.type = 'single' THEN
            IF(r.end_date IS NULL OR r.start_date <= r.end_date,
               r.start_date,
               NULL)
        WHEN r.type = 'repeating' AND r.cycle = 'daily' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL (TIMESTAMPDIFF(DAY, r.start_date, CURDATE()) + 1) * 1 DAY) <= r.end_date,
               IF(CURDATE() < r.start_date,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL (TIMESTAMPDIFF(DAY, r.start_date, CURDATE()) + 1) * 1 DAY)
               ),
               NULL
            )
        WHEN r.type = 'repeating' AND r.cycle = 'weekly' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL (TIMESTAMPDIFF(WEEK, r.start_date, CURDATE()) + 1) * 7 DAY) <= r.end_date,
               IF(CURDATE() < r.start_date,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL (TIMESTAMPDIFF(WEEK, r.start_date, CURDATE()) + 1) * 7 DAY)
               ),
               NULL
            )
        WHEN r.type = 'repeating' AND r.cycle = 'monthly' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL (TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) + 1) * 1 MONTH) <= r.end_date,
               IF(CURDATE() < r.start_date,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL (TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) + 1) * 1 MONTH)
               ),
               NULL
            )
        WHEN r.type = 'repeating' AND r.cycle = 'quarterly' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) / 3) + 1) * 3 MONTH) <= r.end_date,
               IF(CURDATE() < r.start_date,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) / 3) + 1) * 3 MONTH)
               ),
               NULL
            )
        WHEN r.type = 'repeating' AND r.cycle = 'biannually' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) / 6) + 1) * 6 MONTH) <= r.end_date,
               IF(CURDATE() < r.start_date,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, r.start_date, CURDATE()) / 6) + 1) * 6 MONTH)
               ),
               NULL
            )
        WHEN r.type = 'repeating' AND r.cycle = 'yearly' THEN
            IF(r.end_date IS NULL OR DATE_ADD(r.start_date, INTERVAL TIMESTAMPDIFF(YEAR, r.start_date, CURDATE()) + 1 YEAR) <= r.end_date,
               IF(CURDATE() < r.start_date,
                  r.start_date,
                  DATE_ADD(r.start_date, INTERVAL TIMESTAMPDIFF(YEAR, r.start_date, CURDATE()) + 1 YEAR)
               ),
               NULL
            )
        END AS next_execution_date,
    COUNT(*) OVER () AS total_count
FROM
    transactions AS r
    INNER JOIN categories c ON r.category_id = c.id
    INNER JOIN currencies cur ON r.currency_id = cur.id
    LEFT JOIN vats v ON r.vat_id = v.id
    LEFT JOIN employees emp ON r.employee_id = emp.id
WHERE
    r.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)
ORDER BY
    {{.sortBy}} IS NULL,
    {{.sortBy}} {{.sortOrder}},
    r.name ASC
LIMIT ?
OFFSET
    ?