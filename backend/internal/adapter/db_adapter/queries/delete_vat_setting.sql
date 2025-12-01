DELETE FROM vat_settings
WHERE organisation_id = get_current_user_organisation_id(?)
