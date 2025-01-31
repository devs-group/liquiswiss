DELETE te
FROM transaction_exclusions te
JOIN transactions AS t ON t.id = te.transaction_id
WHERE te.exclude_month = ?
  AND te.transaction_id = ?
  AND t.organisation_id = get_current_user_organisation_id(?)