SELECT
    r.id,
    COUNT(*) OVER () AS total_count
FROM
    transactions AS r
    INNER JOIN categories c ON r.category_id = c.id
    INNER JOIN currencies cur ON r.currency_id = cur.id
    LEFT JOIN vats v ON r.vat_id = v.id
    LEFT JOIN employees emp ON r.employee_id = emp.id
WHERE
    r.organisation_id = get_current_user_organisation_id(?)
    {{if .hasSearch}}AND LOWER(r.name) LIKE LOWER(?){{end}}
ORDER BY
    {{.sortBy}} IS NULL,
    {{.sortBy}} {{.sortOrder}},
    r.name ASC
LIMIT ?
OFFSET
    ?