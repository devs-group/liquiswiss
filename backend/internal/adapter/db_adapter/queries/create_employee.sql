INSERT INTO employees (name, organisation_id, uuid, scenario_id)
VALUES (?, get_current_user_organisation_id(?), UUID(), (
    SELECT current_scenario_id FROM users WHERE id = ?
))