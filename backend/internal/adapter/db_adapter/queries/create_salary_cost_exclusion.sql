INSERT IGNORE INTO salary_cost_exclusions (exclude_month, label_id, scenario_id)
SELECT ?, ehc.label_id, e.scenario_id
FROM salary_costs ehc
JOIN salaries eh ON eh.id = ehc.salary_id
JOIN employees e ON e.id = eh.employee_id
WHERE ehc.id = ?
  AND e.scenario_id = get_current_user_scenario_id(?)
LIMIT 1;