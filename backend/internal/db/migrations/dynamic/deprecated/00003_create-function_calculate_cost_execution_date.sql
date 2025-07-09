-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION calculate_cost_execution_date(
    type ENUM('single', 'repeating'),
    from_date DATE,
    to_date DATE,
    cycle ENUM('monthly', 'quarterly', 'biannually', 'yearly'),
    target_date DATE,
    cost_cycle ENUM('once', 'monthly', 'quarterly', 'biannually', 'yearly'),
    relative_offset INT,
    curr_date DATE,
    is_next BOOLEAN
)
    RETURNS DATE
    DETERMINISTIC
BEGIN
    DECLARE history_execution DATE;
    DECLARE previous_history_execution DATE;
    DECLARE last_possible_execution_date DATE;
    DECLARE cost_execution_date DATE;

    -- Calculate the relevant execution date for the history cycle
    SET history_execution = calculate_next_history_execution_date(
        type,
        from_date,
        to_date,
        cycle,
        curr_date
    );
    SET previous_history_execution = calculate_next_history_execution_date(
        type,
        from_date,
        to_date,
        cycle,
        DATE_SUB(curr_date, INTERVAL 1 MONTH)
    );

    -- Handle 'once' cycle case
    IF cost_cycle = 'once' THEN
        IF is_next THEN
            -- If today is after the target_date, 'once' costs should no longer execute
            IF target_date IS NOT NULL AND curr_date > target_date THEN
                RETURN NULL;
            END IF;
            RETURN target_date;
        ELSE
            -- For 'previous' on a one-time cycle, return NULL since there's no previous execution
            RETURN NULL;
        END IF;
    END IF;

    SET last_possible_execution_date = CASE cost_cycle
        WHEN 'monthly' THEN DATE_ADD(history_execution, INTERVAL relative_offset MONTH)
        WHEN 'quarterly' THEN DATE_ADD(history_execution, INTERVAL relative_offset * 3 MONTH)
        WHEN 'biannually' THEN DATE_ADD(history_execution, INTERVAL relative_offset * 6 MONTH)
        WHEN 'yearly' THEN DATE_ADD(history_execution, INTERVAL relative_offset YEAR)
    END;

    -- Handle target_date logic
    IF target_date IS NOT NULL THEN
        SET cost_execution_date = target_date;
        -- Forward logic: Handle next cost execution
        IF history_execution >= target_date THEN
            -- Incrementally calculate the next execution date
            CASE cost_cycle
                WHEN 'monthly' THEN
                    cost_loop: WHILE cost_execution_date <= curr_date DO
                        SET @next_temp_cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset MONTH);
                        IF @next_temp_cost_execution_date > last_possible_execution_date THEN
                            LEAVE cost_loop;
                        END IF;
                        SET cost_execution_date = @next_temp_cost_execution_date;
                    END WHILE;
                WHEN 'quarterly' THEN
                    cost_loop: WHILE cost_execution_date <= curr_date DO
                        SET @next_temp_cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset * 3 MONTH);
                        IF @next_temp_cost_execution_date > last_possible_execution_date THEN
                            LEAVE cost_loop;
                        END IF;
                        SET cost_execution_date = @next_temp_cost_execution_date;
                    END WHILE;
                WHEN 'biannually' THEN
                    cost_loop: WHILE cost_execution_date <= curr_date DO
                        SET @next_temp_cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset * 6 MONTH);
                        IF @next_temp_cost_execution_date > last_possible_execution_date THEN
                            LEAVE cost_loop;
                        END IF;
                        SET cost_execution_date = @next_temp_cost_execution_date;
                    END WHILE;
                WHEN 'yearly' THEN
                    cost_loop: WHILE cost_execution_date <= curr_date DO
                        SET @next_temp_cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset YEAR);
                        IF @next_temp_cost_execution_date > last_possible_execution_date THEN
                            LEAVE cost_loop;
                        END IF;
                        SET cost_execution_date = @next_temp_cost_execution_date;
                    END WHILE;
            END CASE;
        END IF;

        IF is_next THEN
            IF curr_date > cost_execution_date THEN
                RETURN null;
            ELSE
                RETURN cost_execution_date;
            END IF;
        END IF;

        -- Backward logic: Handle previous cost execution
        CASE cost_cycle
            WHEN 'monthly' THEN
                SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset MONTH);
            WHEN 'quarterly' THEN
                SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset * 3 MONTH);
            WHEN 'biannually' THEN
                SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset * 6 MONTH);
            WHEN 'yearly' THEN
                SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset YEAR);
        END CASE;

        RETURN cost_execution_date;
    END IF;

    -- Default case: Calculate based on cost_cycle and history_execution
    CASE cost_cycle
        WHEN 'monthly' THEN
            SET cost_execution_date = DATE_ADD(previous_history_execution, INTERVAL relative_offset MONTH);
        WHEN 'quarterly' THEN
            SET cost_execution_date = DATE_ADD(previous_history_execution, INTERVAL relative_offset * 3 MONTH);
        WHEN 'biannually' THEN
            SET cost_execution_date = DATE_ADD(previous_history_execution, INTERVAL relative_offset * 6 MONTH);
        WHEN 'yearly' THEN
            SET cost_execution_date = DATE_ADD(previous_history_execution, INTERVAL relative_offset YEAR);
    END CASE;

    IF is_next THEN
        IF curr_date > cost_execution_date THEN
            RETURN null;
#         ELSEIF curr_date < history_execution AND history_execution < cost_execution_date THEN
#             RETURN history_execution;
        ELSE
            RETURN cost_execution_date;
        END IF;
    END IF;

    -- The previous one becomes the last here
    IF curr_date > cost_execution_date THEN
        RETURN cost_execution_date;
    END IF;

    CASE cost_cycle
        WHEN 'monthly' THEN
            SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset MONTH);
        WHEN 'quarterly' THEN
            SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset * 3 MONTH);
        WHEN 'biannually' THEN
            SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset * 6 MONTH);
        WHEN 'yearly' THEN
            SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset YEAR);
    END CASE;

    IF is_next AND curr_date > cost_execution_date THEN
        return null;
    END IF;

    return cost_execution_date;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS calculate_cost_execution_date;
-- +goose StatementEnd
