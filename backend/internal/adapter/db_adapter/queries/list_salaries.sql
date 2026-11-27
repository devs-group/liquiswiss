SELECT
    h.id,
    COUNT(*) OVER() AS total_count
FROM salaries h
JOIN employees e ON e.id = h.employee_id
WHERE h.employee_id = ?
  AND e.scenario_id = get_current_user_scenario_id(?)
ORDER BY h.from_date DESC
LIMIT ? OFFSET ?