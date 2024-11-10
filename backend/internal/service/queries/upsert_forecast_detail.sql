INSERT INTO go_forecast_details (owner, month, revenue, expense, forecast_id)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    revenue = VALUES(revenue),
    expense = VALUES(expense);