-- +goose Up
-- +goose StatementBegin
UPDATE scenarios SET name = 'Standardszenario' WHERE name = 'Default' AND is_default = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE scenarios SET name = 'Default' WHERE name = 'Standardszenario' AND is_default = true;
-- +goose StatementEnd
