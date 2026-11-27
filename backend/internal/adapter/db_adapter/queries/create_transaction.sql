INSERT INTO transactions
    (
     name,
     link,
     amount,
     cycle,
     type,
     start_date,
     end_date,
     category_id,
     currency_id,
     employee_id,
     organisation_id,
     scenario_id,
     vat_id,
     vat_included
    )
SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, s.organisation_id, s.id, ?, ?
FROM scenarios s
WHERE s.id = get_current_user_scenario_id(?)
