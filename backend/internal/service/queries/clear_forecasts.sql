DELETE FROM forecasts
WHERE organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)