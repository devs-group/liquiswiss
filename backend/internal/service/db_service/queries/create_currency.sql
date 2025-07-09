INSERT INTO currencies (code, description, locale_code)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE description = VALUES(description),
                        locale_code = VALUES(locale_code),
                        code        = VALUES(code);