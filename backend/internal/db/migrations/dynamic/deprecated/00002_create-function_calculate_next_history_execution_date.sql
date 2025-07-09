-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION calculate_next_history_execution_date(
    type ENUM('single', 'repeating'),
    from_date DATE,
    to_date DATE,
    cycle ENUM('monthly', 'quarterly', 'biannually', 'yearly'),
    curr_date DATE
)
    RETURNS DATE
    DETERMINISTIC
BEGIN
    return CASE
       WHEN type = 'single' THEN
           IF(to_date IS NULL OR from_date <= to_date,
              from_date,
              from_date)
       WHEN type = 'repeating' AND cycle = 'monthly' THEN
           IF(to_date IS NULL OR DATE_ADD(from_date, INTERVAL (TIMESTAMPDIFF(MONTH, from_date, curr_date) + 1) * 1 MONTH) <= to_date,
              IF(curr_date < from_date,
                 from_date,
                 DATE_ADD(from_date, INTERVAL (TIMESTAMPDIFF(MONTH, from_date, curr_date) + 1) * 1 MONTH)
              ),
              DATE_ADD(from_date, INTERVAL (TIMESTAMPDIFF(MONTH, from_date, to_date)) * 1 MONTH)
           )
       WHEN type = 'repeating' AND cycle = 'quarterly' THEN
           IF(to_date IS NULL OR DATE_ADD(from_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, from_date, curr_date) / 3) + 1) * 3 MONTH) <= to_date,
              IF(curr_date < from_date,
                 from_date,
                 DATE_ADD(from_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, from_date, curr_date) / 3) + 1) * 3 MONTH)
              ),
              DATE_ADD(from_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, from_date, to_date) / 3)) * 3 MONTH)
           )
       WHEN type = 'repeating' AND cycle = 'biannually' THEN
           IF(to_date IS NULL OR DATE_ADD(from_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, from_date, curr_date) / 6) + 1) * 6 MONTH) <= to_date,
              IF(curr_date < from_date,
                 from_date,
                 DATE_ADD(from_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, from_date, curr_date) / 6) + 1) * 6 MONTH)
              ),
              DATE_ADD(from_date, INTERVAL FLOOR((TIMESTAMPDIFF(MONTH, from_date, to_date) / 6)) * 6 MONTH)
           )
       WHEN type = 'repeating' AND cycle = 'yearly' THEN
           IF(to_date IS NULL OR DATE_ADD(from_date, INTERVAL TIMESTAMPDIFF(YEAR, from_date, curr_date) + 1 YEAR) <= to_date,
              IF(curr_date < from_date,
                 from_date,
                 DATE_ADD(from_date, INTERVAL (TIMESTAMPDIFF(YEAR, from_date, curr_date) + 1) * 1 YEAR)
              ),
              DATE_ADD(from_date, INTERVAL (TIMESTAMPDIFF(YEAR, from_date, to_date)) * 1 YEAR)
           )
    END;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS calculate_next_history_execution_date;
-- +goose StatementEnd
