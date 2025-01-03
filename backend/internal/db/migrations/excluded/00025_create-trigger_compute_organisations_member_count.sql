-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER employee_history_costs_calculated_amount
    BEFORE UPDATE ON employee_history_costs
    FOR EACH ROW
BEGIN
    DECLARE salary_value BIGINT;

    -- Fetch the associated salary
    SELECT h.salary INTO salary_value
    FROM employee_histories h
    WHERE h.id = NEW.employee_history_id;

    -- Calculate the calculated_amount
    SET NEW.calculated_amount = CAST(
        IF(NEW.amount_type = 'percentage', salary_value * NEW.amount / 100000, NEW.amount)
        AS SIGNED
    );
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS employee_history_costs_calculated_amount;
-- +goose StatementEnd
