INSERT IGNORE INTO transaction_exclusions (exclude_month, transaction_id)
SELECT ?, t.id
FROM transactions t
WHERE t.id = ?
    AND t.scenario_id = get_current_user_scenario_id(?)
LIMIT 1;