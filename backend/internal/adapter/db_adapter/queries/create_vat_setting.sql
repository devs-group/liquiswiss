INSERT INTO vat_settings (
    organisation_id,
    enabled,
    billing_date,
    transaction_month_offset,
    `interval`
) VALUES (
    get_current_user_organisation_id(?),
    ?,
    ?,
    ?,
    ?
)
