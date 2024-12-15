INSERT INTO bank_accounts (name, amount, currency_id, organisation_id)
SELECT ?, ?, ?, (SELECT current_organisation FROM users u WHERE u.id = ?)