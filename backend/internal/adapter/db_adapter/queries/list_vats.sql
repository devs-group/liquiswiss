SELECT
    v.id,
    v.value,
    CONCAT(FORMAT(v.value / 100, IF(v.value % 10 = 0, 1, 2)), '%') AS formatted_value,
    IF(v.organisation_id IS NULL, false, true) AS can_edit
FROM vats AS v
WHERE organisation_id IS NULL
   OR v.organisation_id = get_current_user_organisation_id(?)
ORDER BY v.value