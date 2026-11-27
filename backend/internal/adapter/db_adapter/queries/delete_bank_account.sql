DELETE FROM bank_accounts
WHERE
    id = ?
    AND scenario_id = get_current_user_scenario_id(?)
