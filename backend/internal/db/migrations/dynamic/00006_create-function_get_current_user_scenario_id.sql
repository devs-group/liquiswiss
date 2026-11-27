-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION get_current_user_scenario_id(p_user_id BIGINT UNSIGNED)
    RETURNS BIGINT UNSIGNED
    DETERMINISTIC
BEGIN
    DECLARE v_organisation_id BIGINT UNSIGNED;
    DECLARE v_scenario_id BIGINT UNSIGNED DEFAULT NULL;

    -- Already NULL for users who are no longer a member of their current organisation
    SET v_organisation_id = get_current_user_organisation_id(p_user_id);
    IF v_organisation_id IS NULL THEN
        RETURN NULL;
    END IF;

    -- The stored choice only counts while it still belongs to the current organisation
    SELECT s.id
    INTO v_scenario_id
    FROM user_organisation_settings uos
    JOIN scenarios s ON s.id = uos.current_scenario_id AND s.organisation_id = v_organisation_id
    WHERE uos.user_id = p_user_id
      AND uos.organisation_id = v_organisation_id
    LIMIT 1;

    IF v_scenario_id IS NULL THEN
        SELECT s.id
        INTO v_scenario_id
        FROM scenarios s
        WHERE s.organisation_id = v_organisation_id
          AND s.is_base
        LIMIT 1;
    END IF;

    RETURN v_scenario_id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_current_user_scenario_id;
-- +goose StatementEnd
