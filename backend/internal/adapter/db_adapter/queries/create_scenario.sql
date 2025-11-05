INSERT INTO scenarios (
    name,
    is_default,
    parent_scenario_id,
    organisation_id
)
VALUES (
    ?,
    ?,
    ?,
    get_current_user_organisation_id(?)
);

