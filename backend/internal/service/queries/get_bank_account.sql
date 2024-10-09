SELECT
    ba.id,
    ba.name,
    ba.amount,
    cur.id,
    cur.code,
    cur.description,
    cur.locale_code
FROM
    go_bank_accounts AS ba
INNER JOIN go_currencies cur ON ba.currency = cur.id
WHERE
    ba.id = ?
    AND ba.owner = ?