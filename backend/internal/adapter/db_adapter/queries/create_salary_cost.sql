INSERT INTO salary_costs
    (
     cycle,
     amount_type,
     amount,
     distribution_type,
     relative_offset,
     target_date,
     label_id,
     salary_id,
     uuid,
     scenario_id
    )
SELECT ?, ?, ?, ?, ?, ?, ?, ?, UUID(), h.scenario_id
FROM salaries AS h
JOIN employees AS e ON e.id = h.employee_id
WHERE h.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;
