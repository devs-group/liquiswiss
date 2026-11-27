DELETE FROM scenarios
WHERE id = ?
  AND NOT is_base
  AND organisation_id = get_current_user_organisation_id(?)
