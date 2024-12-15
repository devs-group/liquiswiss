DELETE h
FROM employee_history h
JOIN employees e ON e.id = h.employee_id
WHERE h.id = ?
  AND e.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)