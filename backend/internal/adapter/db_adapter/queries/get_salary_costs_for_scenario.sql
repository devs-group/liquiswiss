-- Get salary costs for a scenario with inheritance support
-- Parameters: scenario_id, user_id
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
base_salary_costs AS (
    -- Get salary costs from horizontal scenarios
    SELECT
        sc.id,
        sc.uuid,
        sc.scenario_id,
        sc.salary_id,
        sc.cycle,
        sc.amount_type,
        sc.amount,
        sc.distribution_type,
        sc.relative_offset,
        sc.target_date,
        sc.label_id,
        sh.level
    FROM salary_costs sc
    INNER JOIN scenario_hierarchy sh ON sc.scenario_id = sh.id
    INNER JOIN salaries sal ON sc.salary_id = sal.id
    INNER JOIN employees e ON sal.employee_id = e.id
    WHERE sh.type = 'horizontal'
      AND e.organisation_id = get_current_user_organisation_id(?)
),
salary_cost_with_overrides AS (
    -- Combine base costs with overrides
    SELECT
        COALESCE(sco.id, bsc.id) as id,
        bsc.uuid,
        COALESCE(sco.scenario_id, bsc.scenario_id) as scenario_id,
        COALESCE(sco.salary_id, bsc.salary_id) as salary_id,
        COALESCE(sco.cycle, bsc.cycle) as cycle,
        COALESCE(sco.amount_type, bsc.amount_type) as amount_type,
        COALESCE(sco.amount, bsc.amount) as amount,
        COALESCE(sco.distribution_type, bsc.distribution_type) as distribution_type,
        COALESCE(sco.relative_offset, bsc.relative_offset) as relative_offset,
        COALESCE(sco.target_date, bsc.target_date) as target_date,
        COALESCE(sco.label_id, bsc.label_id) as label_id,
        COALESCE(sco_sh.level, bsc.level) as level
    FROM base_salary_costs bsc
    LEFT JOIN salary_cost_overrides sco ON sco.uuid = bsc.uuid
    LEFT JOIN scenario_hierarchy sco_sh ON sco.scenario_id = sco_sh.id AND sco_sh.type = 'vertical'
)
-- Return most specific version
SELECT DISTINCT ON (uuid)
    id,
    uuid,
    scenario_id,
    salary_id,
    cycle,
    amount_type,
    amount,
    distribution_type,
    relative_offset,
    target_date,
    label_id
FROM salary_cost_with_overrides
ORDER BY uuid, level ASC;
