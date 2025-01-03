DELETE h
FROM employee_histories h
JOIN employees e ON e.id = h.employee_id
WHERE h.id = ?
  AND e.organisation_id = get_current_organisation(?)