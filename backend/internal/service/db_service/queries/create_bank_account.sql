INSERT INTO bank_accounts (name, amount, currency_id, organisation_id)
SELECT ?, ?, ?, get_current_organisation(?)