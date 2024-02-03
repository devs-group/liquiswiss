SELECT
    e.id,
    e.name,
    e.hours_per_month,
    e.vacation_days_per_year,
    e.entry_date,
    e.exit_date
FROM
    go_employees AS e
WHERE
    e.id = ?
    AND e.owner = ?
