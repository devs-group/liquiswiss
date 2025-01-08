DELETE FROM employee_history_cost_labels
WHERE
    id = ?
    AND organisation_id = get_current_user_organisation_id(?)