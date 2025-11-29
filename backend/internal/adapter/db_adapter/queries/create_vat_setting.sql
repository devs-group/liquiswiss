INSERT INTO vat_settings (
    organisation_id,
    enabled,
    billing_date,
    transaction_date,
    `interval`
) VALUES (
    get_current_user_organisation_id(?),
    ?,
    ?,
    ?,
    ?
)
