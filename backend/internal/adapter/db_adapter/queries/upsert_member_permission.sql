INSERT INTO member_permissions (user_id, organisation_id, entity_type, can_view, can_edit, can_delete)
VALUES (?, ?, NULL, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    can_view = VALUES(can_view),
    can_edit = VALUES(can_edit),
    can_delete = VALUES(can_delete)