-- +goose Up
-- +goose StatementBegin
-- This view will be used as a template - actual queries will use CTEs with parameters
-- But we can create a helper function for common scenario operations
CREATE FUNCTION get_scenario_hierarchy_with_types(target_scenario_id BIGINT UNSIGNED)
RETURNS JSON
DETERMINISTIC
READS SQL DATA
BEGIN
    DECLARE result JSON;

    -- Build hierarchy path with types
    WITH RECURSIVE scenario_path AS (
        SELECT
            id,
            type,
            parent_scenario_id,
            0 as level,
            CAST(id AS CHAR(255)) as path
        FROM scenarios
        WHERE id = target_scenario_id

        UNION ALL

        SELECT
            s.id,
            s.type,
            s.parent_scenario_id,
            sp.level + 1,
            CONCAT(sp.path, ',', s.id)
        FROM scenarios s
        INNER JOIN scenario_path sp ON s.id = sp.parent_scenario_id
    )
    SELECT JSON_ARRAYAGG(
        JSON_OBJECT(
            'id', id,
            'type', type,
            'level', level
        )
    ) INTO result
    FROM scenario_path
    ORDER BY level ASC;

    RETURN result;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_scenario_hierarchy_with_types;
-- +goose StatementEnd
