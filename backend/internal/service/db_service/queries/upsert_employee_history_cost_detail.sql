INSERT INTO employee_history_cost_details (month, amount, divider, cost_id)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    month = VALUES(month),
    amount = VALUES(amount),
    divider = VALUES(divider),
    cost_id = VALUES(cost_id);