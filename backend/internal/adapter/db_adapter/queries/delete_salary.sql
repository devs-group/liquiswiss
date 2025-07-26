DELETE h
FROM salaries h
JOIN employees e ON e.id = h.employee_id
WHERE h.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)