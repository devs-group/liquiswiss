INSERT IGNORE INTO employee_history_cost_exclusions (exclude_month, label_id)
SELECT ?, ehc.label_id
FROM employee_history_costs ehc
JOIN employee_histories eh ON eh.id = ehc.employee_history_id
JOIN employees e ON e.id = eh.employee_id
WHERE ehc.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;