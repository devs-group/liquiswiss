-- +goose Up
-- E2E Test User Fixture
-- This creates a test user for E2E testing with Playwright
-- Email: e2e@test.liquiswiss.ch
-- Password: Test123!

-- +goose StatementBegin
-- Create organisation for E2E test user
INSERT INTO organisations (id, name, deleted, created_at)
VALUES (9999, 'E2E Test Organisation', 0, NOW())
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose StatementBegin
-- Create E2E test user with bcrypt-hashed password (Test123!)
INSERT INTO users (id, email, password, name, deleted, created_at, current_organisation_id)
VALUES (9999, 'e2e@test.liquiswiss.ch', '$2a$12$slvyLgYZpFsHbSORSsbzkeVBUDWykOhCFl6S.R7b4zG.8O853Skea', 'E2E Test User', 0, NOW(), 9999)
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose StatementBegin
-- Link user to organisation with owner role
INSERT INTO users_2_organisations (user_id, organisation_id, role, is_default, created_at)
VALUES (9999, 9999, 'owner', 1, NOW())
ON DUPLICATE KEY UPDATE user_id = user_id;
-- +goose StatementEnd

-- +goose StatementBegin
-- Create user settings for E2E test user
INSERT INTO user_settings (id, user_id, created_at)
VALUES (9999, 9999, NOW())
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose StatementBegin
-- Create user organisation settings for E2E test user
INSERT INTO user_organisation_settings (id, user_id, organisation_id, created_at)
VALUES (9999, 9999, 9999, NOW())
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- NO DOWN MIGRATION NEEDED (dynamic migrations are dropped and recreated)
-- +goose StatementEnd
