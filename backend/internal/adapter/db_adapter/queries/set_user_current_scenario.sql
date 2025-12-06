UPDATE users_2_scenarios u2o
SET u2o.scenario_id = ?
WHERE u2o.user_id = ?
AND u2o.organisation_id = get_current_user_organisation_id(?)
AND EXISTS (
    SELECT 1
    FROM scenarios sc
    WHERE sc.id = ? AND
        sc.organisation_id = get_current_user_organisation_id(?)
);