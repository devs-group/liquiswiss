-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users_2_scenarios (
    user_id BIGINT UNSIGNED NOT NULL ,
    organisation_id BIGINT UNSIGNED NOT NULL ,
    scenario_id BIGINT UNSIGNED NOT NULL ,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (organisation_id) REFERENCES organisations (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (scenario_id) REFERENCES scenarios (id) ON DELETE CASCADE ON UPDATE CASCADE,

    PRIMARY KEY (user_id, organisation_id, scenario_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users_2_scenarios;
-- +goose StatementEnd
