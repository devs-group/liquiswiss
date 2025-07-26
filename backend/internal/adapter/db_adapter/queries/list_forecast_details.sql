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
    COALESCE(f.revenue, json_array()) AS revenue,
    COALESCE(f.expense, json_array()) AS expense,
    f.forecast_id AS forecast_id
FROM date_series ds
LEFT JOIN forecast_details f ON DATE_FORMAT(ds.date, '%Y-%m') = f.month
    AND f.organisation_id = get_current_user_organisation_id(?)
ORDER BY ds.date