DELETE FROM transactions
WHERE
    id = ?
    AND organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)
