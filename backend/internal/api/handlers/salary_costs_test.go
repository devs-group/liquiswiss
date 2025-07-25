package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"testing"
)

func TestMonthlySalaryWithoutToDate(t *testing.T) {
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

	salaryCostLabel, err := CreateSalaryCostLabel(dbService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salaryID, _, _, err := dbService.CreateSalary(models.CreateSalary{
		HoursPerMonth: 160,
		// SalaryAmount of 10'000.00 CHF
		Amount:              10000 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-07-08",
		ToDate:              nil,
		// We want to test separate costs
		WithSeparateCosts: true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	salary, err := dbService.GetSalary(user.ID, salaryID)
	assert.NoError(t, err)

	type TestCase struct {
		Description                   string
		DatabaseTime                  string
		ExpectedCalculatedAmount      uint64
		ExpectedNextCost              uint64
		ExpectedNextExecutionDate     string
		ExpectedPreviousExecutionDate string
		ExpectedEmployeeDeductions    uint64
		CreateData                    models.CreateSalaryCost
	}

	// The salary starts from 2025-01-26 (see above)
	testCases := []TestCase{
		// Once
		{
			Description:                   "Once simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500 * 100,
			ExpectedNextCost:              500 * 100,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-02-15",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
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
			ExpectedCalculatedAmount:      2500 * 100,
			ExpectedNextCost:              2500 * 100,
			ExpectedEmployeeDeductions:    2500 * 100,
			ExpectedNextExecutionDate:     "2025-02-15",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "once",
				AmountType:       "percentage",
				Amount:           25 * 1000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       utils.StringAsPointer("2025-02-15"),
				LabelID:          nil,
			},
		},

		// Monthly
		{
			Description:                   "Monthly simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500 * 100,
			ExpectedNextCost:              500 * 100,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:                   "Monthly simple fixed 2",
			DatabaseTime:                  "2025-01-28",
			ExpectedCalculatedAmount:      500 * 100,
			ExpectedNextCost:              500 * 100,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			ExpectedCalculatedAmount:      2500 * 100,
			ExpectedNextCost:              2500 * 100,
			ExpectedEmployeeDeductions:    2500 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			ExpectedCalculatedAmount:      500 * 100,
			ExpectedNextCost:              500 * 100 * 7,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2026-02-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			ExpectedCalculatedAmount:      7500 * 100,
			ExpectedNextCost:              7500 * 100 * 7,
			ExpectedEmployeeDeductions:    7500 * 100,
			ExpectedNextExecutionDate:     "2026-02-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			ExpectedCalculatedAmount:      250 * 100,
			ExpectedNextCost:              250 * 100 * 720,
			ExpectedEmployeeDeductions:    250 * 100,
			ExpectedNextExecutionDate:     "2085-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Quarterly simple fixed",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 500 * 100,
			ExpectedNextCost:         500 * 100,
			// 500 * 100 / 3
			ExpectedEmployeeDeductions:    16667,
			ExpectedNextExecutionDate:     "2025-04-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Quarterly simple percentage",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 2500 * 100,
			ExpectedNextCost:         2500 * 100,
			// 2500 * 100 / 3
			ExpectedEmployeeDeductions:    83333,
			ExpectedNextExecutionDate:     "2025-04-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Quarterly offset fixed",
			DatabaseTime:             "2025-06-26",
			ExpectedCalculatedAmount: 500 * 100,
			ExpectedNextCost:         500 * 100 * 7,
			// 500 * 100 / 3
			ExpectedEmployeeDeductions:    16667,
			ExpectedNextExecutionDate:     "2027-04-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Quarterly offset percentage",
			DatabaseTime:             "2025-06-26",
			ExpectedCalculatedAmount: 7500 * 100,
			ExpectedNextCost:         7500 * 100 * 7,
			// 7500 * 100 / 3
			ExpectedEmployeeDeductions:    250000,
			ExpectedNextExecutionDate:     "2027-04-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Quarterly giga offset fixed",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 250 * 100,
			ExpectedNextCost:         250 * 100 * 720,
			// 250 * 100 / 3
			ExpectedEmployeeDeductions:    8333,
			ExpectedNextExecutionDate:     "2205-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Biannually simple fixed",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 500 * 100,
			ExpectedNextCost:         500 * 100,
			// 500 * 100 / 6
			ExpectedEmployeeDeductions:    8333,
			ExpectedNextExecutionDate:     "2025-07-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Biannually simple percentage",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 2500 * 100,
			ExpectedNextCost:         2500 * 100,
			// 2500 * 100 / 6
			ExpectedEmployeeDeductions:    41667,
			ExpectedNextExecutionDate:     "2025-07-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Biannually offset fixed",
			DatabaseTime:             "2025-06-26",
			ExpectedCalculatedAmount: 500 * 100,
			ExpectedNextCost:         500 * 100 * 7,
			// 500 * 100 / 6
			ExpectedEmployeeDeductions:    8333,
			ExpectedNextExecutionDate:     "2029-01-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Biannually offset percentage",
			DatabaseTime:             "2025-06-26",
			ExpectedCalculatedAmount: 7500 * 100,
			ExpectedNextCost:         7500 * 100 * 7,
			// 7500 * 100 / 6
			ExpectedEmployeeDeductions:    125000,
			ExpectedNextExecutionDate:     "2029-01-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Biannually giga offset fixed",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 250 * 100,
			ExpectedNextCost:         250 * 100 * 720,
			// 250 * 100 / 6
			ExpectedEmployeeDeductions:    4167,
			ExpectedNextExecutionDate:     "2385-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Yearly simple fixed",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 500 * 100,
			ExpectedNextCost:         500 * 100,
			// 500 * 100 / 12
			ExpectedEmployeeDeductions:    4167,
			ExpectedNextExecutionDate:     "2026-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Yearly simple percentage",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 2500 * 100,
			ExpectedNextCost:         2500 * 100,
			// 2500 * 100 / 12
			ExpectedEmployeeDeductions:    20833,
			ExpectedNextExecutionDate:     "2026-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Yearly offset fixed",
			DatabaseTime:             "2025-06-26",
			ExpectedCalculatedAmount: 500 * 100,
			ExpectedNextCost:         500 * 100 * 7,
			// 2500 * 100 / 12
			ExpectedEmployeeDeductions:    4167,
			ExpectedNextExecutionDate:     "2032-07-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Yearly offset percentage",
			DatabaseTime:             "2025-06-26",
			ExpectedCalculatedAmount: 7500 * 100,
			ExpectedNextCost:         7500 * 100 * 7,
			// 7500 * 100 / 12
			ExpectedEmployeeDeductions:    62500,
			ExpectedNextExecutionDate:     "2032-07-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description:              "Yearly giga offset fixed",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 250 * 100,
			ExpectedNextCost:         250 * 100 * 720,
			// 250 * 100 / 12
			ExpectedEmployeeDeductions:    2083,
			ExpectedNextExecutionDate:     "2745-01-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			salaryCostID, err := dbService.CreateSalaryCost(testCase.CreateData, user.ID, salary.ID)
			assert.NoError(t, err)

			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			salaryCost, err := dbService.GetSalaryCost(user.ID, salaryCostID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := dbService.GetSalary(user.ID, salaryID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = dbService.DeleteSalaryCost(salaryCost.ID, user.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedCalculatedAmount), int64(salaryCost.CalculatedAmount), "salaryCost.CalculatedAmount")
			assert.Equal(t, int64(testCase.ExpectedNextCost), int64(salaryCost.CalculatedNextCost), "salaryCost.CalculatedNextCost")
			assert.Equal(t, testCase.ExpectedNextExecutionDate, salaryCost.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat), "salaryCost.CalculatedNextExecutionDate")
			assert.Equal(t, testCase.ExpectedPreviousExecutionDate, salaryCost.CalculatedPreviousExecutionDate.ToFormattedTime(utils.InternalDateFormat), "salaryCost.CalculatedPreviousExecutionDate")
			if salaryCost.Label != nil {
				assert.Equal(t, salaryCostLabel.Name, salaryCost.Label.Name)
			} else {
				assert.Nil(t, salaryCost.Label)
			}
		})
	}
}

func TestMonthlySalaryWithToDate(t *testing.T) {
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

	label, err := CreateSalaryCostLabel(dbService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salaryID, _, _, err := dbService.CreateSalary(models.CreateSalary{
		HoursPerMonth: 160,
		// SalaryAmount of 10'000.00 CHF
		Amount:              10000 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-26",
		ToDate:              utils.StringAsPointer("2025-07-26"),
		// We want to test separate costs
		WithSeparateCosts: true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	salary, err := dbService.GetSalary(user.ID, salaryID)
	assert.NoError(t, err)

	type TestCase struct {
		Description                   string
		DatabaseTime                  string
		ExpectedCalculatedAmount      uint64
		ExpectedNextCost              uint64
		ExpectedNextExecutionDate     string
		ExpectedPreviousExecutionDate string
		ExpectedEmployeeDeductions    uint64
		CreateData                    models.CreateSalaryCost
	}

	// The salary starts from 2025-01-26 and ends at 2025-07-26
	testCases := []TestCase{
		// Monthly
		{
			Description:                   "Monthly simple fixed before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500 * 100,
			ExpectedNextCost:              500 * 100,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:                   "Monthly simple percentage before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2000 * 100,
			ExpectedNextCost:              2000 * 100,
			ExpectedEmployeeDeductions:    2000 * 100,
			ExpectedNextExecutionDate:     "2025-02-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:                   "Monthly simple fixed passed salary end but in range",
			DatabaseTime:                  "2025-08-26",
			ExpectedCalculatedAmount:      250 * 100,
			ExpectedNextCost:              250 * 100,
			ExpectedEmployeeDeductions:    250 * 100,
			ExpectedNextExecutionDate:     "2025-08-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description:                   "Monthly simple percentage passed salary end but in range",
			DatabaseTime:                  "2025-08-26",
			ExpectedCalculatedAmount:      3500 * 100,
			ExpectedNextCost:              3500 * 100,
			ExpectedEmployeeDeductions:    3500 * 100,
			ExpectedNextExecutionDate:     "2025-08-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description: "Monthly simple fixed after salary out of range",
			// Mind one day after a month after the salary ended
			DatabaseTime:                  "2025-08-27",
			ExpectedCalculatedAmount:      300 * 100,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    300 * 100,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "2025-08-26",
			CreateData: models.CreateSalaryCost{
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
			Description: "Monthly simple percentage after salary out of range",
			// Mind one day after a month after the salary ended
			DatabaseTime:                  "2025-08-27",
			ExpectedCalculatedAmount:      1500 * 100,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    1500 * 100,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "2025-08-26",
			CreateData: models.CreateSalaryCost{
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
			Description:                   "Monthly offset fixed before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500 * 100,
			ExpectedNextCost:              500 * 100 * 10,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2025-11-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:                   "Monthly percentage fixed before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2000 * 100,
			ExpectedNextCost:              2000 * 100 * 10,
			ExpectedEmployeeDeductions:    2000 * 100,
			ExpectedNextExecutionDate:     "2025-11-26",
			ExpectedPreviousExecutionDate: "2025-01-26",
			CreateData: models.CreateSalaryCost{
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
			Description:                   "Monthly offset fixed passed salary end but in range",
			DatabaseTime:                  "2026-05-26",
			ExpectedCalculatedAmount:      500 * 100,
			ExpectedNextCost:              500 * 100 * 10,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "2026-05-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description:                   "Monthly offset percentage passed salary end but in range",
			DatabaseTime:                  "2026-05-26",
			ExpectedCalculatedAmount:      3500 * 100,
			ExpectedNextCost:              3500 * 100 * 10,
			ExpectedEmployeeDeductions:    3500 * 100,
			ExpectedNextExecutionDate:     "2026-05-26",
			ExpectedPreviousExecutionDate: "2025-07-26",
			CreateData: models.CreateSalaryCost{
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
			Description: "Monthly offset fixed passed salary end out of range",
			// Mind one day 10 months after the salary ended
			DatabaseTime:                  "2026-05-27",
			ExpectedCalculatedAmount:      500 * 100,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    500 * 100,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "2026-05-26",
			CreateData: models.CreateSalaryCost{
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
			Description:                   "Monthly offset percentage passed salary end out of range",
			DatabaseTime:                  "2026-05-27",
			ExpectedCalculatedAmount:      1500 * 100,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    1500 * 100,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "2026-05-26",
			CreateData: models.CreateSalaryCost{
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
			salaryCostID, err := dbService.CreateSalaryCost(testCase.CreateData, user.ID, salary.ID)
			assert.NoError(t, err)

			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			salaryCost, err := dbService.GetSalaryCost(user.ID, salaryCostID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := dbService.GetSalary(user.ID, salaryID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = dbService.DeleteSalaryCost(salaryCost.ID, user.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedCalculatedAmount), int64(salaryCost.CalculatedAmount), "salaryCost.CalculatedAmount")
			assert.Equal(t, int64(testCase.ExpectedNextCost), int64(salaryCost.CalculatedNextCost), "salaryCost.CalculatedNextCost")
			assert.Equal(t, testCase.ExpectedNextExecutionDate, salaryCost.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat), "salaryCost.CalculatedNextExecutionDate")
			assert.Equal(t, testCase.ExpectedPreviousExecutionDate, salaryCost.CalculatedPreviousExecutionDate.ToFormattedTime(utils.InternalDateFormat), "salaryCost.CalculatedPreviousExecutionDate")
			if salaryCost.Label != nil {
				assert.Equal(t, label.Name, salaryCost.Label.Name)
			} else {
				assert.Nil(t, salaryCost.Label)
			}
		})
	}
}
