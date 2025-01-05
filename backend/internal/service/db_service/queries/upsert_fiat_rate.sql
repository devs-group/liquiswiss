INSERT INTO fiat_rates (base, target, rate)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE
    rate = VALUES(rate)