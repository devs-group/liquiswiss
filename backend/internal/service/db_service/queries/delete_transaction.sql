DELETE FROM transactions
WHERE
    id = ?
    AND organisation_id = get_current_user_organisation_id(?)
