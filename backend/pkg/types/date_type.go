package types

import (
	"fmt"
	"time"
)

// CustomDate is a custom type for date formatting
type AsDate time.Time

func (cd AsDate) MarshalJSON() ([]byte, error) {
	formattedDate := fmt.Sprintf("\"%s\"", time.Time(cd).Format("2006-01-02"))
	return []byte(formattedDate), nil
}

func (cd AsDate) ToString() string {
	formattedDate := fmt.Sprintf("%s", time.Time(cd).Format("2006-01-02"))
	return formattedDate
}

// UnmarshalJSON parses the date from "YYYY-MM-DD" format for JSON deserialization
func (cd *AsDate) UnmarshalJSON(data []byte) error {
	// Remove the quotes from the string if necessary
	strDate := string(data)
	strDate = strDate[1 : len(strDate)-1]

	// Parse the date
	parsedDate, err := time.Parse("2006-01-02", strDate)
	if err != nil {
		return err
	}

	*cd = AsDate(parsedDate)
	// Set the parsed time to the AsDate
	return nil
}
