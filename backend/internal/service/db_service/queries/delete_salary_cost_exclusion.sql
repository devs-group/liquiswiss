DELETE ehce
FROM salary_cost_exclusions ehce
JOIN salary_costs ehc ON ehc.label_id = ehce.label_id
JOIN salaries eh ON eh.id = ehc.salary_id
JOIN employees AS e ON e.id = eh.employee_id
WHERE ehce.exclude_month = ?
  AND ehc.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)