SELECT
    id
FROM
    reset_password
WHERE
    email = ?
  AND code = ?
  AND created_at >= NOW() - INTERVAL ? HOUR;