SELECT
    u.id,
    u.name,
    u.email,
    u.current_organisation_id,
    c.id AS currency_id,
    c.code AS currency_code,
    c.description AS currency_description,
    c.locale_code AS currency_locale_code
FROM users u
JOIN currencies c ON c.id = get_current_user_organisation_currency_id(get_current_user_organisation_id(u.id))
WHERE u.id = ?