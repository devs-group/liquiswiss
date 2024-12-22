-- +goose Up
-- +goose StatementBegin
ALTER TABLE employee_history
    CHANGE salary_per_month salary BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employee_history
    CHANGE salary salary_per_month BIGINT;
-- +goose StatementEnd
