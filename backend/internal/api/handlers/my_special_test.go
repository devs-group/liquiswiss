package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"testing"
)

func TestSpecificCases(t *testing.T) {
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
	historyID1, err := dbService.CreateEmployeeHistory(models.CreateEmployeeHistory{
		HoursPerMonth:       160,
		Salary:              7500 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 29,
		FromDate:            "2025-01-31",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)

	historyID2, err := dbService.CreateEmployeeHistory(models.CreateEmployeeHistory{
		HoursPerMonth:       160,
		Salary:              25000 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 29,
		FromDate:            "2025-12-31",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	history1, err := dbService.GetEmployeeHistory(user.ID, historyID1)
	assert.NoError(t, err)
	history2, err := dbService.GetEmployeeHistory(user.ID, historyID2)
	assert.NoError(t, err)

	type TestCase struct {
		Description                   string
		DatabaseTime                  string
		ExpectedCalculatedAmount      uint64
		ExpectedNextCost              uint64
		ExpectedNextExecutionDate     string
		ExpectedPreviousExecutionDate string
		ExpectedEmployeeDeductions    uint64
		CreateData                    models.CreateEmployeeHistoryCost
	}

	testCases1 := []TestCase{
		//{
		//	Description:                   "BU 1",
		//	DatabaseTime:                  "2025-07-08",
		//	ExpectedCalculatedAmount:      6.08 * 100,
		//	ExpectedNextCost:              66.88 * 100,
		//	ExpectedEmployeeDeductions:    6.07 * 100,
		//	ExpectedNextExecutionDate:     "2026-01-31",
		//	ExpectedPreviousExecutionDate: "2025-02-28",
		//	CreateData: models.CreateEmployeeHistoryCost{
		//		Cycle:            "monthly",
		//		AmountType:       "percentage",
		//		Amount:           0.081 * 1000,
		//		DistributionType: "employee",
		//		RelativeOffset:   11,
		//		TargetDate:       utils.StringAsPointer("2026-01-31"),
		//		LabelID:          nil,
		//	},
		//},
		{
			Description:                   "BVG 1",
			DatabaseTime:                  "2025-05-08",
			ExpectedCalculatedAmount:      302_60,
			ExpectedNextCost:              907_80,
			ExpectedEmployeeDeductions:    302_60,
			ExpectedNextExecutionDate:     "2025-07-30",
			ExpectedPreviousExecutionDate: "2025-04-30",
			CreateData: models.CreateEmployeeHistoryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           302_60,
				DistributionType: "employee",
				RelativeOffset:   3,
				TargetDate:       utils.StringAsPointer("2025-04-30"),
				LabelID:          nil,
			},
		},
		//{
		//	Description:                   "BVG 2",
		//	DatabaseTime:                  "2025-08-08",
		//	ExpectedCalculatedAmount:      30260,
		//	ExpectedNextCost:              121040,
		//	ExpectedEmployeeDeductions:    30260,
		//	ExpectedNextExecutionDate:     "2025-10-30",
		//	ExpectedPreviousExecutionDate: "2025-07-30",
		//	CreateData: models.CreateEmployeeHistoryCost{
		//		Cycle:            "monthly",
		//		AmountType:       "fixed",
		//		Amount:           30260,
		//		DistributionType: "employee",
		//		RelativeOffset:   3,
		//		TargetDate:       utils.StringAsPointer("2025-04-30"),
		//		LabelID:          nil,
		//	},
		//},
	}

	testCases2 := []TestCase{
		//{
		//	Description:                   "BU 2",
		//	DatabaseTime:                  "2025-07-08",
		//	ExpectedCalculatedAmount:      20.25 * 100,
		//	ExpectedNextCost:              20.25 * 100,
		//	ExpectedEmployeeDeductions:    20.25 * 100,
		//	ExpectedNextExecutionDate:     "2026-01-31",
		//	ExpectedPreviousExecutionDate: "2025-12-31",
		//	CreateData: models.CreateEmployeeHistoryCost{
		//		Cycle:            "monthly",
		//		AmountType:       "percentage",
		//		Amount:           0.081 * 1000,
		//		DistributionType: "employee",
		//		RelativeOffset:   1,
		//		TargetDate:       utils.StringAsPointer("2026-01-31"),
		//		LabelID:          nil,
		//	},
		//},
	}

	for _, testCase := range testCases1 {
		t.Run(testCase.Description, func(t *testing.T) {
			historyCostID, err := dbService.CreateEmployeeHistoryCost(testCase.CreateData, user.ID, history1.ID)
			assert.NoError(t, err)

			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			historyCost, err := dbService.GetEmployeeHistoryCost(user.ID, historyCostID)
			assert.NoError(t, err)

			// Check deduction
			updatedHistory, err := dbService.GetEmployeeHistory(user.ID, history1.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedHistory.EmployeeDeductions), "updatedHistory.EmployeeDeductions")

			err = dbService.DeleteEmployeeHistoryCost(historyCost.ID, user.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedCalculatedAmount), int64(historyCost.CalculatedAmount), "historyCost.CalculatedAmount")
			assert.Equal(t, int64(testCase.ExpectedNextCost), int64(historyCost.CalculatedNextCost), "historyCost.CalculatedNextCost")
			assert.Equal(t, testCase.ExpectedNextExecutionDate, historyCost.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat), "historyCost.CalculatedNextExecutionDate")
			assert.Equal(t, testCase.ExpectedPreviousExecutionDate, historyCost.CalculatedPreviousExecutionDate.ToFormattedTime(utils.InternalDateFormat), "historyCost.CalculatedPreviousExecutionDate")
			if historyCost.Label != nil {
				assert.Equal(t, label.Name, historyCost.Label.Name)
			} else {
				assert.Nil(t, historyCost.Label)
			}
		})
	}

	for _, testCase := range testCases2 {
		t.Run(testCase.Description, func(t *testing.T) {
			historyCostID, err := dbService.CreateEmployeeHistoryCost(testCase.CreateData, user.ID, history2.ID)
			assert.NoError(t, err)

			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			historyCost, err := dbService.GetEmployeeHistoryCost(user.ID, historyCostID)
			assert.NoError(t, err)

			// Check deduction
			updatedHistory, err := dbService.GetEmployeeHistory(user.ID, history2.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedHistory.EmployeeDeductions), "updatedHistory.EmployeeDeductions")

			err = dbService.DeleteEmployeeHistoryCost(historyCost.ID, user.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedCalculatedAmount), int64(historyCost.CalculatedAmount), "historyCost.CalculatedAmount")
			assert.Equal(t, int64(testCase.ExpectedNextCost), int64(historyCost.CalculatedNextCost), "historyCost.CalculatedNextCost")
			assert.Equal(t, testCase.ExpectedNextExecutionDate, historyCost.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat), "historyCost.CalculatedNextExecutionDate")
			assert.Equal(t, testCase.ExpectedPreviousExecutionDate, historyCost.CalculatedPreviousExecutionDate.ToFormattedTime(utils.InternalDateFormat), "historyCost.CalculatedPreviousExecutionDate")
			if historyCost.Label != nil {
				assert.Equal(t, label.Name, historyCost.Label.Name)
			} else {
				assert.Nil(t, historyCost.Label)
			}
		})
	}
}
