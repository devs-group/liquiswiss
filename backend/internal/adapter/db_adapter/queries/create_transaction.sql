INSERT INTO transactions
    (
     name,
     amount,
     cycle,
     type,
     start_date,
     end_date,
     category_id,
     currency_id,
     employee_id,
     organisation_id,
     vat_id,
     vat_included,
     uuid,
     scenario_id
    )
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, get_current_user_organisation_id(?), ?, ?, UUID(), (
    SELECT current_scenario_id FROM users WHERE id = ?
))
