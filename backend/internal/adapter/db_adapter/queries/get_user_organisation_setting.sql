SELECT
    uos.id,
    uos.user_id,
    uos.organisation_id,
    uos.forecast_months,
    uos.forecast_performance,
    uos.forecast_revenue_details,
    uos.forecast_expense_details,
    uos.forecast_child_details,
    uos.employee_display,
    uos.employee_sort_by,
    uos.employee_sort_order,
    uos.employee_hide_terminated,
    uos.transaction_display,
    uos.transaction_sort_by,
    uos.transaction_sort_order,
    uos.transaction_hide_disabled,
    uos.bank_account_display,
    uos.bank_account_sort_by,
    uos.bank_account_sort_order,
    uos.created_at,
    uos.updated_at
FROM
    user_organisation_settings AS uos
WHERE
    uos.user_id = ?
    AND uos.organisation_id = get_current_user_organisation_id(?)
