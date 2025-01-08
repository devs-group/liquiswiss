INSERT INTO employee_history_costs
    (
     cycle,
     amount_type,
     amount,
     distribution_type,
     relative_offset,
     target_date,
     label_id,
     employee_history_id
    )
SELECT ?, ?, ?, ?, ?, ?, ?, ?
FROM employee_histories AS h
JOIN employees AS e ON e.id = h.employee_id
WHERE h.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;