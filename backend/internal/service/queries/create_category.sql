INSERT INTO categories (name, organisation_id)
VALUES (?, (SELECT current_organisation FROM users u WHERE u.id = ?))