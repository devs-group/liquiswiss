WITH RECURSIVE date_series AS (
    -- Start at the current month
    SELECT LAST_DAY(CURDATE()) AS date
    UNION ALL
    -- Increment each iteration by one month
    SELECT LAST_DAY(DATE_ADD(date, INTERVAL 1 MONTH))
    FROM date_series
    -- Generate up to N months
    WHERE date < LAST_DAY(DATE_ADD(CURDATE(), INTERVAL (?-1) MONTH))
)
SELECT
    DATE_FORMAT(ds.date, '%Y-%m') AS month,
    COALESCE(f.revenue, 0) AS revenue,
    COALESCE(f.expense, 0) AS expense,
    COALESCE(f.cashflow, 0) AS cashflow,
    f.updated_at AS updated_at
FROM date_series ds
LEFT JOIN forecasts f ON DATE_FORMAT(ds.date, '%Y-%m') = f.month
    AND f.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)
ORDER BY ds.date