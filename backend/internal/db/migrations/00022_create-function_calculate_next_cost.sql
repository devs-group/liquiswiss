-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION calculate_next_cost_amount(
    type ENUM('single', 'repeating'),
    from_date DATE,
    to_date DATE,
    cycle ENUM('daily', 'weekly', 'monthly', 'quarterly', 'biannually', 'yearly'),
    target_date DATE,
    cost_cycle ENUM('once', 'daily', 'weekly', 'monthly', 'quarterly', 'biannually', 'yearly'),
    relative_offset INT,
    curr_date DATE,
    amount_type ENUM('fixed', 'percentage'),
    amount BIGINT UNSIGNED,
    salary BIGINT
)
    RETURNS BIGINT
    DETERMINISTIC
BEGIN
    DECLARE full_cost BIGINT DEFAULT 0;
    DECLARE span_start DATE;
    DECLARE span_end DATE;

    IF cost_cycle = 'once' THEN
        -- The cost occurs only once, based on the target_date
        IF curr_date < target_date THEN
            -- Cost applies as the target_date is in the future
            CASE amount_type
                WHEN 'fixed' THEN
                    RETURN amount;
                WHEN 'percentage' THEN
                    -- Precision of 3 decimals
                    RETURN salary * amount / 100000;
                END CASE;
        ELSE
            -- No cost as the target_date has already passed
            RETURN 0;
        END IF;
    END IF;

    -- Step 1: Determine span start and end based on curr_date and target_date
    IF target_date IS NULL THEN
        -- curr_date is after or equal to history start date
        SET span_start = calculate_next_history_execution_date(
            type,
            from_date,
            to_date,
            cycle,
            curr_date
        );
        SET span_end = calculate_cost_execution_date(
            type,
            from_date,
            to_date,
            cycle,
            NULL, -- No target_date
            cost_cycle,
            relative_offset,
            curr_date,
            TRUE
        );
    ELSE
        -- If target_date is NOT NULL, calculate normally
        IF curr_date < target_date THEN
            -- Current date is before the target date
            SET span_start = calculate_cost_execution_date(
                type,
                from_date,
                to_date,
                cycle,
                target_date,
                cost_cycle,
                relative_offset,
                curr_date,
                FALSE
            ); -- Calculate previous execution date
            SET span_end = target_date;
        ELSE
            -- Current date is on or after the target date
            SET span_end = calculate_cost_execution_date(
                type,
                from_date,
                to_date,
                cycle,
                target_date,
                cost_cycle,
                relative_offset,
                curr_date,
                TRUE
            ); -- Calculate next execution date
            SET span_start = calculate_cost_execution_date(
                type,
                from_date,
                to_date,
                cycle,
                span_end,
                cost_cycle,
                relative_offset,
                curr_date,
                FALSE
            );
        END IF;
    END IF;

    IF span_start = span_end THEN
        CASE amount_type
            WHEN 'fixed' THEN
                SET full_cost = amount;
            WHEN 'percentage' THEN
                -- Precision of 3 decimals
                SET full_cost = (salary * amount / 100000);
            END CASE;
        RETURN full_cost;
    END IF;

    -- Step 2: Iterate over the span and calculate costs
    span_loop: WHILE span_start < span_end DO
        SET @next_span_start = CASE cost_cycle
            WHEN 'daily' THEN DATE_ADD(span_start, INTERVAL 1 DAY)
            WHEN 'weekly' THEN DATE_ADD(span_start, INTERVAL 1 WEEK)
            WHEN 'monthly' THEN DATE_ADD(span_start, INTERVAL 1 MONTH)
            WHEN 'quarterly' THEN DATE_ADD(span_start, INTERVAL 3 MONTH)
            WHEN 'biannually' THEN DATE_ADD(span_start, INTERVAL 6 MONTH)
            WHEN 'yearly' THEN DATE_ADD(span_start, INTERVAL 1 YEAR)
        END;

        -- Check if the next span_start would exceed span_end
        IF @next_span_start > span_end THEN
            LEAVE span_loop;
        END IF;

        -- Add cost for the current execution
        CASE amount_type
            WHEN 'fixed' THEN
                SET full_cost = full_cost + amount;
            WHEN 'percentage' THEN
                -- Precision of 3 decimals
                SET full_cost = full_cost + (salary * amount / 100000);
            END CASE;

        -- Increment span_start
        SET span_start = @next_span_start;
    END WHILE;

    RETURN full_cost;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS calculate_next_cost_amount;
-- +goose StatementEnd
