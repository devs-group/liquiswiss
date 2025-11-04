SELECT
    r.id,
    r.name,
    r.amount,
    IF(v.id,
       calculate_vat_amount(r.vat_included, r.amount, v.value),
       0
    ) AS vat_amount,
    r.vat_included,
    r.is_disabled,
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
    CURDATE() AS db_date
FROM
    transactions r
    INNER JOIN categories c ON r.category_id = c.id
    INNER JOIN currencies cur ON r.currency_id = cur.id
    LEFT JOIN vats v ON r.vat_id = v.id
    LEFT JOIN employees emp ON r.employee_id = emp.id
WHERE
    r.id = ?
    AND r.organisation_id = get_current_user_organisation_id(?)
    AND r.scenario_id = (SELECT current_scenario_id FROM users WHERE id = ?)
