INSERT INTO user_organisation_settings (user_id, organisation_id, current_scenario_id)
SELECT ?, s.organisation_id, s.id
FROM scenarios s
WHERE s.id = ?
  AND s.organisation_id = get_current_user_organisation_id(?)
ON DUPLICATE KEY UPDATE current_scenario_id = VALUES(current_scenario_id)
