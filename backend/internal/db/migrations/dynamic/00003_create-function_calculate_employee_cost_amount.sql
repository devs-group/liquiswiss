-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION calculate_employee_cost_amount(
    salary BIGINT,
    amount BIGINT,
    amount_type ENUM('fixed', 'percentage')
)
    RETURNS BIGINT
    DETERMINISTIC
BEGIN
    DECLARE calculated_amount BIGINT;

    SET calculated_amount = CAST(
        -- Precision of 3 decimals
        IF(amount_type = 'percentage', salary * amount / 100000, amount)
        AS SIGNED
    );
    RETURN calculated_amount;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS calculate_employee_cost_amount;
-- +goose StatementEnd
