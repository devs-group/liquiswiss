INSERT INTO employees (name, organisation_id, scenario_id)
SELECT ?, s.organisation_id, s.id
FROM scenarios s
WHERE s.id = get_current_user_scenario_id(?)