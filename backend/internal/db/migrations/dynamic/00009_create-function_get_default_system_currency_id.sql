-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION get_default_system_currency_id()
    RETURNS BIGINT UNSIGNED
    DETERMINISTIC
BEGIN
    DECLARE currency_id BIGINT UNSIGNED;

    SELECT id
    INTO currency_id
    FROM currencies
    -- Can be changed to globally set the system currency
    WHERE code = 'CHF';

    RETURN currency_id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_default_system_currency_id;
-- +goose StatementEnd
