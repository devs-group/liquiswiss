INSERT INTO salary_cost_details (month, amount, divider, is_extra_month, cost_id)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    month = VALUES(month),
    amount = VALUES(amount),
    divider = VALUES(divider),
    is_extra_month = VALUES(is_extra_month),
    cost_id = VALUES(cost_id);