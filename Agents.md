In [static](backend/internal/db/migrations/static) add only one up and down statement per goose block, like this:

-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions
ADD COLUMN is_disabled TINYINT(1) NOT NULL DEFAULT 0 AFTER vat_included;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE salaries
DROP COLUMN is_disabled;
-- +goose StatementEnd

not like this:

-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions
ADD COLUMN is_disabled TINYINT(1) NOT NULL DEFAULT 0 AFTER vat_included;

ALTER TABLE salaries
ADD COLUMN is_disabled TINYINT(1) NOT NULL DEFAULT 0 AFTER is_termination;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE salaries
DROP COLUMN is_disabled;

ALTER TABLE transactions
DROP COLUMN is_disabled;
-- +goose StatementEnd

Run "go test ./..." with privilege, if not possible ask for privilege

Generate with "go generate ./..." if not possible ask for privilege

Add test in backend whenever we have a new feature or change something in the backend

Do not necessarily refactor unrelated code if it's not required