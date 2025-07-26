INSERT IGNORE INTO salary_exclusions (exclude_month, salary_id)
SELECT ?, eh.id
FROM salaries eh
JOIN employees e ON e.id = eh.employee_id
WHERE eh.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)
LIMIT 1;