UPDATE users
SET current_organisation = ?
WHERE id = ?
  AND EXISTS (
    SELECT 1
    FROM users_2_organisations
    WHERE user_id = users.id
      AND organisation_id = ?
);