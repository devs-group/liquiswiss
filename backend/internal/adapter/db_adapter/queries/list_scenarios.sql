SELECT
    s.id,
    s.name,
    s.is_default,
    s.created_at,
    s.parent_scenario_id,
    s.organisation_id,
    COUNT(*) OVER () AS total_count
FROM scenarios AS s
WHERE s.organisation_id = get_current_user_organisation_id(?)
ORDER BY s.created_at DESC
LIMIT ?
OFFSET ?;

