DELETE FROM forecasts
WHERE organisation_id = get_current_organisation(?)