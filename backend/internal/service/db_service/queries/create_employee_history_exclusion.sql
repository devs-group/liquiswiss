INSERT IGNORE INTO employee_histories_exclusions (exclude_month, employee_history_id)
SELECT ?, eh.id
FROM employee_histories eh
JOIN employees e ON e.id = eh.employee_id
WHERE eh.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;