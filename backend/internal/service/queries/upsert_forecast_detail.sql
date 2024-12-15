INSERT INTO forecast_details (month, revenue, expense, forecast_id, organisation_id)
VALUES (?, ?, ?, ?, (SELECT current_organisation FROM users u WHERE u.id = ?))
ON DUPLICATE KEY UPDATE
    revenue = VALUES(revenue),
    expense = VALUES(expense);