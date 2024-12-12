SELECT
    v.id,
    v.value,
    CONCAT(FORMAT(v.value / 100, IF(v.value % 10 = 0, 1, 2)), '%') AS formatted_value,
    IF(v.owner IS NULL, false, true) AS can_edit
FROM vats AS v
WHERE owner IS NULL OR v.owner = ?
ORDER BY v.value