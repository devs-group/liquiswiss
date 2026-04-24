-- +goose Up
ALTER TABLE organisation_chatbots ADD COLUMN IF NOT EXISTS skill_id VARCHAR(255) DEFAULT NULL AFTER chatbot_id;

-- +goose Down
ALTER TABLE organisation_chatbots DROP COLUMN IF EXISTS skill_id;
