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

-- +goose StatementBegin
-- Create a test invitation for E2E testing
-- Token: e2e-test-invitation-token
-- Email: e2e-invite-new@test.liquiswiss.ch (new user)
INSERT INTO organisation_invitations (id, organisation_id, email, role, token, invited_by, expires_at, created_at)
VALUES (9999, 9999, 'e2e-invite-new@test.liquiswiss.ch', 'editor', 'e2e-test-invitation-token-new', 9999, DATE_ADD(NOW(), INTERVAL 7 DAY), NOW())
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose StatementBegin
-- Create a user that will receive an invitation (existing user)
INSERT INTO users (id, email, password, name, deleted, created_at, current_organisation_id)
VALUES (9998, 'e2e-invite-existing@test.liquiswiss.ch', '$2a$12$slvyLgYZpFsHbSORSsbzkeVBUDWykOhCFl6S.R7b4zG.8O853Skea', 'E2E Existing Invite User', 0, NOW(), NULL)
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose StatementBegin
-- Create user settings for the existing user
INSERT INTO user_settings (id, user_id, created_at)
VALUES (9998, 9998, NOW())
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose StatementBegin
-- Create an invitation for an existing user for E2E testing
-- Token: e2e-test-invitation-token-existing
INSERT INTO organisation_invitations (id, organisation_id, email, role, token, invited_by, expires_at, created_at)
VALUES (9998, 9999, 'e2e-invite-existing@test.liquiswiss.ch', 'admin', 'e2e-test-invitation-token-existing', 9999, DATE_ADD(NOW(), INTERVAL 7 DAY), NOW())
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose StatementBegin
-- Create an expired invitation for testing error handling
INSERT INTO organisation_invitations (id, organisation_id, email, role, token, invited_by, expires_at, created_at)
VALUES (9997, 9999, 'e2e-invite-expired@test.liquiswiss.ch', 'read-only', 'e2e-test-invitation-token-expired', 9999, DATE_SUB(NOW(), INTERVAL 1 DAY), NOW())
ON DUPLICATE KEY UPDATE id = id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- NO DOWN MIGRATION NEEDED (dynamic migrations are dropped and recreated)
-- +goose StatementEnd
