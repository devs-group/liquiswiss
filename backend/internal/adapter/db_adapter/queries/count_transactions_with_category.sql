SELECT COUNT(*)
FROM transactions
WHERE category_id = ? AND organisation_id = get_current_user_organisation_id(?)
