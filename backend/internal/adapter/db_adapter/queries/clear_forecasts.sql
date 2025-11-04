DELETE FROM forecasts
WHERE organisation_id = get_current_user_organisation_id(?)
  AND scenario_id = (SELECT current_scenario_id FROM users WHERE id = ?)