SELECT
    ehe.id,
    ehe.exclude_month,
    ehe.employee_history_id
FROM employee_histories_exclusions ehe
JOIN employee_histories AS eh ON eh.id = ehe.employee_history_id
JOIN employees AS e ON e.id = eh.employee_id
WHERE ehe.employee_history_id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)