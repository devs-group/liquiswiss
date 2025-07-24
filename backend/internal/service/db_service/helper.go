package db_service

import (
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
	"time"
)

func (s *DatabaseService) CalculateHistoryExecutionDate(fromDatePtr types.AsDate, toDatePtr *types.AsDate, cycle *string, currDatePtr types.AsDate, relativeOffset int64, isNext bool) *time.Time {
	fromDate := time.Time(fromDatePtr)
	currDate := time.Time(currDatePtr)
	var toDate *time.Time
	if toDatePtr != nil {
		val := time.Time(*toDatePtr)
		toDate = &val
	}

	if cycle == nil {
		if toDate != nil && fromDate.After(*toDate) {
			return toDate
		}
		return &fromDate
	}

	if currDate.Before(fromDate) {
		if toDate != nil && fromDate.After(*toDate) {
			return toDate
		}
		return &fromDate
	}

	offset := relativeOffset
	for {
		exec := addCycle(fromDate, *cycle, offset)

		if isNext {
			if exec.After(currDate) || exec.Equal(currDate) {
				if toDate != nil && exec.After(*toDate) {
					return nil
				}
				return &exec
			}
		} else {
			next := addCycle(fromDate, *cycle, offset+relativeOffset)
			if next.After(currDate) {
				if toDate != nil && exec.After(*toDate) {
					return nil
				}
				return &exec
			}
		}
		offset += relativeOffset
	}
}

func (s *DatabaseService) CalculateSalaryAdjustments(
	salary uint64,
	historyCycle string,
	distributionType string,
	costs []models.EmployeeHistoryCost,
) uint64 {
	var total uint64 = 0

	for _, cost := range costs {
		if cost.DistributionType != distributionType {
			continue
		}

		var value uint64

		if cost.AmountType == "percentage" {
			switch cost.Cycle {
			case utils.CycleOnce:
				value = cost.CalculatedAmount
			case utils.CycleMonthly:
				value = multiplyByHistoryScaleUint(cost.CalculatedAmount, historyCycle, 1)
			case utils.CycleQuarterly:
				value = multiplyByHistoryScaleUint(cost.CalculatedAmount, historyCycle, 3)
			case utils.CycleBiannually:
				value = multiplyByHistoryScaleUint(cost.CalculatedAmount, historyCycle, 6)
			case utils.CycleYearly:
				value = multiplyByHistoryScaleUint(cost.CalculatedAmount, historyCycle, 12)
			}

		} else if cost.AmountType == "fixed" {
			switch cost.Cycle {
			case utils.CycleOnce:
				value = cost.Amount
			case utils.CycleMonthly:
				value = multiplyByHistoryScaleUint(cost.Amount, historyCycle, 1)
			case utils.CycleQuarterly:
				value = multiplyByHistoryScaleUint(cost.Amount, historyCycle, 3)
			case utils.CycleBiannually:
				value = multiplyByHistoryScaleUint(cost.Amount, historyCycle, 6)
			case utils.CycleYearly:
				value = multiplyByHistoryScaleUint(cost.Amount, historyCycle, 12)
			}
		}

		total += value
	}

	return total
}

func (s *DatabaseService) CalculateCostExecutionDate(
	fromDatePtr types.AsDate,
	toDatePtr *types.AsDate,
	historyCycle string,
	targetDatePtr *types.AsDate,
	costCycle string,
	relativeOffset int64,
	currDatePtr types.AsDate,
	isNext bool,
) *types.AsDate {
	currDate := time.Time(currDatePtr)

	nextHistoryExecution := s.CalculateHistoryExecutionDate(fromDatePtr, toDatePtr, &historyCycle, currDatePtr, 1, true)
	//previousHistoryExecution := s.CalculateHistoryExecutionDate(fromDatePtr, toDatePtr, historyCycle, currDatePtr, relativeOffset, false)

	if costCycle == utils.CycleOnce {
		if isNext {
			if targetDatePtr == nil || currDate.After(time.Time(*targetDatePtr)) {
				return nil
			}
			return targetDatePtr
		}
		return nil
	}

	if targetDatePtr != nil {
		targetDate := time.Time(*targetDatePtr)
		lastPossibleExecutionDate := addCycle(*nextHistoryExecution, costCycle, relativeOffset)

		if nextHistoryExecution.After(targetDate) || nextHistoryExecution.Equal(targetDate) {
			for targetDate.Before(currDate) || targetDate.Equal(currDate) {
				next := addCycle(targetDate, costCycle, relativeOffset)
				if next.After(lastPossibleExecutionDate) {
					break
				}
				targetDate = next
			}
		}

		if isNext {
			if currDate.After(targetDate) {
				return nil
			}
			as := types.AsDate(targetDate)
			return &as
		}

		prev := subtractCycle(targetDate, costCycle, relativeOffset)
		as := types.AsDate(prev)
		return &as
	}

	costDate := addCycle(*nextHistoryExecution, costCycle, relativeOffset)

	if isNext {
		if currDate.After(costDate) {
			return nil
		}
		as := types.AsDate(costDate)
		return &as
	}

	if currDate.After(costDate) {
		as := types.AsDate(costDate)
		return &as
	}

	previous := subtractCycle(costDate, costCycle, relativeOffset)
	as := types.AsDate(previous)
	return &as
}

func (s *DatabaseService) CalculateCostAmount(cost models.EmployeeHistoryCost, history models.EmployeeHistory) uint64 {
	if cost.AmountType == "fixed" {
		return cost.Amount
	}
	if cost.AmountType == "percentage" {
		return (history.Salary * cost.Amount) / 100_000
	}
	return 0
}

func multiplyByHistoryScaleUint(base uint64, historyCycle string, costMonths uint64) uint64 {
	switch historyCycle {
	case utils.CycleMonthly:
		return base / costMonths
	case utils.CycleQuarterly:
		return base * 3 / costMonths
	case utils.CycleBiannually:
		return base * 6 / costMonths
	case utils.CycleYearly:
		return base * 12 / costMonths
	default:
		return base
	}
}

func addCycle(t time.Time, cycle string, offset int64) time.Time {
	var months int
	switch cycle {
	case utils.CycleMonthly:
		months = int(offset)
	case utils.CycleQuarterly:
		months = int(offset * 3)
	case utils.CycleBiannually:
		months = int(offset * 6)
	case utils.CycleYearly:
		months = int(offset * 12)
	default:
		return t
	}

	isEndOfMonth := t.Day() == lastDayOfMonth(t)

	year, month := t.Year(), int(t.Month())+months
	for month > 12 {
		year++
		month -= 12
	}
	for month <= 0 {
		year--
		month += 12
	}

	var day int
	if isEndOfMonth {
		day = lastDayOfMonth(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, t.Location()))
	} else {
		day = min(t.Day(), lastDayOfMonth(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, t.Location())))
	}

	return time.Date(year, time.Month(month), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func subtractCycle(t time.Time, cycle string, offset int64) time.Time {
	return addCycle(t, cycle, -offset)
}

func lastDayOfMonth(t time.Time) int {
	year, month := t.Year(), t.Month()
	return time.Date(year, month+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
