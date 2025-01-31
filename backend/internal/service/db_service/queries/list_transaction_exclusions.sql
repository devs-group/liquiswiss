SELECT
    te.id,
    te.exclude_month,
    te.transaction_id
FROM transaction_exclusions te
JOIN transactions AS t ON t.id = te.transaction_id
WHERE te.transaction_id = ?
  AND t.organisation_id = get_current_user_organisation_id(?)