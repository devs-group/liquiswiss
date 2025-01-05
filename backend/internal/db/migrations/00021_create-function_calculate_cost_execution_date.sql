-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION calculate_cost_execution_date(
    type ENUM('single', 'repeating'),
    from_date DATE,
    to_date DATE,
    cycle ENUM('daily', 'weekly', 'monthly', 'quarterly', 'biannually', 'yearly'),
    target_date DATE,
    cost_cycle ENUM('once', 'daily', 'weekly', 'monthly', 'quarterly', 'biannually', 'yearly'),
    relative_offset INT,
    curr_date DATE,
    is_next BOOLEAN
)
    RETURNS DATE
    DETERMINISTIC
BEGIN
    DECLARE history_execution DATE;
    DECLARE cost_execution_date DATE;

    -- Calculate the relevant execution date for the history cycle
    SET history_execution = calculate_next_history_execution_date(
        type,
        from_date,
        to_date,
        cycle,
        curr_date
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

    -- Handle target_date logic
    IF target_date IS NOT NULL THEN
        SET cost_execution_date = target_date;
        -- Forward logic: Handle next cost execution
        IF history_execution >= target_date THEN
            SET cost_execution_date = target_date;
            -- Incrementally calculate the next execution date
            CASE cost_cycle
                WHEN 'daily' THEN
                    WHILE cost_execution_date <= curr_date DO
                        SET cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset DAY);
                    END WHILE;
                WHEN 'weekly' THEN
                    WHILE cost_execution_date <= curr_date DO
                        SET cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset WEEK);
                    END WHILE;
                WHEN 'monthly' THEN
                    WHILE cost_execution_date <= curr_date DO
                        SET cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset MONTH);
                    END WHILE;
                WHEN 'quarterly' THEN
                    WHILE cost_execution_date <= curr_date DO
                        SET cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset * 3 MONTH);
                    END WHILE;
                WHEN 'biannually' THEN
                    WHILE cost_execution_date <= curr_date DO
                        SET cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset * 6 MONTH);
                    END WHILE;
                WHEN 'yearly' THEN
                    WHILE cost_execution_date <= curr_date DO
                        SET cost_execution_date = DATE_ADD(cost_execution_date, INTERVAL relative_offset YEAR);
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
            WHEN 'daily' THEN
                SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset DAY);
            WHEN 'weekly' THEN
                SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset WEEK);
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
        WHEN 'daily' THEN
            SET cost_execution_date = DATE_ADD(history_execution, INTERVAL relative_offset DAY);
        WHEN 'weekly' THEN
            SET cost_execution_date = DATE_ADD(history_execution, INTERVAL relative_offset WEEK);
        WHEN 'monthly' THEN
            SET cost_execution_date = DATE_ADD(history_execution, INTERVAL relative_offset MONTH);
        WHEN 'quarterly' THEN
            SET cost_execution_date = DATE_ADD(history_execution, INTERVAL relative_offset * 3 MONTH);
        WHEN 'biannually' THEN
            SET cost_execution_date = DATE_ADD(history_execution, INTERVAL relative_offset * 6 MONTH);
        WHEN 'yearly' THEN
            SET cost_execution_date = DATE_ADD(history_execution, INTERVAL relative_offset YEAR);
    END CASE;
    IF is_next THEN
        IF curr_date > cost_execution_date THEN
            RETURN null;
        ELSE
            RETURN cost_execution_date;
        END IF;
    END IF;
    CASE cost_cycle
        WHEN 'daily' THEN
            SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset DAY);
        WHEN 'weekly' THEN
            SET cost_execution_date = DATE_SUB(cost_execution_date, INTERVAL relative_offset WEEK);
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
