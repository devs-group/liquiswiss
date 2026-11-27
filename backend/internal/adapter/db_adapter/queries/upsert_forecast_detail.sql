INSERT INTO forecast_details (month, revenue, expense, forecast_id, organisation_id, scenario_id)
SELECT ?, ?, ?, ?, s.organisation_id, s.id
FROM scenarios s
WHERE s.id = get_current_user_scenario_id(?)
ON DUPLICATE KEY UPDATE
    revenue = VALUES(revenue),
    expense = VALUES(expense);