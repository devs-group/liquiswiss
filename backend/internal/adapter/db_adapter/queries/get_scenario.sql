SELECT
    s.id,
    s.name,
    s.is_default,
    s.created_at,
    s.parent_scenario_id,
    s.organisation_id
FROM scenarios AS s
WHERE s.id = ?
  AND s.organisation_id = get_current_user_organisation_id(?);

