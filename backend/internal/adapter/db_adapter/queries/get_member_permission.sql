SELECT
    id,
    user_id,
    organisation_id,
    entity_type,
    can_view,
    can_edit,
    can_delete,
    created_at,
    updated_at
FROM member_permissions
WHERE user_id = ? AND organisation_id = ? AND entity_type IS NULL