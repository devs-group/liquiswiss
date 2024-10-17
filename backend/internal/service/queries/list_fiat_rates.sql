SELECT id, base, target, rate, updated_at
FROM go_fiat_rates
WHERE base = ?;