DELETE FROM forecasts
WHERE scenario_id = get_current_user_scenario_id(?)