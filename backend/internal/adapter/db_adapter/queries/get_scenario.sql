SELECT
    s.id,
    s.name,
    s.is_base,
    s.parent_id,
    s.created_at,
    COALESCE(s.id = get_current_user_scenario_id(?), false) AS is_current
FROM scenarios s
WHERE s.id = ?
  AND s.organisation_id = get_current_user_organisation_id(?)
