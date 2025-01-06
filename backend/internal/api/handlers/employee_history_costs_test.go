package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"testing"
)

func TestEmployeeHistoryCosts(t *testing.T) {
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

	label, err := CreateEmployeeHistoryCostLabel(dbService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	historyID, err := dbService.CreateEmployeeHistory(models.CreateEmployeeHistory{
		HoursPerMonth: 160,
		// Salary of 10'000.00 CHF
		Salary:              10000 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-26",
		ToDate:              nil,
		// We want to test separate costs
		WithSeparateCosts: true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	history, err := dbService.GetEmployeeHistory(user.ID, historyID)
	assert.NoError(t, err)

	type TestCase struct {
		Description                   string
		DatabaseTime                  string
		ExpectedCalculcatedAmount     uint64
		ExpectedNextCost              uint64
		ExpectedNextExecutionDate     string
		ExpectedPreviousExecutionDate string
		CreateData                    models.CreateEmployeeHistoryCost
	}

	testCases := []TestCase{
		// Fixed Amounts
		{
			Description:                   "Monthly fixed with offset of 1",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     1250 * 100,
			ExpectedNextCost:              1250 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           1250 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Monthly fixed with offset of 6",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 2000 * 100,
			// Expecting 6 times the amount due to monthly cycle with offset of 6
			ExpectedNextCost:              12000 * 100,
			ExpectedNextExecutionDate:     "2025-07-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           2000 * 100,
				DistributionType: "employee",
				RelativeOffset:   6,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly fixed with offset of 2",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     3750 * 100,
			ExpectedNextCost:              7500 * 100,
			ExpectedNextExecutionDate:     "2025-07-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           3750 * 100,
				DistributionType: "employer",
				RelativeOffset:   2,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly fixed with Target Date",
			DatabaseTime:                  "2025-06-15",
			ExpectedCalculcatedAmount:     15000 * 100,
			ExpectedNextCost:              15000 * 100,
			ExpectedNextExecutionDate:     "2026-06-15",
			ExpectedPreviousExecutionDate: "2025-06-15",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           15000 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       utils.StringAsPointer("2025-06-15"),
				LabelID:          &label.ID,
			},
		},

		// Percentage Amounts
		{
			Description:  "Monthly percentage with offset of 1",
			DatabaseTime: "2025-01-01",
			// 15% of 10'000 CHF
			ExpectedCalculcatedAmount:     1500 * 100,
			ExpectedNextCost:              1500 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           15 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:  "Biannually percentage with offset of 1",
			DatabaseTime: "2025-01-01",
			// 20% of 10'000 CHF
			ExpectedCalculcatedAmount:     2000 * 100,
			ExpectedNextCost:              2000 * 100,
			ExpectedNextExecutionDate:     "2025-07-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "biannually",
				AmountType:       "percentage",
				Amount:           20 * 1000, // 20%
				DistributionType: "employer",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:  "Daily percentage with Target Date",
			DatabaseTime: "2025-01-02",
			// 2% of 10'000 CHF
			ExpectedCalculcatedAmount: 200 * 100,
			// Expecting 7 time the calculated amount
			ExpectedNextCost:              1400 * 100,
			ExpectedNextExecutionDate:     "2025-01-08",
			ExpectedPreviousExecutionDate: "2025-01-01",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "daily",
				AmountType:       "percentage",
				Amount:           2 * 1000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       utils.StringAsPointer("2025-01-01"),
				LabelID:          &label.ID,
			},
		},

		// Edge Cases
		{
			Description:                   "Weekly fixed with offset of 4",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     1000 * 100,
			ExpectedNextCost:              4000 * 100,
			ExpectedNextExecutionDate:     "2025-02-23",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "weekly",
				AmountType:       "fixed",
				Amount:           1000 * 100,
				DistributionType: "employer",
				RelativeOffset:   4,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:  "Quarterly percentage with Target Date",
			DatabaseTime: "2025-01-01",
			// 30% of 10'000 CHF
			ExpectedCalculcatedAmount:     3000 * 100,
			ExpectedNextCost:              3000 * 100,
			ExpectedNextExecutionDate:     "2025-01-15",
			ExpectedPreviousExecutionDate: "2024-10-15",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "quarterly",
				AmountType:       "percentage",
				Amount:           30 * 1000, // 30%
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       utils.StringAsPointer("2024-04-15"),
				LabelID:          nil,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			historyCostID, err := dbService.CreateEmployeeHistoryCost(testCase.CreateData, user.ID, history.ID)
			assert.NoError(t, err)

			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			historyCost, err := dbService.GetEmployeeHistoryCost(user.ID, historyCostID)
			assert.NoError(t, err)

			err = dbService.DeleteEmployeeHistoryCost(historyCost.ID, user.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedCalculcatedAmount), int64(historyCost.CalculatedAmount), "historyCost.CalculatedAmount")
			assert.Equal(t, int64(testCase.ExpectedNextCost), int64(historyCost.NextCost), "historyCost.NextCost")
			assert.Equal(t, testCase.ExpectedNextExecutionDate, historyCost.NextExecutionDate.ToFormattedTime(utils.InternalDateFormat), "historyCost.NextExecutionDate")
			assert.Equal(t, testCase.ExpectedPreviousExecutionDate, historyCost.PreviousExecutionDate.ToFormattedTime(utils.InternalDateFormat), "historyCost.PreviousExecutionDate")
			if historyCost.Label != nil {
				assert.Equal(t, label.Name, historyCost.Label.Name)
			} else {
				assert.Nil(t, historyCost.Label)
			}
		})
	}
}

