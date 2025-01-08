DELETE FROM forecasts
WHERE organisation_id = get_current_user_organisation_id(?)