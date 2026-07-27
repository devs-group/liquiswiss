UPDATE transactions
SET category_id = ?
WHERE category_id = ? AND organisation_id = get_current_user_organisation_id(?)
