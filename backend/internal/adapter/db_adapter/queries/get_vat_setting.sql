SELECT
    vs.id,
    vs.organisation_id,
    vs.enabled,
    vs.billing_date,
    vs.transaction_date,
    vs.interval,
    vs.created_at,
    vs.updated_at
FROM
    vat_settings AS vs
WHERE
    vs.organisation_id = get_current_user_organisation_id(?)
