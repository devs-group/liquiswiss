INSERT INTO employees (name, organisation_id)
VALUES (?, get_current_user_organisation_id(?))