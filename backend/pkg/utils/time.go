package utils

import (
	"time"
)

const InternalDateFormat = "2006-01-02"

type Clock interface {
	SetFixedTime(t *time.Time)
	Today() time.Time
}

type dynamicClock struct {
	FixedTime *time.Time
}

func (dc *dynamicClock) SetFixedTime(t *time.Time) {
	dc.FixedTime = t
}

func (dc *dynamicClock) Today() time.Time {
	if dc.FixedTime != nil {
		now := *dc.FixedTime
		return now.UTC()
	} else {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}
}

var DefaultClock Clock = &dynamicClock{}

func GetTodayAsUTC() time.Time {
	return DefaultClock.Today()
}

func GetNextDate(referenceDate, currentDate time.Time, months int) time.Time {
	dayDiff := referenceDate.Day() - currentDate.Day()

	nextDate := currentDate.AddDate(0, months, dayDiff)
	if currentDate.Day() > nextDate.Day() {
		nextDate = time.Date(nextDate.Year(), nextDate.Month(), 0, referenceDate.Hour(), referenceDate.Minute(), referenceDate.Second(), referenceDate.Nanosecond(), referenceDate.Location())
	}

	return nextDate
}

func GetLastDayOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month()+1, 0, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location())
}

func GetNextAvailableDate(fromDate, limitDate time.Time, cycle string) time.Time {
	current := fromDate
	for {
		switch cycle {
		case CycleMonthly:
			nextDate := GetNextDate(fromDate, current, 1)
			lastDayOfMonth := GetLastDayOfMonth(nextDate)
			if !lastDayOfMonth.Before(limitDate) {
				return current
			}
			current = nextDate
		case CycleQuarterly:
			nextDate := GetNextDate(fromDate, current, 3)
			lastDayOfMonth := GetLastDayOfMonth(nextDate)
			if !lastDayOfMonth.Before(limitDate) {
				return current
			}
			current = nextDate
		case CycleBiannually:
			nextDate := GetNextDate(fromDate, current, 6)
			lastDayOfMonth := GetLastDayOfMonth(nextDate)
			if !lastDayOfMonth.Before(limitDate) {
				return current
			}
			current = nextDate
		case CycleYearly:
			nextDate := GetNextDate(fromDate, current, 12)
			lastDayOfMonth := GetLastDayOfMonth(nextDate)
			if !lastDayOfMonth.Before(limitDate) {
				return current
			}
			current = nextDate
		}
	}
}

func GetTotalMonthsForMaxForecastYears() float64 {
	return float64(MaxForecastYears*12 + 1)
}
