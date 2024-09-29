DELETE h
FROM go_employee_history h
JOIN go_employees e ON e.id = h.employee_id
WHERE h.id = ?   -- The ID of the history entry
  AND e.owner = ?; -- The current user (owner) ID