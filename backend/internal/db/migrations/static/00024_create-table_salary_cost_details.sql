-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS salary_cost_details (
    id SERIAL PRIMARY KEY,
    month VARCHAR(7) NOT NULL,
    amount BIGINT UNSIGNED NOT NULL DEFAULT 0,
    divider INT UNSIGNED NOT NULL DEFAULT 1,

    cost_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Detail_SalaryCost FOREIGN KEY (cost_id) REFERENCES salary_costs(id) ON DELETE CASCADE,

    CONSTRAINT CK_Month_Not_Empty CHECK (month <> ''),

    CONSTRAINT UQ_Month_SalaryCostID UNIQUE (month, cost_id),

    INDEX IDX_Month_SalaryCostID (month, cost_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS salary_cost_details;
-- +goose StatementEnd