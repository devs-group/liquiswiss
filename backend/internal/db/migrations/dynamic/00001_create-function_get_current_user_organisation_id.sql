-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION get_current_user_organisation_id(p_user_id BIGINT UNSIGNED)
    RETURNS BIGINT UNSIGNED
    DETERMINISTIC
BEGIN
    DECLARE v_organisation_id BIGINT UNSIGNED;
    DECLARE v_is_member INT;

    SELECT current_organisation_id
    INTO v_organisation_id
    FROM users
    WHERE id = p_user_id;

    IF v_organisation_id IS NULL THEN
        RETURN NULL;
    END IF;

    -- Return NULL if user is no longer a member of the organisation
    -- (e.g. removed by an owner/admin). Prevents a stale
    -- users.current_organisation_id from leaking data across org boundaries.
    SELECT COUNT(*) INTO v_is_member
    FROM users_2_organisations
    WHERE user_id = p_user_id
      AND organisation_id = v_organisation_id;

    IF v_is_member = 0 THEN
        RETURN NULL;
    END IF;

    RETURN v_organisation_id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_current_user_organisation_id;
-- +goose StatementEnd
