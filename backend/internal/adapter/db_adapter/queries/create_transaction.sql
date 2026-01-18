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
     vat_id,
     vat_included
    )
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, get_current_user_organisation_id(?), ?, ?)
