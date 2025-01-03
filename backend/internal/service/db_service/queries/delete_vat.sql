DELETE FROM vats
WHERE
    id = ?
    AND organisation_id = get_current_organisation(?)
