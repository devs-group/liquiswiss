-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS salary_cost_base_links (
    cost_id BIGINT UNSIGNED NOT NULL,
    base_cost_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (cost_id, base_cost_id),
    CONSTRAINT FK_BaseLink_Cost FOREIGN KEY (cost_id) REFERENCES salary_costs (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_BaseLink_BaseCost FOREIGN KEY (base_cost_id) REFERENCES salary_costs (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS salary_cost_base_links;
-- +goose StatementEnd
