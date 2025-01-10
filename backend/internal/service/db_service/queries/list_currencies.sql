SELECT
    c.id,
    c.code,
    c.description,
    c.locale_code
FROM
    currencies AS c
ORDER BY
    -- Sort the organisations base currency always on top
    IF(get_current_user_organisation_currency_id(get_current_user_organisation_id(?)) = c.id, 0, 1),
    code
