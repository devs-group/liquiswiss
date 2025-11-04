UPDATE users
SET current_scenario_id = ?
WHERE id = ?
  AND EXISTS (
    SELECT 1
    FROM scenarios
    WHERE scenarios.id = ?
      AND scenarios.organisation_id = users.current_organisation_id
);
