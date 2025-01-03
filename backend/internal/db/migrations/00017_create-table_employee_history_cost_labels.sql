-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employee_history_cost_labels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,

    organisation_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Cost_Organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE ON UPDATE CASCADE,

    CONSTRAINT CK_Name_Not_Empty CHECK (name <> ''),

    CONSTRAINT UQ_Organisation_Name UNIQUE (organisation_id, name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employee_history_cost_labels;
-- +goose StatementEnd
