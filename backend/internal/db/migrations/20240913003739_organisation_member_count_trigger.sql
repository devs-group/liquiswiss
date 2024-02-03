-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER increase_member_count
    AFTER INSERT ON go_users_2_organisations
    FOR EACH ROW
BEGIN
    UPDATE go_organisations
    SET member_count = member_count + 1
    WHERE id = NEW.organisation_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER decrease_member_count
    AFTER DELETE ON go_users_2_organisations
    FOR EACH ROW
BEGIN
    UPDATE go_organisations
    SET member_count = member_count - 1
    WHERE id = OLD.organisation_id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS decrease_member_count;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TRIGGER IF EXISTS increase_member_count;
-- +goose StatementEnd
