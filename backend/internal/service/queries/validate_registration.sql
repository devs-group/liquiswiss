SELECT
    id
FROM
    registrations
WHERE
    email = ?
  AND code = ?
  AND created_at >= NOW() - INTERVAL ? HOUR;