INSERT INTO vats (value, organisation_id)
VALUES (?, (SELECT current_organisation FROM users u WHERE u.id = ?))