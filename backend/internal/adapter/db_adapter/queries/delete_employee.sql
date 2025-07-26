DELETE FROM employees
WHERE
    id = ?
    AND organisation_id = get_current_user_organisation_id(?)
