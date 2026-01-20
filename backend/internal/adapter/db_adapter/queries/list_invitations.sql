SELECT
    i.id,
    i.organisation_id,
    i.email,
    i.role,
    i.token,
    i.invited_by,
    u.name AS invited_by_name,
    i.expires_at,
    i.created_at
FROM organisation_invitations i
INNER JOIN users u ON u.id = i.invited_by
WHERE i.organisation_id = ?
ORDER BY i.created_at DESC