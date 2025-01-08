DELETE hc
FROM employee_history_costs hc
JOIN employee_histories AS h ON h.id = hc.employee_history_id
JOIN employees e ON e.id = h.employee_id
WHERE hc.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)