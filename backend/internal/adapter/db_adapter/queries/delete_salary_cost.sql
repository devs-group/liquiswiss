DELETE hc
FROM salary_costs hc
JOIN salaries AS h ON h.id = hc.salary_id
JOIN employees e ON e.id = h.employee_id
WHERE hc.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)