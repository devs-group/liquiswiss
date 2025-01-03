-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION get_current_organisation(user_id BIGINT UNSIGNED)
    RETURNS BIGINT UNSIGNED
    DETERMINISTIC
BEGIN
    DECLARE organisation_id BIGINT UNSIGNED;

    SELECT current_organisation_id
    INTO organisation_id
    FROM users
    WHERE id = user_id;

    RETURN organisation_id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_current_organisation;
-- +goose StatementEnd
