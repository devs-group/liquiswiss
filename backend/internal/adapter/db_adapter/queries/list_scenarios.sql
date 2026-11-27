SELECT
    s.id,
    s.name,
    s.is_base,
    s.parent_id,
    s.created_at,
    COALESCE(s.id = get_current_user_scenario_id(?), false) AS is_current
FROM scenarios s
WHERE s.organisation_id = get_current_user_organisation_id(?)
ORDER BY s.is_base DESC, s.name ASC, s.id ASC
