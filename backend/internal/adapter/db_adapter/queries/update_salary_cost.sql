UPDATE salary_costs AS hc
SET
    cycle = ?,
    amount_type = ?,
    amount = ?,
    distribution_type = ?,
    relative_offset = ?,
    target_date = ?,
    label_id = ?
WHERE hc.id = ?
    AND EXISTS (
        SELECT 1
        FROM employees AS e
        WHERE e.organisation_id = get_current_user_organisation_id(?)
    )
LIMIT 1;
