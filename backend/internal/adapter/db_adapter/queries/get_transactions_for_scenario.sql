-- Get transactions for a scenario with inheritance support
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
base_transactions AS (
    -- Get transactions from horizontal scenarios
    SELECT
        t.id,
        t.uuid,
        t.scenario_id,
        t.name,
        t.amount,
        t.vat_included,
        t.is_disabled,
        t.cycle,
        t.type,
        t.start_date,
        t.end_date,
        t.deleted,
        t.vat_id,
        t.category_id,
        t.employee_id,
        t.currency_id,
        t.organisation_id,
        t.created_at,
        sh.level
    FROM transactions t
    INNER JOIN scenario_hierarchy sh ON t.scenario_id = sh.id
    WHERE sh.type = 'horizontal'
      AND t.organisation_id = get_current_user_organisation_id(?)
),
transaction_with_overrides AS (
    -- Combine base transactions with overrides
    SELECT
        COALESCE(tro.id, bt.id) as id,
        bt.uuid,
        COALESCE(tro.scenario_id, bt.scenario_id) as scenario_id,
        COALESCE(tro.name, bt.name) as name,
        COALESCE(tro.amount, bt.amount) as amount,
        COALESCE(tro.vat_included, bt.vat_included) as vat_included,
        COALESCE(tro.is_disabled, bt.is_disabled) as is_disabled,
        COALESCE(tro.cycle, bt.cycle) as cycle,
        COALESCE(tro.type, bt.type) as type,
        COALESCE(tro.start_date, bt.start_date) as start_date,
        COALESCE(tro.end_date, bt.end_date) as end_date,
        COALESCE(tro.deleted, bt.deleted) as deleted,
        COALESCE(tro.vat_id, bt.vat_id) as vat_id,
        COALESCE(tro.category_id, bt.category_id) as category_id,
        COALESCE(tro.employee_id, bt.employee_id) as employee_id,
        COALESCE(tro.currency_id, bt.currency_id) as currency_id,
        COALESCE(tro.organisation_id, bt.organisation_id) as organisation_id,
        COALESCE(tro.created_at, bt.created_at) as created_at,
        COALESCE(tro_sh.level, bt.level) as level
    FROM base_transactions bt
    LEFT JOIN transaction_overrides tro ON tro.uuid = bt.uuid
    LEFT JOIN scenario_hierarchy tro_sh ON tro.scenario_id = tro_sh.id AND tro_sh.type = 'vertical'
)
-- Return most specific version
SELECT DISTINCT ON (uuid)
    id,
    uuid,
    scenario_id,
    name,
    amount,
    vat_included,
    is_disabled,
    cycle,
    type,
    start_date,
    end_date,
    deleted,
    vat_id,
    category_id,
    employee_id,
    currency_id,
    organisation_id,
    created_at
FROM transaction_with_overrides
WHERE deleted = false OR deleted IS NULL
ORDER BY uuid, level ASC;
