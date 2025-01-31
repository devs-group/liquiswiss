DELETE ehce
FROM employee_history_cost_exclusions ehce
JOIN employee_history_costs ehc ON ehc.label_id = ehce.label_id
JOIN employee_histories eh ON eh.id = ehc.employee_history_id
JOIN employees AS e ON e.id = eh.employee_id
WHERE ehce.exclude_month = ?
  AND ehc.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)