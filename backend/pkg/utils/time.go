package utils

import (
	"time"
)

const InternalDateFormat = "2006-01-02"

func GetTodayAsUTC() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
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

func GetNextAvailableDate(fromDate, limitDate time.Time) time.Time {
	current := fromDate
	for {
		nextDate := GetNextDate(fromDate, current, 1)
		lastDayOfMonth := GetLastDayOfMonth(nextDate)
		if !lastDayOfMonth.Before(limitDate) {
			return current
		}
		current = nextDate
	}
}
