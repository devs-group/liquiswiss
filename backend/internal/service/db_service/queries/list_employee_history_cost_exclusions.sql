SELECT
    ehce.id,
    ehce.exclude_month,
    ehc.id
FROM employee_history_cost_exclusions ehce
JOIN employee_history_costs AS ehc ON ehc.label_id = ehce.label_id
JOIN employee_histories AS eh ON eh.id = ehc.employee_history_id
JOIN employees AS e ON e.id = eh.employee_id
WHERE ehc.id = ?
  AND e.organisation_id = get_current_user_organisation_id(?)