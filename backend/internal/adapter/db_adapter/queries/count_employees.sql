SELECT
    COUNT(*) OVER () AS total_count
FROM
    employees AS e
WHERE
    e.scenario_id = get_current_user_scenario_id(?)
LIMIT ?
OFFSET ?
