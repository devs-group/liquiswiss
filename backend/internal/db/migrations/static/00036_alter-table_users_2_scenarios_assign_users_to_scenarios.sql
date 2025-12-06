-- +goose Up
-- +goose StatementBegin
INSERT INTO users_2_scenarios (user_id, organisation_id, scenario_id)
    SELECT u.id, u2o.organisation_id, sc.id FROM users AS u
    JOIN users_2_organisations AS u2o ON u2o.user_id = u.id
    JOIN scenarios AS sc ON sc.organisation_id = u2o.organisation_id;
-- +goose StatementEnd
