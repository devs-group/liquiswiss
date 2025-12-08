UPDATE users
SET
  current_organisation_id = ?,
  current_scenario_id = (
    SELECT scenario_id
    FROM users_2_scenarios
    WHERE user_id = ?
      AND organisation_id = ?
      AND is_current = TRUE
    LIMIT 1
  )
WHERE id = ?
  AND EXISTS (
    SELECT 1
    FROM users_2_organisations
    WHERE user_id = users.id
      AND organisation_id = ?
  );