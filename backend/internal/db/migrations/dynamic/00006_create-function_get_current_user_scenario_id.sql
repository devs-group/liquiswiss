-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION get_current_user_scenario_id(user_id BIGINT UNSIGNED)
    RETURNS BIGINT UNSIGNED
    DETERMINISTIC
BEGIN
    DECLARE scenario_id BIGINT UNSIGNED;

    SELECT current_scenario_id
    INTO scenario_id
    FROM users
    WHERE id = user_id;

    RETURN scenario_id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_current_user_scenario_id;
-- +goose StatementEnd
