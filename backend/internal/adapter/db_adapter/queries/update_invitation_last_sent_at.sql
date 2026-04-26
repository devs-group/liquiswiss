UPDATE organisation_invitations
SET last_sent_at = NOW()
WHERE id = ? AND organisation_id = ?
