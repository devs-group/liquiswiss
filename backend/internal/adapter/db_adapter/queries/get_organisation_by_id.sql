SELECT
    o.id,
    o.name,
    c.id,
    c.code,
    c.description,
    c.locale_code,
    member_counts.member_count AS member_count
FROM organisations o
LEFT JOIN currencies c ON c.id = get_current_user_organisation_currency_id(o.id)
LEFT JOIN (
    SELECT organisation_id, COUNT(*) AS member_count
    FROM users_2_organisations
    GROUP BY organisation_id
) AS member_counts ON member_counts.organisation_id = o.id
WHERE o.id = ?
