package db_service

import (
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
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
			percentValue := (salary * cost.Amount) / 100_000

			switch cost.Cycle {
			case "once":
				value = percentValue
			case "monthly":
				value = multiplyByHistoryScaleUint(percentValue, historyCycle, 1)
			case "quarterly":
				value = multiplyByHistoryScaleUint(percentValue, historyCycle, 3)
			case "biannually":
				value = multiplyByHistoryScaleUint(percentValue, historyCycle, 6)
			case "yearly":
				value = multiplyByHistoryScaleUint(percentValue, historyCycle, 12)
			}

		} else if cost.AmountType == "fixed" {
			switch cost.Cycle {
			case "once":
				value = cost.Amount
			case "monthly":
				value = multiplyByHistoryScaleUint(cost.Amount, historyCycle, 1)
			case "quarterly":
				value = multiplyByHistoryScaleUint(cost.Amount, historyCycle, 3)
			case "biannually":
				value = multiplyByHistoryScaleUint(cost.Amount, historyCycle, 6)
			case "yearly":
				value = multiplyByHistoryScaleUint(cost.Amount, historyCycle, 12)
			}
		}

		total += value
	}

	return total
}

func (s *DatabaseService) CalculateEmployeeCostAmount(
	salary uint64,
	amount uint64,
	amountType string,
) uint64 {
	if amountType == "percentage" {
		return (salary * amount) / 100_000
	}
	return amount
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

	nextHistoryExecution := s.CalculateHistoryExecutionDate(fromDatePtr, toDatePtr, &historyCycle, currDatePtr, relativeOffset, true)
	//previousHistoryExecution := s.CalculateHistoryExecutionDate(fromDatePtr, toDatePtr, historyCycle, currDatePtr, relativeOffset, false)

	if costCycle == "once" {
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

func (s *DatabaseService) CalculateNextCostAmount(
	fromDatePtr types.AsDate,
	toDatePtr *types.AsDate,
	historyCycle string,
	targetDatePtr *types.AsDate,
	costCycle string,
	relativeOffset int64,
	currDatePtr types.AsDate,
	amountType string,
	amount uint64,
	salary uint64,
) uint64 {
	var fullCost uint64

	currDate := time.Time(currDatePtr)

	if costCycle == "once" {
		if targetDatePtr != nil && currDate.Before(time.Time(*targetDatePtr)) {
			if amountType == "fixed" {
				return amount
			}
			if amountType == "percentage" {
				return (salary * amount) / 100_000
			}
		}
		return 0
	}

	nextHistoryExecution := s.CalculateHistoryExecutionDate(fromDatePtr, toDatePtr, &historyCycle, currDatePtr, relativeOffset, true)

	var spanStart time.Time
	var spanEnd time.Time

	if targetDatePtr == nil {
		spanStart = *nextHistoryExecution
		spanEndPtr := s.CalculateCostExecutionDate(
			fromDatePtr,
			toDatePtr,
			historyCycle,
			nil,
			costCycle,
			relativeOffset,
			currDatePtr,
			true,
		)
		if spanEndPtr == nil {
			return 0
		}
		spanEnd = time.Time(*spanEndPtr)
	} else {
		targetDate := time.Time(*targetDatePtr)
		if currDate.Before(targetDate) {
			spanStartPtr := s.CalculateCostExecutionDate(
				fromDatePtr,
				toDatePtr,
				historyCycle,
				targetDatePtr,
				costCycle,
				relativeOffset,
				currDatePtr,
				false,
			)
			if spanStartPtr == nil {
				return 0
			}
			spanStart = time.Time(*spanStartPtr)
			spanEnd = targetDate
		} else {
			spanEndPtr := s.CalculateCostExecutionDate(
				fromDatePtr,
				toDatePtr,
				historyCycle,
				targetDatePtr,
				costCycle,
				relativeOffset,
				currDatePtr,
				true,
			)
			if spanEndPtr == nil {
				return 0
			}
			spanEnd = time.Time(*spanEndPtr)

			spanStartPtr := s.CalculateCostExecutionDate(
				fromDatePtr,
				toDatePtr,
				historyCycle,
				spanEndPtr,
				costCycle,
				relativeOffset,
				currDatePtr,
				false,
			)
			if spanStartPtr == nil {
				return 0
			}
			spanStart = time.Time(*spanStartPtr)

			//if spanEnd.After(*previousHistoryExec) {
			//	var adjustedSpanEnd time.Time
			//	switch costCycle {
			//	case "monthly":
			//		adjustedSpanEnd = nextHistoryExecution.AddDate(0, 1, 0)
			//	case "quarterly":
			//		adjustedSpanEnd = nextHistoryExecution.AddDate(0, 3, 0)
			//	case "biannually":
			//		adjustedSpanEnd = nextHistoryExecution.AddDate(0, 6, 0)
			//	case "yearly":
			//		adjustedSpanEnd = nextHistoryExecution.AddDate(1, 0, 0)
			//	}
			//	if adjustedSpanEnd.After(spanEnd) {
			//		spanEnd = adjustedSpanEnd
			//	}
			//}
		}
	}

	if spanStart.Equal(spanEnd) {
		if amountType == "fixed" {
			return amount
		}
		if amountType == "percentage" {
			return (salary * amount) / 100_000
		}
	}

	for spanStart.Before(spanEnd) {
		nextSpan := addCycle(spanStart, costCycle, 1)

		if !nextSpan.After(spanEnd) {
			if amountType == "fixed" {
				fullCost += amount
			} else if amountType == "percentage" {
				fullCost += (salary * amount) / 100_000
			}
		}

		spanStart = nextSpan
	}

	return fullCost
}

func multiplyByHistoryScaleUint(base uint64, historyCycle string, costMonths uint64) uint64 {
	switch historyCycle {
	case "monthly":
		return base / costMonths
	case "quarterly":
		return base * 3 / costMonths
	case "biannually":
		return base * 6 / costMonths
	case "yearly":
		return base * 12 / costMonths
	default:
		return base
	}
}

func addCycle(t time.Time, cycle string, offset int64) time.Time {
	var months int
	switch cycle {
	case "monthly":
		months = int(offset)
	case "quarterly":
		months = int(offset * 3)
	case "biannually":
		months = int(offset * 6)
	case "yearly":
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
