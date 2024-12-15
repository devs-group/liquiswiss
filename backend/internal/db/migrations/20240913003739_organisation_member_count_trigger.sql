-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER Increase_Member_Count
    AFTER INSERT ON users_2_organisations
    FOR EACH ROW
BEGIN
    UPDATE organisations
    SET member_count = member_count + 1
    WHERE id = NEW.organisation_id;
END;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TRIGGER Decrease_Member_Count
    AFTER DELETE ON users_2_organisations
    FOR EACH ROW
BEGIN
    UPDATE organisations
    SET member_count = member_count - 1
    WHERE id = OLD.organisation_id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS Decrease_Member_Count;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS Increase_Member_Count;
-- +goose StatementEnd
