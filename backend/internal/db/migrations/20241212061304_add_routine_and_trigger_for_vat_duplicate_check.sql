-- +goose Up
-- +goose StatementBegin
CREATE PROCEDURE CheckUniqueVat(id BIGINT, value BIGINT, organisation_id BIGINT UNSIGNED)
BEGIN
    -- Check for duplicates
    IF EXISTS (
        SELECT 1
        FROM vats AS v
        WHERE
            value = v.value
          AND id != v.id
          AND (
              -- System-level VAT
              v.organisation_id IS NULL
              -- Organisation-level VAT
              OR (v.organisation_id = organisation_id AND v.organisation_id IS NOT NULL)
            )
    ) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'Duplicate VAT entries are not allowed';
    END IF;
END;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TRIGGER Enforce_Unique_Vat_Insert
    BEFORE INSERT ON vats
    FOR EACH ROW
BEGIN
    CALL CheckUniqueVat(0, NEW.value, NEW.organisation_id);
END;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TRIGGER Enforce_Unique_Vat_Update
    BEFORE UPDATE ON vats
    FOR EACH ROW
BEGIN
    CALL CheckUniqueVat(NEW.id, NEW.value, NEW.organisation_id);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS Enforce_Unique_Vat_Update;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS Enforce_Unique_Vat_Insert;
-- +goose StatementEnd
-- +goose StatementBegin
DROP PROCEDURE IF EXISTS CheckUniqueVat;
-- +goose StatementEnd
