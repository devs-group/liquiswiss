SELECT
    ba.id,
    ba.name,
    ba.amount,
    cur.id,
    cur.code,
    cur.description,
    cur.locale_code
FROM bank_accounts AS ba
INNER JOIN currencies cur ON ba.currency_id = cur.id
INNER JOIN scenarios s ON s.id = ba.scenario_id
WHERE
    ba.scenario_id = ?
    AND s.organisation_id = get_current_user_organisation_id(?)
ORDER BY ba.name ASC
