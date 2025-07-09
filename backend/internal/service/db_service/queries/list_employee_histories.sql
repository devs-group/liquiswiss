SELECT
    h.id,
    COUNT(*) OVER() AS total_count
FROM employee_histories h
JOIN employees e ON e.id = h.employee_id
WHERE h.employee_id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
ORDER BY h.from_date DESC
LIMIT ? OFFSET ?