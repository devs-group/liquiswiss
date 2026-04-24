SELECT
    i.id,
    i.organisation_id,
    o.name AS organisation_name,
    i.email,
    i.role,
    i.token,
    i.invited_by,
    u.name AS invited_by_name,
    i.expires_at,
    i.created_at
FROM organisation_invitations i
INNER JOIN users u ON u.id = i.invited_by
INNER JOIN organisations o ON o.id = i.organisation_id
WHERE i.email = ?
  AND i.expires_at > NOW()
ORDER BY i.created_at DESC