func TestEmployeeHistoryCostDeductions(t *testing.T) {
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
	historyID, err := dbService.CreateEmployeeHistory(models.CreateEmployeeHistory{
		HoursPerMonth: 160,
		// Salary of 10'000.00 CHF
		Salary:              10000 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-26",
		ToDate:              nil,
		// We want to test separate costs
		WithSeparateCosts: true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	history, err := dbService.GetEmployeeHistory(user.ID, historyID)
	assert.NoError(t, err)

	type TestCase struct {
		Description                string
		DatabaseTime               string
		ExpectedEmployeeDeductions uint64
		CreateData                 models.CreateEmployeeHistoryCost
	}

	testCases := []TestCase{
		{
			Description:                "Once with fixed amount",
			DatabaseTime:               "2025-01-01",
			ExpectedEmployeeDeductions: 500 * 100,
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "once",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       utils.StringAsPointer("2025-01-15"),
			},
		},
		{
			Description:                "Once with percentage amount",
			DatabaseTime:               "2025-01-01",
			ExpectedEmployeeDeductions: 2000 * 100,
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "once",
				AmountType:       "percentage",
				Amount:           20 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       utils.StringAsPointer("2025-01-15"),
			},
		},
		{
			Description:                "Daily fixed",
			DatabaseTime:               "2025-01-01",
			ExpectedEmployeeDeductions: 1250 * 100 * 6,
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "daily",
				AmountType:       "fixed",
				Amount:           1250 * 100,
				DistributionType: "employee",
				RelativeOffset:   6,
				TargetDate:       nil,
			},
		},
		{
			Description:                "Weekly fixed",
			DatabaseTime:               "2025-01-01",
			ExpectedEmployeeDeductions: 1250 * 100 * 6,
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "weekly",
				AmountType:       "fixed",
				Amount:           1250 * 100,
				DistributionType: "employee",
				RelativeOffset:   6,
				TargetDate:       nil,
			},
		},
		{
			Description:                "Monthly fixed",
			DatabaseTime:               "2025-01-01",
			ExpectedEmployeeDeductions: 1250 * 100,
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           1250 * 100,
				DistributionType: "employee",
				RelativeOffset:   3,
				TargetDate:       nil,
			},
		},
		{
			Description:                "Quarterly fixed",
			DatabaseTime:               "2025-01-01",
			ExpectedEmployeeDeductions: 41667,
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           1250 * 100,
				DistributionType: "employee",
				RelativeOffset:   3,
				TargetDate:       nil,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			historyCostID, err := dbService.CreateEmployeeHistoryCost(testCase.CreateData, user.ID, history.ID)
			assert.NoError(t, err)

			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			historyCost, err := dbService.GetEmployeeHistoryCost(user.ID, historyCostID)
			assert.NoError(t, err)

			// Check deduction
			updatedHistory, err := dbService.GetEmployeeHistory(user.ID, historyID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedHistory.EmployeeDeductions))

			err = dbService.DeleteEmployeeHistoryCost(historyCost.ID, user.ID)
			assert.NoError(t, err)
		})
	}
}
