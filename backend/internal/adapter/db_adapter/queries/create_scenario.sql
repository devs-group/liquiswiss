INSERT INTO scenarios (name, parent_scenario_id, is_default, organisation_id)
SELECT ?, ?, ?, get_current_user_organisation_id(?)