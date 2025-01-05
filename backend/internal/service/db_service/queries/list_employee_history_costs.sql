SELECT
    hc.id,
    hcl.id AS label_id,
    hcl.name AS label_name,
    hc.cycle,
    hc.amount_type,
    hc.amount,
    hc.distribution_type,
    calculate_employee_cost_amount(
        h.salary,
        hc.amount,
        hc.amount_type
    ) AS calculated_amount,
    hc.relative_offset,
    hc.target_date,
    calculate_cost_execution_date(
        'repeating',
        h.from_date,
        h.to_date,
        h.cycle,
        hc.target_date,
        hc.cycle,
        hc.relative_offset,
        CURDATE(),
        false
    ) AS previous_execution_date,
    calculate_cost_execution_date(
        'repeating',
        h.from_date,
        h.to_date,
        h.cycle,
        hc.target_date,
        hc.cycle,
        hc.relative_offset,
        CURDATE(),
        true
    ) AS next_execution_date,
    calculate_next_cost_amount(
        'repeating',
        h.from_date,
        h.to_date,
        h.cycle,
        hc.target_date,
        hc.cycle,
        hc.relative_offset,
        CURDATE(),
        hc.amount_type,
        hc.amount,
        h.salary
    ) AS next_cost,
    hc.employee_history_id,
    COUNT(*) OVER() AS total_count
FROM employee_history_costs hc
JOIN employee_histories h ON h.id = hc.employee_history_id
JOIN employees e ON e.id = h.employee_id
LEFT JOIN employee_history_cost_labels hcl ON hcl.id = hc.label_id
WHERE hc.employee_history_id = ?
  AND e.organisation_id = get_current_organisation(?)
ORDER BY
    hcl.name,
    hc.distribution_type DESC,
    next_execution_date IS NULL,
    next_execution_date
LIMIT ? OFFSET ?