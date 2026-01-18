SELECT
    us.id,
    us.user_id,
    us.settings_tab,
    us.skip_organisation_switch_question,
    us.created_at,
    us.updated_at
FROM
    user_settings AS us
WHERE
    us.user_id = ?
