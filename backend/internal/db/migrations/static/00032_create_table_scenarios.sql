-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS scenarios (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    is_default BOOL NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    parent_scenario_id BIGINT UNSIGNED NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,
    CONSTRAINT fk_scenarios_parent FOREIGN KEY (parent_scenario_id) REFERENCES scenarios (id) ON DELETE SET NULL,
    CONSTRAINT fk_scenarios_organisation FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE CASCADE
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO scenarios (name, is_default, organisation_id)
SELECT 'Standard', TRUE, o.id
FROM organisations o
         LEFT JOIN scenarios s ON s.organisation_id = o.id AND s.is_default = TRUE
WHERE s.id IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS scenarios;
-- +goose StatementEnd
