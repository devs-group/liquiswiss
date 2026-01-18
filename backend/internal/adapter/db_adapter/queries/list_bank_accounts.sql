SELECT
    ba.id,
    ba.name,
    ba.amount,
    cur.id,
    cur.code,
    cur.description,
    cur.locale_code,
    COUNT(*) OVER () AS total_count
FROM bank_accounts AS ba
INNER JOIN currencies cur ON ba.currency_id = cur.id
WHERE
    ba.organisation_id = get_current_user_organisation_id(?)
    {{if .hasSearch}}AND LOWER(ba.name) LIKE LOWER(?){{end}}
ORDER BY
    {{.sortBy}} {{.sortOrder}},
    ba.name ASC
LIMIT ? OFFSET ?
