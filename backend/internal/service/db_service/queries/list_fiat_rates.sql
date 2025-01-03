SELECT id, base, target, rate, updated_at
FROM fiat_rates
WHERE base = ?;