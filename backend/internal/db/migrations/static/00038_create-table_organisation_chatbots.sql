-- +goose Up
CREATE TABLE IF NOT EXISTS organisation_chatbots (
    id SERIAL PRIMARY KEY,
    organisation_id BIGINT UNSIGNED NOT NULL,
    chatbot_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE,
    UNIQUE KEY unique_org (organisation_id)
);

-- +goose Down
DROP TABLE IF EXISTS organisation_chatbots;
