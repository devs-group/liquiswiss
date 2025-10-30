SELECT
    sc.id,
    scl.id      AS label_id,
    scl.name    AS label_name,
    sc.cycle,
    sc.amount_type,
    sc.amount,
    sc.distribution_type,
    sc.relative_offset,
    sc.target_date,
    sc.salary_id,
    s.cycle     AS salary_cycle,
    s.amount,
    s.from_date AS salary_from_date,
    s.to_date AS salary_to_date,
    CURDATE() AS db_date
FROM salary_costs sc
JOIN salaries s ON s.id = sc.salary_id
JOIN employees e ON e.id = s.employee_id
LEFT JOIN salary_cost_labels scl ON scl.id = sc.label_id
WHERE sc.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
