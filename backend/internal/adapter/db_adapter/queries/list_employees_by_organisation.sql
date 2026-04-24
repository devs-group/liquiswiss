SELECT
    e.id,
    e.name
FROM employees e
WHERE e.organisation_id = ?
ORDER BY e.name ASC
