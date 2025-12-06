SELECT
    sc.id,
    sc.name,
    sc.is_default,
    sc.created_at,
    sc.parent_scenario_id
FROM scenarios AS sc
WHERE
    sc.organisation_id = get_current_user_organisation_id(?)