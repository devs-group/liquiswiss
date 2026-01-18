-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_organisation_settings (
    id SERIAL PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    organisation_id BIGINT UNSIGNED NOT NULL,
    forecast_months INT NOT NULL DEFAULT 13,
    forecast_performance INT NOT NULL DEFAULT 100,
    forecast_revenue_details BOOL NOT NULL DEFAULT false,
    forecast_expense_details BOOL NOT NULL DEFAULT false,
    forecast_child_details JSON NOT NULL DEFAULT '[]',
    employee_display ENUM('grid', 'list') NOT NULL DEFAULT 'grid',
    employee_sort_by VARCHAR(50) NOT NULL DEFAULT 'name',
    employee_sort_order ENUM('ASC', 'DESC') NOT NULL DEFAULT 'ASC',
    employee_hide_terminated BOOL NOT NULL DEFAULT true,
    transaction_display ENUM('grid', 'list') NOT NULL DEFAULT 'grid',
    transaction_sort_by VARCHAR(50) NOT NULL DEFAULT 'name',
    transaction_sort_order ENUM('ASC', 'DESC') NOT NULL DEFAULT 'ASC',
    transaction_hide_disabled BOOL NOT NULL DEFAULT true,
    bank_account_display ENUM('grid', 'list') NOT NULL DEFAULT 'grid',
    bank_account_sort_by VARCHAR(50) NOT NULL DEFAULT 'name',
    bank_account_sort_order ENUM('ASC', 'DESC') NOT NULL DEFAULT 'ASC',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_organisation_setting (user_id, organisation_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_organisation_settings;
-- +goose StatementEnd
