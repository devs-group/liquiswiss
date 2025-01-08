-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION get_current_user_organisation_currency_id(organisation_id BIGINT)
    RETURNS BIGINT UNSIGNED
    DETERMINISTIC
BEGIN
    DECLARE currency_id BIGINT UNSIGNED;

    SELECT IF(o.main_currency_id, o.main_currency_id, get_default_system_currency_id())
    INTO currency_id
    FROM organisations o
    WHERE o.id = organisation_id;

    RETURN currency_id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_current_user_organisation_currency_id;
-- +goose StatementEnd
