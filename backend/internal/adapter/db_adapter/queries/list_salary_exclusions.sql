SELECT
    ehe.id,
    ehe.exclude_month,
    ehe.salary_id
FROM salary_exclusions ehe
JOIN salaries AS eh ON eh.id = ehe.salary_id
JOIN employees AS e ON e.id = eh.employee_id
WHERE ehe.salary_id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)