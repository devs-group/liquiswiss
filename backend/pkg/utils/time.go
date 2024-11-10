package utils

import "time"

const InternalDateFormat = "2006-01-02"

func GetTodayAsUTC() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
