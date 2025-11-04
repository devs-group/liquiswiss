-- Get salaries for a scenario with inheritance support
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
base_salaries AS (
    -- Get salaries from horizontal scenarios
    SELECT
        sal.id,
        sal.uuid,
        sal.scenario_id,
        sal.employee_id,
        sal.cycle,
        sal.hours_per_month,
        sal.amount,
        sal.vacation_days_per_year,
        sal.from_date,
        sal.to_date,
        sal.is_termination,
        sal.is_disabled,
        sal.deleted,
        sal.currency_id,
        sal.created_at,
        sh.level
    FROM salaries sal
    INNER JOIN scenario_hierarchy sh ON sal.scenario_id = sh.id
    INNER JOIN employees e ON sal.employee_id = e.id
    WHERE sh.type = 'horizontal'
      AND e.organisation_id = get_current_user_organisation_id(?)
),
salary_with_overrides AS (
    -- Combine base salaries with overrides
    SELECT
        COALESCE(so.id, bs.id) as id,
        bs.uuid,
        COALESCE(so.scenario_id, bs.scenario_id) as scenario_id,
        COALESCE(so.employee_id, bs.employee_id) as employee_id,
        COALESCE(so.cycle, bs.cycle) as cycle,
        COALESCE(so.hours_per_month, bs.hours_per_month) as hours_per_month,
        COALESCE(so.amount, bs.amount) as amount,
        COALESCE(so.vacation_days_per_year, bs.vacation_days_per_year) as vacation_days_per_year,
        COALESCE(so.from_date, bs.from_date) as from_date,
        COALESCE(so.to_date, bs.to_date) as to_date,
        COALESCE(so.is_termination, bs.is_termination) as is_termination,
        COALESCE(so.is_disabled, bs.is_disabled) as is_disabled,
        COALESCE(so.deleted, bs.deleted) as deleted,
        COALESCE(so.currency_id, bs.currency_id) as currency_id,
        COALESCE(so.created_at, bs.created_at) as created_at,
        COALESCE(so_sh.level, bs.level) as level
    FROM base_salaries bs
    LEFT JOIN salary_overrides so ON so.uuid = bs.uuid
    LEFT JOIN scenario_hierarchy so_sh ON so.scenario_id = so_sh.id AND so_sh.type = 'vertical'
)
-- Return most specific version
SELECT DISTINCT ON (uuid)
    id,
    uuid,
    scenario_id,
    employee_id,
    cycle,
    hours_per_month,
    amount,
    vacation_days_per_year,
    from_date,
    to_date,
    is_termination,
    is_disabled,
    deleted,
    currency_id,
    created_at
FROM salary_with_overrides
WHERE deleted = false OR deleted IS NULL
ORDER BY uuid, level ASC;
