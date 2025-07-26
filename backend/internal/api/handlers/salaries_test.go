package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"testing"
	"time"
)

func TestSalaryExecutionDates(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	// Preparations
	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "john@doe.com", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(apiService, user.ID, "Tom Riddle")
	assert.NoError(t, err)

	// Tests
	salaryAmount := uint64(5000_75)

	type TestCase struct {
		Description           string
		DatabaseTime          string
		ExpectedExecutionDate *string
		CreateData            models.CreateSalary
	}

	testCases := []TestCase{
		{
			Description:           "Salary in past, still running",
			DatabaseTime:          "2025-01-01",
			ExpectedExecutionDate: utils.StringAsPointer("2025-01-26"),
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
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
			Description:           "Salary in past, expired",
			DatabaseTime:          "2025-01-01",
			ExpectedExecutionDate: nil,
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2024-03-26",
				// This ends the salary before the January salary, so December must be the last
				ToDate:            utils.StringAsPointer("2025-01-25"),
				WithSeparateCosts: false,
			},
		},
		{
			Description:  "End of January adjust to end of February",
			DatabaseTime: "2025-02-01",
			// 2025-01-31 should become 2025-02-28
			ExpectedExecutionDate: utils.StringAsPointer("2025-02-28"),
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
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
			ExpectedExecutionDate: utils.StringAsPointer("2025-03-31"),
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-31",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description:           "Salary in future in same month",
			DatabaseTime:          "2025-01-01",
			ExpectedExecutionDate: utils.StringAsPointer("2025-01-26"),
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
				Cycle:               "monthly",
				CurrencyID:          *currency.ID,
				VacationDaysPerYear: 25,
				FromDate:            "2025-01-26",
				ToDate:              nil,
				WithSeparateCosts:   false,
			},
		},
		{
			Description: "Salary in past in future month",
			// In the future here
			DatabaseTime:          "2025-02-01",
			ExpectedExecutionDate: utils.StringAsPointer("2025-02-26"),
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
				Cycle:               "monthly",
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
			ExpectedExecutionDate: utils.StringAsPointer("2025-04-26"),
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
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
			ExpectedExecutionDate: utils.StringAsPointer("2026-01-26"),
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
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
			ExpectedExecutionDate: utils.StringAsPointer("2028-01-26"),
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
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
			ExpectedExecutionDate: nil,
			CreateData: models.CreateSalary{
				HoursPerMonth:       160,
				Amount:              salaryAmount,
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
			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, testCase.DatabaseTime)
			assert.NoError(t, err)

			utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)
			defer func() {
				utils.DefaultClock.SetFixedTime(nil)
			}()

			salary, err := apiService.CreateSalary(testCase.CreateData, user.ID, employee.ID)
			assert.NoError(t, err)

			err = apiService.DeleteSalary(user.ID, salary.ID)
			assert.NoError(t, err)

			if salary.NextExecutionDate != nil {
				assert.Equal(t, *testCase.ExpectedExecutionDate, time.Time(*salary.NextExecutionDate).Format(utils.InternalDateFormat))
			} else {
				assert.Nil(t, testCase.ExpectedExecutionDate)
			}
			assert.Equal(t, salaryAmount, salary.Amount)
			assert.Equal(t, "Swiss Franc", *salary.Currency.Description)
		})
	}
}
