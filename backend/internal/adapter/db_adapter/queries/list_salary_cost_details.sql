SELECT
    hcd.id,
    hcd.month,
    hcd.amount,
    hcd.divider,
    hcd.is_extra_month,
    hcd.cost_id
FROM salary_cost_details hcd
WHERE hcd.cost_id = ?
ORDER BY
    hcd.month