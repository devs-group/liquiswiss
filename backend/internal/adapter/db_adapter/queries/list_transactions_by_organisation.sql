SELECT
    r.id,
    r.name,
    r.amount,
    r.cycle,
    r.type,
    r.is_disabled,
    cur.code
FROM transactions AS r
INNER JOIN currencies cur ON r.currency_id = cur.id
WHERE r.organisation_id = ?
ORDER BY r.name ASC
