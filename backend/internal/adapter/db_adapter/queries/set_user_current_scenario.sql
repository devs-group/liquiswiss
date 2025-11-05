UPDATE users
SET current_scenario_id = ?
WHERE id = ?
  AND EXISTS (
    SELECT 1
    FROM scenarios
    WHERE id = ?
      AND organisation_id = users.current_organisation_id
);
