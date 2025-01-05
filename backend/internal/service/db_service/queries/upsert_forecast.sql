INSERT INTO forecasts (month, revenue, expense, cashflow, organisation_id)
VALUES (?, ?, ?, ?, (SELECT current_organisation_id FROM users u WHERE u.id = ?))
ON DUPLICATE KEY UPDATE
    revenue = VALUES(revenue),
    expense = VALUES(expense),
    cashflow = VALUES(cashflow);