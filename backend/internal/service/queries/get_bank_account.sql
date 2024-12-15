SELECT
    ba.id,
    ba.name,
    ba.amount,
    cur.id,
    cur.code,
    cur.description,
    cur.locale_code
FROM
    bank_accounts AS ba
INNER JOIN currencies cur ON ba.currency_id = cur.id
WHERE
    ba.id = ?
    AND ba.organisation_id = (SELECT current_organisation FROM users u WHERE u.id = ?)