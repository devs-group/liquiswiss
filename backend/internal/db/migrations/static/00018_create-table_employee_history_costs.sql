-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employee_history_costs (
    id SERIAL PRIMARY KEY,
    cycle ENUM('once', 'monthly', 'quarterly', 'biannually', 'yearly') NOT NULL DEFAULT 'monthly',
    amount_type ENUM('fixed', 'percentage') NOT NULL DEFAULT 'fixed',
    amount BIGINT UNSIGNED NOT NULL DEFAULT 0,
    distribution_type ENUM('employer', 'employee') NOT NULL DEFAULT 'employee',
    relative_offset BIGINT NOT NULL DEFAULT 1,
    target_date DATE,

    label_id BIGINT UNSIGNED,
    employee_history_id BIGINT UNSIGNED NOT NULL,

    CONSTRAINT FK_Cost_Label FOREIGN KEY (label_id) REFERENCES employee_history_cost_labels (id) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT FK_Cost_EmployeeHistory FOREIGN KEY (employee_history_id) REFERENCES employee_histories (id) ON DELETE CASCADE ON UPDATE CASCADE,

    CONSTRAINT CK_Amount_Type CHECK (
        (amount_type = 'fixed' AND amount >= 0) OR
        -- Percentage can be only between 0 and 100 with a precision of 3 decimals
        (amount_type = 'percentage' AND amount >= 0 AND amount <= 100000)
    ),
    CONSTRAINT CK_Target_Date CHECK (
        (cycle = 'once' AND target_date IS NOT NULL) OR
        (cycle != 'once')
    ),
    CONSTRAINT CK_Relative_Offset CHECK (
        relative_offset > 0
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employee_history_costs;
-- +goose StatementEnd
