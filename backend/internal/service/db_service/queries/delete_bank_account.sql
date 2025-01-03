DELETE FROM bank_accounts
WHERE
    id = ?
    AND organisation_id = get_current_organisation(?)
