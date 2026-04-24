SELECT role
FROM users_2_organisations
WHERE user_id = ?
  AND organisation_id = get_current_user_organisation_id(?)
