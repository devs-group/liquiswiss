-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users_2_scenarios (
    user_id BIGINT UNSIGNED NOT NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,
    scenario_id BIGINT UNSIGNED NOT NULL,
    is_current BOOLEAN DEFAULT FALSE,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,

    PRIMARY KEY (user_id, organisation_id, scenario_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for performance when querying current scenario
CREATE INDEX IF NOT EXISTS idx_users_2_scenarios_current
ON users_2_scenarios(user_id, organisation_id, is_current);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO users_2_scenarios (user_id, organisation_id, scenario_id, is_current)
SELECT u.id, u2o.organisation_id, sc.id, sc.is_default
FROM users AS u
JOIN users_2_organisations AS u2o ON u2o.user_id = u.id
JOIN scenarios AS sc ON sc.organisation_id = u2o.organisation_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users_2_scenarios;
-- +goose StatementEnd
