DELETE FROM transactions
WHERE
    id = ?
    AND scenario_id = get_current_user_scenario_id(?)
