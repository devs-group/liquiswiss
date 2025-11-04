-- Get a single employee by ID for a scenario with inheritance support
-- Parameters: employee_id, scenario_id, user_id
WITH RECURSIVE scenario_hierarchy AS (
    SELECT
        id,
        type,
        parent_scenario_id,
        0 as level
    FROM scenarios
    WHERE id = ?

    UNION ALL

    SELECT
        s.id,
        s.type,
        s.parent_scenario_id,
        sh.level + 1
    FROM scenarios s
    INNER JOIN scenario_hierarchy sh ON s.id = sh.parent_scenario_id
),
base_employee AS (
    -- Get employee from horizontal scenarios
    SELECT
        e.id,
        e.uuid,
        e.scenario_id,
        e.name,
        e.deleted,
        e.organisation_id,
        e.created_at,
        sh.level
    FROM employees e
    INNER JOIN scenario_hierarchy sh ON e.scenario_id = sh.id
    WHERE sh.type = 'horizontal'
      AND e.id = ?
      AND e.organisation_id = get_current_user_organisation_id(?)
    LIMIT 1
),
employee_with_override AS (
    -- Apply override if exists
    SELECT
        COALESCE(eo.id, be.id) as id,
        be.uuid,
        COALESCE(eo.scenario_id, be.scenario_id) as scenario_id,
        COALESCE(eo.name, be.name) as name,
        COALESCE(eo.deleted, be.deleted) as deleted,
        COALESCE(eo.organisation_id, be.organisation_id) as organisation_id,
        COALESCE(eo.created_at, be.created_at) as created_at,
        COALESCE(eo_sh.level, be.level) as level
    FROM base_employee be
    LEFT JOIN employee_overrides eo ON eo.uuid = be.uuid
    LEFT JOIN scenario_hierarchy eo_sh ON eo.scenario_id = eo_sh.id AND eo_sh.type = 'vertical'
)
SELECT
    id,
    uuid,
    scenario_id,
    name,
    deleted,
    organisation_id,
    created_at
FROM employee_with_override
ORDER BY level ASC
LIMIT 1;
