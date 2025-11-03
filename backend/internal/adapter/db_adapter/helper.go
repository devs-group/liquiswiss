package db_adapter

import (
	"fmt"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
	"time"
)

func (d *DatabaseAdapter) CalculateSalaryExecutionDate(fromDatePtr types.AsDate, toDatePtr *types.AsDate, cycle *string, currDatePtr types.AsDate, relativeOffset int64, isNext bool) *time.Time {
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
					return toDate
				}
				return &exec
			}
		}
		offset += relativeOffset
	}
}

func (d *DatabaseAdapter) CalculateCostExecutionDate(
	fromDatePtr types.AsDate,
	toDatePtr *types.AsDate,
	salaryCycle string,
	targetDatePtr *types.AsDate,
	costCycle string,
	relativeOffset int64,
	currDatePtr types.AsDate,
	isNext bool,
) *types.AsDate {
	currDate := time.Time(currDatePtr)

	nextSalaryExecution := d.CalculateSalaryExecutionDate(fromDatePtr, toDatePtr, &salaryCycle, currDatePtr, 1, true)
	if nextSalaryExecution == nil {
		nextSalaryExecution = d.CalculateSalaryExecutionDate(fromDatePtr, toDatePtr, &salaryCycle, currDatePtr, 1, false)
	}

	if nextSalaryExecution == nil {
		return nil
	}

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
		lastPossibleExecutionDate := addCycle(*nextSalaryExecution, costCycle, relativeOffset)

		if nextSalaryExecution.After(targetDate) || nextSalaryExecution.Equal(targetDate) {
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

	costDate := addCycle(*nextSalaryExecution, costCycle, relativeOffset)

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

func (d *DatabaseAdapter) CalculateCostAmount(userID int64, cost models.SalaryCost, salary models.Salary, visited map[int64]struct{}) (uint64, error) {
	if visited == nil {
		visited = make(map[int64]struct{})
	}
	if _, alreadySeen := visited[cost.ID]; alreadySeen && cost.ID != 0 {
		return 0, fmt.Errorf("circular base salary cost reference detected")
	}
	if cost.ID != 0 {
		visited[cost.ID] = struct{}{}
		defer delete(visited, cost.ID)
	}

	switch cost.AmountType {
	case "fixed":
		return cost.Amount, nil
	case "percentage":
		var baseAmount uint64
		if len(cost.BaseSalaryCostIDs) > 0 {
			seen := make(map[int64]struct{}, len(cost.BaseSalaryCostIDs))
			for _, baseID := range cost.BaseSalaryCostIDs {
				if _, exists := seen[baseID]; exists {
					continue
				}
				seen[baseID] = struct{}{}

				baseCost, err := d.GetSalaryCost(userID, baseID)
				if err != nil {
					return 0, err
				}
				if baseCost.SalaryID != cost.SalaryID {
					return 0, fmt.Errorf("base salary cost does not belong to the same salary")
				}
				amount, err := d.CalculateCostAmount(userID, *baseCost, salary, visited)
				if err != nil {
					return 0, err
				}
			baseAmount += amount * models.SalaryCostDistributionMultiplier(baseCost.DistributionType)
			}
		} else {
			baseAmount = salary.Amount
		}
		return (baseAmount * cost.Amount) / 100_000, nil
	default:
		return 0, nil
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
