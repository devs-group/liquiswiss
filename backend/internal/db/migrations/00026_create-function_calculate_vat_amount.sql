-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION calculate_vat_amount(
    vat_included BOOL,
    amount BIGINT,
    value BIGINT
)
    RETURNS BIGINT
    DETERMINISTIC
BEGIN
    DECLARE calculated_amount BIGINT;

    SET calculated_amount = IF(
            vat_included,
            (amount * value) DIV (10000 + value),
            (amount * value) DIV 10000
    );

    RETURN calculated_amount;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS calculate_vat_amount;
-- +goose StatementEnd
