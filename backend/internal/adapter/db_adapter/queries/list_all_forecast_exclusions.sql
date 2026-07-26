SELECT te.exclude_month, 'transactions' AS related_table, t.id, t.name, t.amount
FROM transaction_exclusions te
JOIN transactions t ON t.id = te.transaction_id
WHERE t.organisation_id = get_current_user_organisation_id(?)

UNION ALL

SELECT se.exclude_month, 'salaries' AS related_table, s.id, e.name, s.amount
FROM salary_exclusions se
JOIN salaries s ON s.id = se.salary_id
JOIN employees e ON e.id = s.employee_id
WHERE e.organisation_id = get_current_user_organisation_id(?)

UNION ALL

SELECT sce.exclude_month, 'salary_cost_labels' AS related_table, scl.id, scl.name, 0
FROM salary_cost_exclusions sce
JOIN salary_cost_labels scl ON scl.id = sce.label_id
WHERE scl.organisation_id = get_current_user_organisation_id(?)

ORDER BY exclude_month
