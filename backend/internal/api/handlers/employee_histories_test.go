package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"testing"
	"time"
)

func TestEmployeeHistoryExecutionDates(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbService := db_service.NewDatabaseService(conn)

	// Preparations
	user, _, err := CreateUserWithOrganisation(
		dbService, "John Doe", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(dbService, user.ID, "Tom Riddle")
	assert.NoError(t, err)

	currency, err := CreateCurrency(dbService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	// Tests
	salary := uint64(5000.75 * 100)

	type TestCase struct {
		Description           string
		DatabaseTime          string
		ExpectedExecutionDate string
		CreateData            models.CreateEmployeeHistory
	}

	testCases := []TestCase{
		{
			Description:           "History in past, still running",
			DatabaseTime:          "2025-01-01",
			ExpectedExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				// This begins in the past and due to monthly cycle and no ToDate this is due in 2025-01
				FromDate:          "2024-03-26",
				ToDate:            nil,
				WithSeparateCosts: false,
			},
		},
		{
			Description:           "History in past, expired",
			DatabaseTime:          "2025-01-01",
			ExpectedExecutionDate: "2024-12-26",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2024-03-26",
				// This ends the history before the January salary, so December must be the last
				ToDate:            utils.StringAsPointer("2025-01-25"),
				WithSeparateCosts: false,
			},
		},
		{
			Description:  "End of January adjust to end of February",
			DatabaseTime: "2025-02-01",
			// 2025-01-31 should become 2025-02-28
			ExpectedExecutionDate: "2025-02-28",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-31",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description:  "End of January adjust to end of March after February",
			DatabaseTime: "2025-03-01",
			// 2025-01-31 should become 2025-03-31 after February has been 2025-02-28
			ExpectedExecutionDate: "2025-03-31",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-31",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description:           "History in future in same month",
			DatabaseTime:          "2025-01-01",
			ExpectedExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-26",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description: "History in past in future month",
			// In the future here
			DatabaseTime:          "2025-02-01",
			ExpectedExecutionDate: "2025-02-26",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-26",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description:  "Daily cycle is always day after today",
			DatabaseTime: "2025-02-28",
			// On a daily cycle the next execution date is always a day after today
			ExpectedExecutionDate: "2025-03-01",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "daily",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-26",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description:  "A weekly cycle matches",
			DatabaseTime: "2025-02-02",
			// 2025-01-26 + a week would be 2025-02-02, since today is 2025-02-02 the next week would be 2025-02-09
			ExpectedExecutionDate: "2025-02-09",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "weekly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-26",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description:  "A quarterly cycle matches",
			DatabaseTime: "2025-04-24",
			// 2025-01-26 + 3 months which is after today so this will match
			ExpectedExecutionDate: "2025-04-26",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "quarterly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-26",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description:  "A biannually cycle matches",
			DatabaseTime: "2025-07-27",
			// 2025-01-26 + 6 months which is 2025-07-26 since today is after that we add another 6 months so we end up next year
			ExpectedExecutionDate: "2026-01-26",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "biannually",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-26",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description:  "A yearly cycle matches",
			DatabaseTime: "2028-01-25",
			// 2025-01-26 + 12 months which is 2026-01-26 but we are in 2028 already, so we would expect this to be in January 2028
			ExpectedExecutionDate: "2028-01-26",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "biannually",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-26",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description:  "Month end edge case",
			DatabaseTime: "2025-01-01",
			// Chicken or Egg, we decided to not handle the month edge case yet
			ExpectedExecutionDate: "2024-10-31",
			CreateData: models.CreateEmployeeHistory{
				HoursPerMonth:       160,
				Salary:              salary,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2024-01-31",
				ToDate:              utils.StringAsPointer("2024-11-30"),
				WithSeparateCosts:   false,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			historyID, err := dbService.CreateEmployeeHistory(testCase.CreateData, user.ID, employee.ID)
			assert.NoError(t, err)

			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			history, err := dbService.GetEmployeeHistory(user.ID, historyID)
			assert.NoError(t, err)

			err = dbService.DeleteEmployeeHistory(history, user.ID)
			assert.NoError(t, err)

			assert.NotNil(t, history.NextExecutionDate)
			assert.Equal(t, testCase.ExpectedExecutionDate, time.Time(*history.NextExecutionDate).Format(utils.InternalDateFormat))
			assert.Equal(t, salary, history.Salary)
			assert.Equal(t, "Swiss Franc", *history.Currency.Description)
		})
	}
}
