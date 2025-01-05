package types

import (
	"fmt"
	"liquiswiss/pkg/utils"
	"time"
)

// CustomDate is a custom type for date formatting
type AsDate time.Time

func (cd AsDate) MarshalJSON() ([]byte, error) {
	formattedDate := fmt.Sprintf("\"%s\"", time.Time(cd).Format(utils.InternalDateFormat))
	return []byte(formattedDate), nil
}

func (cd AsDate) ToString() string {
	formattedDate := fmt.Sprintf("%s", time.Time(cd).Format(utils.InternalDateFormat))
	return formattedDate
}

// ToFormattedTime returns either a formatted string date or an empty string
func (cd *AsDate) ToFormattedTime(format string) string {
	if cd != nil {
		return time.Time(*cd).Format(format)
	}
	return ""
}

// UnmarshalJSON parses the date from "YYYY-MM-DD" format for JSON deserialization
func (cd *AsDate) UnmarshalJSON(data []byte) error {
	// Remove the quotes from the string if necessary
	strDate := string(data)
	strDate = strDate[1 : len(strDate)-1]

	// Parse the date
	parsedDate, err := time.Parse(utils.InternalDateFormat, strDate)
	if err != nil {
		return err
	}

	*cd = AsDate(parsedDate)
	// Set the parsed time to the AsDate
	return nil
}
