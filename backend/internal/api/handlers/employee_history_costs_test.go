package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"testing"
)

func TestMonthlyEmployeeHistoryWithoutToDate(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbService := db_service.NewDatabaseService(conn)

	// Preparations
	currency, err := CreateCurrency(dbService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		dbService, "John Doe", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(dbService, user.ID, "Tom Riddle")
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
		ExpectedEmployeeDeductions    uint64
		CreateData                    models.CreateEmployeeHistoryCost
	}

	// The history starts from 2025-01-26 (see above)
	testCases := []TestCase{
		// Once
		{
			Description:                   "Once simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-02-15",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "once",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       utils.StringAsPointer("2025-02-15"),
				LabelID:          nil,
			},
		},
		{
			Description:                   "Once simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     2500 * 100,
			ExpectedNextCost:              2500 * 100,
			ExpectedEmployeeDeductions:    2500 * 100,
			ExpectedNextExecutionDate:     "2025-02-15",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "once",
				AmountType:       "percentage",
				Amount:           25 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       utils.StringAsPointer("2025-02-15"),
				LabelID:          nil,
			},
		},

		// Daily
		{
			Description:                   "Daily simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-01-27",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "daily",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Daily simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     2500 * 100,
			ExpectedNextCost:              2500 * 100,
			ExpectedEmployeeDeductions:    2500 * 100,
			ExpectedNextExecutionDate:     "2025-01-27",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "daily",
				AmountType:       "percentage",
				Amount:           25 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Daily offset fixed",
			DatabaseTime:                  "2025-06-26",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100 * 7,
			ExpectedEmployeeDeductions:    500 * 100 * 7,
			ExpectedNextExecutionDate:     "2025-08-02",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "daily",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Daily offset percentage",
			DatabaseTime:                  "2025-06-26",
			ExpectedCalculcatedAmount:     7500 * 100,
			ExpectedNextCost:              7500 * 100 * 7,
			ExpectedEmployeeDeductions:    7500 * 100 * 7,
			ExpectedNextExecutionDate:     "2025-08-02",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "daily",
				AmountType:       "percentage",
				Amount:           75 * 1000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Daily giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     250 * 100,
			ExpectedNextCost:              250 * 100 * 720,
			ExpectedEmployeeDeductions:    250 * 100 * 720,
			ExpectedNextExecutionDate:     "2027-01-16",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "daily",
				AmountType:       "fixed",
				Amount:           250 * 100,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Weekly
		{
			Description:                   "Weekly simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-02-02",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "weekly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Weekly simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     2500 * 100,
			ExpectedNextCost:              2500 * 100,
			ExpectedEmployeeDeductions:    2500 * 100,
			ExpectedNextExecutionDate:     "2025-02-02",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "weekly",
				AmountType:       "percentage",
				Amount:           25 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Weekly offset fixed",
			DatabaseTime:                  "2025-06-26",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100 * 7,
			ExpectedEmployeeDeductions:    500 * 100 * 7,
			ExpectedNextExecutionDate:     "2025-09-13",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "weekly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Weekly offset percentage",
			DatabaseTime:                  "2025-06-26",
			ExpectedCalculcatedAmount:     7500 * 100,
			ExpectedNextCost:              7500 * 100 * 7,
			ExpectedEmployeeDeductions:    7500 * 100 * 7,
			ExpectedNextExecutionDate:     "2025-09-13",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "weekly",
				AmountType:       "percentage",
				Amount:           75 * 1000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Weekly giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     250 * 100,
			ExpectedNextCost:              250 * 100 * 720,
			ExpectedEmployeeDeductions:    250 * 100 * 720,
			ExpectedNextExecutionDate:     "2038-11-14",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "weekly",
				AmountType:       "fixed",
				Amount:           250 * 100,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Monthly
		{
			Description:                   "Monthly simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     2500 * 100,
			ExpectedNextCost:              2500 * 100,
			ExpectedEmployeeDeductions:    2500 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           25 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset fixed",
			DatabaseTime:                  "2025-06-26",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100 * 7,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2026-02-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset percentage",
			DatabaseTime:                  "2025-06-26",
			ExpectedCalculcatedAmount:     7500 * 100,
			ExpectedNextCost:              7500 * 100 * 7,
			ExpectedEmployeeDeductions:    7500 * 100,
			ExpectedNextExecutionDate:     "2026-02-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           75 * 1000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     250 * 100,
			ExpectedNextCost:              250 * 100 * 720,
			ExpectedEmployeeDeductions:    250 * 100,
			ExpectedNextExecutionDate:     "2085-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           250 * 100,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Quarterly
		{
			Description:               "Quarterly simple fixed",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 500 * 100,
			ExpectedNextCost:          500 * 100,
			// 500 * 100 / 3
			ExpectedEmployeeDeductions:    16667,
			ExpectedNextExecutionDate:     "2025-04-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Quarterly simple percentage",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 2500 * 100,
			ExpectedNextCost:          2500 * 100,
			// 2500 * 100 / 3
			ExpectedEmployeeDeductions:    83333,
			ExpectedNextExecutionDate:     "2025-04-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "quarterly",
				AmountType:       "percentage",
				Amount:           25 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Quarterly offset fixed",
			DatabaseTime:              "2025-06-26",
			ExpectedCalculcatedAmount: 500 * 100,
			ExpectedNextCost:          500 * 100 * 7,
			// 500 * 100 / 3
			ExpectedEmployeeDeductions:    16667,
			ExpectedNextExecutionDate:     "2027-04-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Quarterly offset percentage",
			DatabaseTime:              "2025-06-26",
			ExpectedCalculcatedAmount: 7500 * 100,
			ExpectedNextCost:          7500 * 100 * 7,
			// 7500 * 100 / 3
			ExpectedEmployeeDeductions:    250000,
			ExpectedNextExecutionDate:     "2027-04-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "quarterly",
				AmountType:       "percentage",
				Amount:           75 * 1000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Quarterly giga offset fixed",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 250 * 100,
			ExpectedNextCost:          250 * 100 * 720,
			// 250 * 100 / 3
			ExpectedEmployeeDeductions:    8333,
			ExpectedNextExecutionDate:     "2205-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           250 * 100,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Biannually
		{
			Description:               "Biannually simple fixed",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 500 * 100,
			ExpectedNextCost:          500 * 100,
			// 500 * 100 / 6
			ExpectedEmployeeDeductions:    8333,
			ExpectedNextExecutionDate:     "2025-07-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Biannually simple percentage",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 2500 * 100,
			ExpectedNextCost:          2500 * 100,
			// 2500 * 100 / 6
			ExpectedEmployeeDeductions:    41667,
			ExpectedNextExecutionDate:     "2025-07-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "biannually",
				AmountType:       "percentage",
				Amount:           25 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Biannually offset fixed",
			DatabaseTime:              "2025-06-26",
			ExpectedCalculcatedAmount: 500 * 100,
			ExpectedNextCost:          500 * 100 * 7,
			// 500 * 100 / 6
			ExpectedEmployeeDeductions:    8333,
			ExpectedNextExecutionDate:     "2029-01-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Biannually offset percentage",
			DatabaseTime:              "2025-06-26",
			ExpectedCalculcatedAmount: 7500 * 100,
			ExpectedNextCost:          7500 * 100 * 7,
			// 7500 * 100 / 6
			ExpectedEmployeeDeductions:    125000,
			ExpectedNextExecutionDate:     "2029-01-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "biannually",
				AmountType:       "percentage",
				Amount:           75 * 1000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Biannually giga offset fixed",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 250 * 100,
			ExpectedNextCost:          250 * 100 * 720,
			// 250 * 100 / 6
			ExpectedEmployeeDeductions:    4167,
			ExpectedNextExecutionDate:     "2385-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           250 * 100,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Yearly
		{
			Description:               "Yearly simple fixed",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 500 * 100,
			ExpectedNextCost:          500 * 100,
			// 500 * 100 / 12
			ExpectedEmployeeDeductions:    4167,
			ExpectedNextExecutionDate:     "2026-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Yearly simple percentage",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 2500 * 100,
			ExpectedNextCost:          2500 * 100,
			// 2500 * 100 / 12
			ExpectedEmployeeDeductions:    20833,
			ExpectedNextExecutionDate:     "2026-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "yearly",
				AmountType:       "percentage",
				Amount:           25 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Yearly offset fixed",
			DatabaseTime:              "2025-06-26",
			ExpectedCalculcatedAmount: 500 * 100,
			ExpectedNextCost:          500 * 100 * 7,
			// 2500 * 100 / 12
			ExpectedEmployeeDeductions:    4167,
			ExpectedNextExecutionDate:     "2032-07-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Yearly offset percentage",
			DatabaseTime:              "2025-06-26",
			ExpectedCalculcatedAmount: 7500 * 100,
			ExpectedNextCost:          7500 * 100 * 7,
			// 7500 * 100 / 12
			ExpectedEmployeeDeductions:    62500,
			ExpectedNextExecutionDate:     "2032-07-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "yearly",
				AmountType:       "percentage",
				Amount:           75 * 1000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:               "Yearly giga offset fixed",
			DatabaseTime:              "2025-01-01",
			ExpectedCalculcatedAmount: 250 * 100,
			ExpectedNextCost:          250 * 100 * 720,
			// 250 * 100 / 12
			ExpectedEmployeeDeductions:    2083,
			ExpectedNextExecutionDate:     "2745-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           250 * 100,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
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

			// Check deduction
			updatedHistory, err := dbService.GetEmployeeHistory(user.ID, historyID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedHistory.EmployeeDeductions), "updatedHistory.EmployeeDeductions")

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

func TestMonthlyEmployeeHistoryWithToDate(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbService := db_service.NewDatabaseService(conn)

	// Preparations
	currency, err := CreateCurrency(dbService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		dbService, "John Doe", "test", "Test Organisation",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(dbService, user.ID, "Tom Riddle")
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
		ToDate:              utils.StringAsPointer("2025-07-26"),
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
		ExpectedEmployeeDeductions    uint64
		CreateData                    models.CreateEmployeeHistoryCost
	}

	// The history starts from 2025-01-26 and ends at 2025-07-26
	testCases := []TestCase{
		// Monthly
		{
			Description:                   "Monthly simple fixed before history ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple percentage before history ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     2000 * 100,
			ExpectedNextCost:              2000 * 100,
			ExpectedEmployeeDeductions:    2000 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           20 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple fixed passed history end but in range",
			DatabaseTime:                  "2025-08-26",
			ExpectedCalculcatedAmount:     250 * 100,
			ExpectedNextCost:              250 * 100,
			ExpectedEmployeeDeductions:    250 * 100,
			ExpectedNextExecutionDate:     "2025-08-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           250 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple percentage passed history end but in range",
			DatabaseTime:                  "2025-08-26",
			ExpectedCalculcatedAmount:     3500 * 100,
			ExpectedNextCost:              3500 * 100,
			ExpectedEmployeeDeductions:    3500 * 100,
			ExpectedNextExecutionDate:     "2025-08-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           35 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly simple fixed after history out of range",
			// Mind one day after a month after the history ended
			DatabaseTime:                  "2025-08-27",
			ExpectedCalculcatedAmount:     300 * 100,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    300 * 100,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "2025-08-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           300 * 100,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly simple percentage after history out of range",
			// Mind one day after a month after the history ended
			DatabaseTime:                  "2025-08-27",
			ExpectedCalculcatedAmount:     1500 * 100,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    1500 * 100,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "2025-08-26",
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
			Description:                   "Monthly offset fixed before history ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100 * 10,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-11-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly percentage fixed before history ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculcatedAmount:     2000 * 100,
			ExpectedNextCost:              2000 * 100 * 10,
			ExpectedEmployeeDeductions:    2000 * 100,
			ExpectedNextExecutionDate:     "2025-11-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           20 * 1000,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset fixed passed history end but in range",
			DatabaseTime:                  "2026-05-26",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              500 * 100 * 10,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2026-05-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset percentage passed history end but in range",
			DatabaseTime:                  "2026-05-26",
			ExpectedCalculcatedAmount:     3500 * 100,
			ExpectedNextCost:              3500 * 100 * 10,
			ExpectedEmployeeDeductions:    3500 * 100,
			ExpectedNextExecutionDate:     "2026-05-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           35 * 1000,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly offset fixed passed history end out of range",
			// Mind one day 10 months after the history ended
			DatabaseTime:                  "2026-05-27",
			ExpectedCalculcatedAmount:     500 * 100,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "2026-05-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500 * 100,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset percentage passed history end out of range",
			DatabaseTime:                  "2026-05-27",
			ExpectedCalculcatedAmount:     1500 * 100,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    1500 * 100,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "2026-05-26",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           15 * 1000,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
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

			// Check deduction
			updatedHistory, err := dbService.GetEmployeeHistory(user.ID, historyID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedHistory.EmployeeDeductions), "updatedHistory.EmployeeDeductions")

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
