INSERT INTO go_forecasts (owner, month, revenue, expense, cashflow)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    revenue = VALUES(revenue),
    expense = VALUES(expense),
    cashflow = VALUES(cashflow);