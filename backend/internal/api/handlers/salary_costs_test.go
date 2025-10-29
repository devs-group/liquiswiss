package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"math"
	"testing"
	"time"
)

func TestMonthlySalaryAtTheEndOfMonthWithoutToDate(t *testing.T) {
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

	salaryCostLabel, err := CreateSalaryCostLabel(apiService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-31",
		ToDate:              nil,
		// We want to test separate costs
		WithSeparateCosts: true,
	}, user.ID, employee.ID)
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

	testCases := []TestCase{
		// Once
		{
			Description:                   "Once simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_000,
			ExpectedNextCost:              500_000,
			ExpectedEmployeeDeductions:    500_000,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "once",
				AmountType:       "fixed",
				Amount:           500_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       utils.StringAsPointer("2025-02-15"),
				LabelID:          nil,
			},
		},
		{
			Description:                   "Once simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "once",
				AmountType:       "percentage",
				Amount:           25_000,
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
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple fixed 2",
			DatabaseTime:                  "2025-01-28",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           25_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset fixed before salary",
			DatabaseTime:                  "2025-06-29",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset fixed after salary",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2026-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset percentage",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      7500_00,
			ExpectedNextCost:              7500_00 * 7,
			ExpectedEmployeeDeductions:    7500_00,
			ExpectedNextExecutionDate:     "2026-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           75_000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              uint64(250_00 * utils.GetTotalMonthsForMaxForecastYears()),
			ExpectedEmployeeDeductions:    250_00,
			ExpectedNextExecutionDate:     "2085-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Quarterly
		{
			Description:                   "Quarterly simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00 / 3,
			ExpectedNextExecutionDate:     "2025-04-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00 / 3,
			ExpectedNextExecutionDate:     "2025-04-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "percentage",
				Amount:           25_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly offset fixed before salary",
			DatabaseTime:                  "2025-06-29",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00 / 3,
			ExpectedNextExecutionDate:     "2027-03-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly offset fixed after salary",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00 / 3,
			ExpectedNextExecutionDate:     "2027-04-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly offset percentage",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      7500_00,
			ExpectedNextCost:              7500_00 * 7,
			ExpectedEmployeeDeductions:    7500_00 / 3,
			ExpectedNextExecutionDate:     "2027-04-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "percentage",
				Amount:           75_000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              uint64(250_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/3)),
			ExpectedEmployeeDeductions:    250_00 / 3,
			ExpectedNextExecutionDate:     "2205-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Biannually
		{
			Description:                   "Biannually simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00 / 6,
			ExpectedNextExecutionDate:     "2025-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00 / 6,
			ExpectedNextExecutionDate:     "2025-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "percentage",
				Amount:           25_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually offset fixed before salary",
			DatabaseTime:                  "2025-06-29",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00 / 6,
			ExpectedNextExecutionDate:     "2028-12-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually offset fixed after salary",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00 / 6,
			ExpectedNextExecutionDate:     "2029-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually offset percentage",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      7500_00,
			ExpectedNextCost:              7500_00 * 7,
			ExpectedEmployeeDeductions:    7500_00 / 6,
			ExpectedNextExecutionDate:     "2029-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "percentage",
				Amount:           75_000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              uint64(250_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/6)),
			ExpectedEmployeeDeductions:    250_00 / 6,
			ExpectedNextExecutionDate:     "2385-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Yearly
		{
			Description:                   "Yearly simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00 / 12,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00 / 12,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "percentage",
				Amount:           25_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly offset fixed before salary",
			DatabaseTime:                  "2025-06-29",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              uint64(500_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/12)),
			ExpectedEmployeeDeductions:    500_00 / 12,
			ExpectedNextExecutionDate:     "2032-06-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly offset fixed after salary",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              uint64(500_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/12)),
			ExpectedEmployeeDeductions:    500_00 / 12,
			ExpectedNextExecutionDate:     "2032-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly offset percentage",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      7500_00,
			ExpectedNextCost:              uint64(7500_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/12)),
			ExpectedEmployeeDeductions:    7500_00 / 12,
			ExpectedNextExecutionDate:     "2032-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "percentage",
				Amount:           75_000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              uint64(250_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/12)),
			ExpectedEmployeeDeductions:    250_00 / 12,
			ExpectedNextExecutionDate:     "2745-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
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

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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

func TestMonthlySalaryAtTheEndOfMonthWithToDate(t *testing.T) {
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

	label, err := CreateSalaryCostLabel(apiService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth: 160,
		// SalaryAmount of 10'000.00 CHF
		Amount:              10000 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-31",
		ToDate:              utils.StringAsPointer("2025-07-31"),
		// We want to test separate costs
		WithSeparateCosts: true,
	}, user.ID, employee.ID)
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
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple percentage before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2000_00,
			ExpectedNextCost:              2000_00,
			ExpectedEmployeeDeductions:    2000_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           20_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple fixed passed salary end but in range",
			DatabaseTime:                  "2025-08-30",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              250_00,
			ExpectedEmployeeDeductions:    250_00,
			ExpectedNextExecutionDate:     "2025-08-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple percentage passed salary end but in range",
			DatabaseTime:                  "2025-08-30",
			ExpectedCalculatedAmount:      3500_00,
			ExpectedNextCost:              3500_00,
			ExpectedEmployeeDeductions:    3500_00,
			ExpectedNextExecutionDate:     "2025-08-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           35_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly simple fixed after salary out of range",
			// Mind one day after a month after the salary ended
			DatabaseTime:                  "2025-09-01",
			ExpectedCalculatedAmount:      0,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           300_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly simple percentage after salary out of range",
			// Mind one day after a month after the salary ended
			DatabaseTime:                  "2025-09-01",
			ExpectedCalculatedAmount:      0,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           15_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset fixed before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 5,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-06-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   5,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly percentage fixed before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2000_00,
			ExpectedNextCost:              2000_00 * 5,
			ExpectedEmployeeDeductions:    2000_00,
			ExpectedNextExecutionDate:     "2025-06-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           20_000,
				DistributionType: "employee",
				RelativeOffset:   5,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:              "Monthly offset fixed passed salary end but in range",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 500_00,
			// Because salary ends on 31.07 we have a maximum of 7 months
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-11-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:              "Monthly offset percentage passed salary end but in range",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 3500_00,
			// Because salary ends on 31.07 we have a maximum of 7 months
			ExpectedNextCost:              3500_00 * 7,
			ExpectedEmployeeDeductions:    3500_00,
			ExpectedNextExecutionDate:     "2025-11-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           35_000,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly offset fixed passed salary end with current date but in range",
			// Because salary ends on 31.07 adding 10 months would be the last on 31.05 so we are in range
			DatabaseTime:                  "2026-05-27",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2026-05-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly offset percentage passed salary end with current date but in range",
			// Because salary ends on 31.07 adding 10 months would be the last on 31.05 so we are in range
			DatabaseTime:                  "2026-05-27",
			ExpectedCalculatedAmount:      1500_00,
			ExpectedNextCost:              1500_00,
			ExpectedEmployeeDeductions:    1500_00,
			ExpectedNextExecutionDate:     "2026-05-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           15_000,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly offset fixed passed salary end with current date out of range",
			// Because salary ends on 31.07 adding 10 months would be the last on 31.05
			DatabaseTime:                  "2026-06-01",
			ExpectedCalculatedAmount:      0,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly offset percentage passed salary end with current date out of range",
			// Because salary ends on 31.07 adding 10 months would be the last on 31.05
			DatabaseTime:                  "2026-06-01",
			ExpectedCalculatedAmount:      0,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           15_000,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
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

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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

func TestMultipleSalaryAtTheEndOfMonthCases(t *testing.T) {
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

	salaryCostLabel, err := CreateSalaryCostLabel(apiService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salary1, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              7_500 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 29,
		FromDate:            "2025-01-31",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	salary2, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10_000 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 29,
		FromDate:            "2025-09-30",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
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

	testCases1 := []TestCase{
		{
			Description:                   "BVG 2. Quartal",
			DatabaseTime:                  "2025-07-09",
			ExpectedCalculatedAmount:      200_00,
			ExpectedNextCost:              600_00,
			ExpectedEmployeeDeductions:    200_00,
			ExpectedNextExecutionDate:     "2025-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           200_00,
				DistributionType: "employee",
				RelativeOffset:   3,
				TargetDate:       utils.StringAsPointer("2025-01-31"),
				LabelID:          &salaryCostLabel.ID,
			},
		},
		{
			Description:                   "BVG 3. Quartal",
			DatabaseTime:                  "2025-08-01",
			ExpectedCalculatedAmount:      200_00,
			ExpectedNextCost:              400_00,
			ExpectedEmployeeDeductions:    200_00,
			ExpectedNextExecutionDate:     "2025-10-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           200_00,
				DistributionType: "employee",
				RelativeOffset:   3,
				TargetDate:       utils.StringAsPointer("2025-01-31"),
				LabelID:          &salaryCostLabel.ID,
			},
		},
	}

	testCases2 := []TestCase{
		{
			Description:                   "BVG 3. Quartal Neues Gehalt",
			DatabaseTime:                  "2025-07-25",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-10-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   3,
				TargetDate:       utils.StringAsPointer("2025-01-31"),
				LabelID:          &salaryCostLabel.ID,
			},
		},
	}

	for _, testCase := range testCases1 {
		t.Run(testCase.Description, func(t *testing.T) {
			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, testCase.DatabaseTime)
			assert.NoError(t, err)

			utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)
			defer func() {
				utils.DefaultClock.SetFixedTime(nil)
			}()

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary1.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary1.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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

	for _, testCase := range testCases2 {
		t.Run(testCase.Description, func(t *testing.T) {
			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, testCase.DatabaseTime)
			assert.NoError(t, err)

			utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)
			defer func() {
				utils.DefaultClock.SetFixedTime(nil)
			}()

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary2.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary2.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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

func TestLongOffsetScenariosAtTheEndOfMonth(t *testing.T) {
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

	salaryCostLabel, err := CreateSalaryCostLabel(apiService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-31",
		ToDate:              nil,
		// We want to test separate costs
		WithSeparateCosts: true,
	}, user.ID, employee.ID)
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

	testCases := []TestCase{
		{
			Description:                   "Monthly with 12 months relative offset",
			DatabaseTime:                  "2025-01-26",
			ExpectedCalculatedAmount:      15_00,
			ExpectedNextCost:              180_00,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           15_00,
				DistributionType: "employer",
				RelativeOffset:   12,
				TargetDate:       utils.StringAsPointer("2025-01-31"),
				LabelID:          nil,
			},
		},
		{
			Description:              "Monthly with 12 months relative offset",
			DatabaseTime:             "2025-11-26",
			ExpectedCalculatedAmount: 15_00,
			// Should be the same independent of the date since it still counts as 1 year
			ExpectedNextCost:              180_00,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           15_00,
				DistributionType: "employer",
				RelativeOffset:   12,
				TargetDate:       utils.StringAsPointer("2025-01-31"),
				LabelID:          nil,
			},
		},
		{
			Description:                "Monthly with 12 months relative offset",
			DatabaseTime:               "2026-02-01",
			ExpectedCalculatedAmount:   15_00,
			ExpectedNextCost:           180_00,
			ExpectedEmployeeDeductions: 0,
			// It should shift one year further
			ExpectedNextExecutionDate:     "2027-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           15_00,
				DistributionType: "employer",
				RelativeOffset:   12,
				TargetDate:       utils.StringAsPointer("2025-01-31"),
				LabelID:          nil,
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

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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

func TestSalaryCostWithPastPaymentsRelativeOffset(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "past-offset@liquiswiss.com", "test", "Past Offset Org",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(apiService, user.ID, "Past Offset Employee")
	assert.NoError(t, err)

	salary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-26",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	err = SetDatabaseTime(conn, "2026-03-15")
	assert.NoError(t, err)

	parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, "2026-03-15")
	assert.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)
	defer func() {
		utils.DefaultClock.SetFixedTime(nil)
	}()

	createPayload := models.CreateSalaryCost{
		Cycle:            "monthly",
		AmountType:       "fixed",
		Amount:           450_00,
		DistributionType: "employee",
		RelativeOffset:   6,
		TargetDate:       nil,
		LabelID:          nil,
	}

	salaryCost, err := apiService.CreateSalaryCost(createPayload, user.ID, salary.ID)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(salaryCost.CalculatedCostDetails), 2)
	firstDetail := salaryCost.CalculatedCostDetails[0]
	assert.Equal(t, "2026-03", firstDetail.Month)
	assert.True(t, firstDetail.IsExtraMonth)
	assert.Equal(t, uint(6), firstDetail.Divider)
	assert.Equal(t, int64(450_00*6), int64(firstDetail.Amount))

	secondDetail := salaryCost.CalculatedCostDetails[1]
	assert.Equal(t, "2026-09", secondDetail.Month)
	assert.False(t, secondDetail.IsExtraMonth)
	assert.Equal(t, uint(6), secondDetail.Divider)
	assert.Equal(t, int64(450_00*6), int64(secondDetail.Amount))

	assert.Equal(t, int64(450_00), int64(salaryCost.CalculatedAmount))
	assert.Equal(t, int64(450_00*6), int64(salaryCost.CalculatedNextCost))
	assert.Equal(t, "2026-09-01", salaryCost.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))

	// When including past salary payments, the next execution should point to the current block.
	costIncludingPast, err := apiService.GetSalaryCost(user.ID, salaryCost.ID, false)
	assert.NoError(t, err)
	assert.Equal(t, "2026-03-01", costIncludingPast.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))
	assert.Equal(t, int64(450_00), int64(costIncludingPast.CalculatedAmount))
	assert.Equal(t, int64(450_00*6), int64(costIncludingPast.CalculatedNextCost))
	assert.True(t, costIncludingPast.CalculatedCostDetails[0].IsExtraMonth)

	listWithPast, _, err := apiService.ListSalaryCosts(user.ID, salary.ID, 1, 10, false)
	assert.NoError(t, err)
	if assert.Len(t, listWithPast, 1) {
		assert.Equal(t, "2026-03-01", listWithPast[0].CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))
		assert.Equal(t, int64(450_00*6), int64(listWithPast[0].CalculatedNextCost))
	}

	err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
	assert.NoError(t, err)
}

func TestSalaryCostWithPastPaymentsAndTargetDate(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "target-date@liquiswiss.com", "test", "Target Date Org",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(apiService, user.ID, "Target Date Employee")
	assert.NoError(t, err)

	salary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-26",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	err = SetDatabaseTime(conn, "2026-05-20")
	assert.NoError(t, err)

	parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, "2026-05-20")
	assert.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)
	defer func() {
		utils.DefaultClock.SetFixedTime(nil)
	}()

	createPayload := models.CreateSalaryCost{
		Cycle:            "monthly",
		AmountType:       "fixed",
		Amount:           300_00,
		DistributionType: "employer",
		RelativeOffset:   12,
		TargetDate:       utils.StringAsPointer("2025-01-26"),
		LabelID:          nil,
	}

	salaryCost, err := apiService.CreateSalaryCost(createPayload, user.ID, salary.ID)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(salaryCost.CalculatedCostDetails), 2)
	firstDetail := salaryCost.CalculatedCostDetails[0]
	assert.Equal(t, "2026-01", firstDetail.Month)
	assert.True(t, firstDetail.IsExtraMonth)
	assert.Equal(t, uint(12), firstDetail.Divider)
	assert.Equal(t, int64(300_00*12), int64(firstDetail.Amount))

	assert.Equal(t, int64(300_00), int64(salaryCost.CalculatedAmount))
	assert.Equal(t, int64(300_00*12), int64(salaryCost.CalculatedNextCost))
	assert.Equal(t, "2027-01-01", salaryCost.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))

	costIncludingPast, err := apiService.GetSalaryCost(user.ID, salaryCost.ID, false)
	assert.NoError(t, err)
	assert.Equal(t, "2026-01-01", costIncludingPast.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))
	assert.Equal(t, int64(300_00*12), int64(costIncludingPast.CalculatedNextCost))

	listWithPast, _, err := apiService.ListSalaryCosts(user.ID, salary.ID, 1, 10, false)
	assert.NoError(t, err)
	if assert.Len(t, listWithPast, 1) {
		assert.Equal(t, "2026-01-01", listWithPast[0].CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))
		assert.Equal(t, int64(300_00*12), int64(listWithPast[0].CalculatedNextCost))
	}

	err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
	assert.NoError(t, err)
}

func TestSalaryCostTargetDateLeapYearMonthly(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "leap-monthly@liquiswiss.com", "test", "Leap Monthly Org",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(apiService, user.ID, "Leap Monthly Employee")
	assert.NoError(t, err)

	salary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2024-01-31",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	err = SetDatabaseTime(conn, "2024-03-05")
	assert.NoError(t, err)

	parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, "2024-03-05")
	assert.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)
	defer func() {
		utils.DefaultClock.SetFixedTime(nil)
	}()

	createPayload := models.CreateSalaryCost{
		Cycle:            "monthly",
		AmountType:       "fixed",
		Amount:           275_00,
		DistributionType: "employer",
		RelativeOffset:   1,
		TargetDate:       utils.StringAsPointer("2024-02-29"),
		LabelID:          nil,
	}

	salaryCost, err := apiService.CreateSalaryCost(createPayload, user.ID, salary.ID)
	assert.NoError(t, err)

	if assert.NotNil(t, salaryCost.CalculatedNextExecutionDate) {
		assert.Equal(t, "2024-03-01", salaryCost.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))
	}
	assert.Equal(t, int64(275_00), int64(salaryCost.CalculatedAmount))
	assert.Equal(t, int64(275_00), int64(salaryCost.CalculatedNextCost))

	assert.GreaterOrEqual(t, len(salaryCost.CalculatedCostDetails), 3)
	firstDetail := salaryCost.CalculatedCostDetails[0]
	assert.Equal(t, "2024-02", firstDetail.Month)
	assert.True(t, firstDetail.IsExtraMonth)
	assert.Equal(t, uint(1), firstDetail.Divider)
	assert.Equal(t, int64(275_00), int64(firstDetail.Amount))

	foundFebNonLeap := false
	for _, detail := range salaryCost.CalculatedCostDetails {
		if detail.Month == "2025-02" {
			foundFebNonLeap = true
			assert.Equal(t, uint(1), detail.Divider)
			assert.Equal(t, int64(275_00), int64(detail.Amount))
			break
		}
	}
	assert.True(t, foundFebNonLeap, "expected a detail for February in the non-leap year 2025")

	costIncludingPast, err := apiService.GetSalaryCost(user.ID, salaryCost.ID, false)
	assert.NoError(t, err)
	if assert.NotNil(t, costIncludingPast.CalculatedNextExecutionDate) {
		assert.Equal(t, "2024-02-01", costIncludingPast.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))
	}
	assert.Equal(t, int64(275_00), int64(costIncludingPast.CalculatedAmount))
	assert.Equal(t, int64(275_00), int64(costIncludingPast.CalculatedNextCost))

	err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
	assert.NoError(t, err)
}

func TestSalaryCostTargetDateLeapYearAnnualOffset(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "leap-annual@liquiswiss.com", "test", "Leap Annual Org",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(apiService, user.ID, "Leap Annual Employee")
	assert.NoError(t, err)

	salary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2024-01-31",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	err = SetDatabaseTime(conn, "2025-03-10")
	assert.NoError(t, err)

	parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, "2025-03-10")
	assert.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)
	defer func() {
		utils.DefaultClock.SetFixedTime(nil)
	}()

	createPayload := models.CreateSalaryCost{
		Cycle:            "monthly",
		AmountType:       "fixed",
		Amount:           3600_00,
		DistributionType: "employer",
		RelativeOffset:   12,
		TargetDate:       utils.StringAsPointer("2024-02-29"),
		LabelID:          nil,
	}

	salaryCost, err := apiService.CreateSalaryCost(createPayload, user.ID, salary.ID)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(salaryCost.CalculatedCostDetails), 2)
	firstDetail := salaryCost.CalculatedCostDetails[0]
	assert.Equal(t, "2025-02", firstDetail.Month)
	assert.True(t, firstDetail.IsExtraMonth)
	assert.Equal(t, uint(12), firstDetail.Divider)
	assert.Equal(t, int64(3600_00*12), int64(firstDetail.Amount))

	secondDetail := salaryCost.CalculatedCostDetails[1]
	assert.Equal(t, "2026-02", secondDetail.Month)
	assert.False(t, secondDetail.IsExtraMonth)
	assert.Equal(t, uint(12), secondDetail.Divider)

	if assert.NotNil(t, salaryCost.CalculatedNextExecutionDate) {
		assert.Equal(t, "2026-02-01", salaryCost.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))
	}
	assert.Equal(t, int64(3600_00), int64(salaryCost.CalculatedAmount))
	assert.Equal(t, int64(3600_00*12), int64(salaryCost.CalculatedNextCost))

	costIncludingPast, err := apiService.GetSalaryCost(user.ID, salaryCost.ID, false)
	assert.NoError(t, err)
	if assert.NotNil(t, costIncludingPast.CalculatedNextExecutionDate) {
		assert.Equal(t, "2025-02-01", costIncludingPast.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))
	}
	assert.Equal(t, int64(3600_00), int64(costIncludingPast.CalculatedAmount))
	assert.Equal(t, int64(3600_00*12), int64(costIncludingPast.CalculatedNextCost))

	err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
	assert.NoError(t, err)
}

func TestSalaryCostPersistsAfterSalaryTransition(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "transition@liquiswiss.com", "test", "Transition Org",
	)
	assert.NoError(t, err)

	employee, err := CreateEmployee(apiService, user.ID, "Transition Employee")
	assert.NoError(t, err)

	err = SetDatabaseTime(conn, "2025-03-15")
	assert.NoError(t, err)

	parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, "2025-03-15")
	assert.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)

	salary1, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              9000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-26",
		ToDate:              utils.StringAsPointer("2025-06-26"),
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	initialCost, err := apiService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            "monthly",
		AmountType:       "fixed",
		Amount:           400_00,
		DistributionType: "employee",
		RelativeOffset:   3,
		TargetDate:       nil,
		LabelID:          nil,
	}, user.ID, salary1.ID)
	assert.NoError(t, err)

	preTransitionCost, err := apiService.GetSalaryCost(user.ID, initialCost.ID, false)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(preTransitionCost.CalculatedCostDetails), 2)
	assert.Equal(t, "2025-03", preTransitionCost.CalculatedCostDetails[0].Month)
	assert.Equal(t, uint(2), preTransitionCost.CalculatedCostDetails[0].Divider)
	assert.Equal(t, int64(preTransitionCost.CalculatedCostDetails[0].Amount), int64(preTransitionCost.CalculatedCostDetails[0].Divider)*int64(400_00))

	err = SetDatabaseTime(conn, "2025-07-05")
	assert.NoError(t, err)

	nextParsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, "2025-07-05")
	assert.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&nextParsedDatabaseTime)

	salary2, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       165,
		Amount:              9500_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-06-27",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)
	assert.NotNil(t, salary2)

	err = SetDatabaseTime(conn, "2025-08-10")
	assert.NoError(t, err)

	finalParsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, "2025-08-10")
	assert.NoError(t, err)
	utils.DefaultClock.SetFixedTime(&finalParsedDatabaseTime)
	defer func() {
		utils.DefaultClock.SetFixedTime(nil)
	}()

	postTransitionCost, err := apiService.GetSalaryCost(user.ID, initialCost.ID, false)
	assert.NoError(t, err)

	if assert.NotNil(t, postTransitionCost.CalculatedNextExecutionDate) {
		assert.Equal(t, "2025-06-01", postTransitionCost.CalculatedNextExecutionDate.ToFormattedTime(utils.InternalDateFormat))
	}
	assert.Equal(t, int64(400_00), int64(postTransitionCost.CalculatedAmount))
	assert.Equal(t, int64(400_00*3), int64(postTransitionCost.CalculatedNextCost))

	assert.GreaterOrEqual(t, len(postTransitionCost.CalculatedCostDetails), 1)
	var detailMonths []string
	var juneDetail *models.SalaryCostDetail
	for i := range postTransitionCost.CalculatedCostDetails {
		detail := postTransitionCost.CalculatedCostDetails[i]
		detailMonths = append(detailMonths, detail.Month)
		if detail.Month == "2025-06" {
			juneDetail = &detail
		}
	}
	if assert.NotNilf(t, juneDetail, "expected salary cost detail for June 2025; got %v", detailMonths) {
		assert.Equal(t, uint(3), juneDetail.Divider)
		assert.Equal(t, int64(400_00*3), int64(juneDetail.Amount))
	}

	err = apiService.DeleteSalaryCost(user.ID, initialCost.ID)
	assert.NoError(t, err)
}

func TestCopySalaryCostsAcrossEmployees(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currency, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	assert.NoError(t, err)

	user, _, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "copy-source@liquiswiss.com", "test", "Copy Org",
	)
	assert.NoError(t, err)

	sourceEmployee, err := CreateEmployee(apiService, user.ID, "Source Employee")
	assert.NoError(t, err)

	targetEmployee, err := CreateEmployee(apiService, user.ID, "Target Employee")
	assert.NoError(t, err)

	sourceSalary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              7000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-01",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, sourceEmployee.ID)
	assert.NoError(t, err)

	targetSalary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              6500_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-03-01",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, targetEmployee.ID)
	assert.NoError(t, err)

	// Existing cost on target salary should be replaced
	originalTargetCost, err := apiService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            "monthly",
		AmountType:       "fixed",
		Amount:           999_00,
		DistributionType: "employee",
		RelativeOffset:   1,
		TargetDate:       nil,
		LabelID:          nil,
	}, user.ID, targetSalary.ID)
	assert.NoError(t, err)

	createdCost, err := apiService.CreateSalaryCost(models.CreateSalaryCost{
		Cycle:            "monthly",
		AmountType:       "fixed",
		Amount:           250_00,
		DistributionType: "employee",
		RelativeOffset:   2,
		TargetDate:       nil,
		LabelID:          nil,
	}, user.ID, sourceSalary.ID)
	assert.NoError(t, err)

	sourceSalaryID := sourceSalary.ID
	err = apiService.CopySalaryCosts(models.CopySalaryCosts{
		SourceSalaryID: &sourceSalaryID,
	}, user.ID, targetSalary.ID)
	assert.NoError(t, err)

	copiedCosts, _, err := apiService.ListSalaryCosts(user.ID, targetSalary.ID, 1, 10, true)
	assert.NoError(t, err)
	assert.Len(t, copiedCosts, 1)
	assert.NotEqual(t, createdCost.ID, copiedCosts[0].ID)
	assert.EqualValues(t, createdCost.Amount, copiedCosts[0].Amount)
	assert.Equal(t, createdCost.Cycle, copiedCosts[0].Cycle)
	assert.Equal(t, createdCost.RelativeOffset, copiedCosts[0].RelativeOffset)
	assert.NotEqual(t, originalTargetCost.ID, copiedCosts[0].ID)
}

func TestMonthlySalaryInBetweenMonthWithoutToDate(t *testing.T) {
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

	salaryCostLabel, err := CreateSalaryCostLabel(apiService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-26",
		ToDate:              nil,
		// We want to test separate costs
		WithSeparateCosts: true,
	}, user.ID, employee.ID)
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

	testCases := []TestCase{
		// Once
		{
			Description:                   "Once simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_000,
			ExpectedNextCost:              500_000,
			ExpectedEmployeeDeductions:    500_000,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "once",
				AmountType:       "fixed",
				Amount:           500_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       utils.StringAsPointer("2025-02-15"),
				LabelID:          nil,
			},
		},
		{
			Description:                   "Once simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "once",
				AmountType:       "percentage",
				Amount:           25_000,
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
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly: Current date after paydate shifts next execution date further",
			DatabaseTime:                  "2025-01-28",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-03-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           25_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly with offset: On day of salary payment",
			DatabaseTime:                  "2025-06-26",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly with offset: On day after salary payment",
			DatabaseTime:                  "2025-06-27",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2026-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset percentage",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      7500_00,
			ExpectedNextCost:              7500_00 * 7,
			ExpectedEmployeeDeductions:    7500_00,
			ExpectedNextExecutionDate:     "2026-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           75_000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              uint64(250_00 * utils.GetTotalMonthsForMaxForecastYears()),
			ExpectedEmployeeDeductions:    250_00,
			ExpectedNextExecutionDate:     "2085-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Quarterly
		{
			Description:                   "Quarterly simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00 / 3,
			ExpectedNextExecutionDate:     "2025-04-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00 / 3,
			ExpectedNextExecutionDate:     "2025-04-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "percentage",
				Amount:           25_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly offset fixed before salary",
			DatabaseTime:                  "2025-06-25",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00 / 3,
			ExpectedNextExecutionDate:     "2027-03-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly offset fixed after salary",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00 / 3,
			ExpectedNextExecutionDate:     "2027-04-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly offset percentage",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      7500_00,
			ExpectedNextCost:              7500_00 * 7,
			ExpectedEmployeeDeductions:    7500_00 / 3,
			ExpectedNextExecutionDate:     "2027-04-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "percentage",
				Amount:           75_000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Quarterly giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              uint64(250_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/3)),
			ExpectedEmployeeDeductions:    250_00 / 3,
			ExpectedNextExecutionDate:     "2205-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "quarterly",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Biannually
		{
			Description:                   "Biannually simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00 / 6,
			ExpectedNextExecutionDate:     "2025-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00 / 6,
			ExpectedNextExecutionDate:     "2025-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "percentage",
				Amount:           25_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually offset fixed before salary",
			DatabaseTime:                  "2025-06-25",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00 / 6,
			ExpectedNextExecutionDate:     "2028-12-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually offset fixed after salary",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00 / 6,
			ExpectedNextExecutionDate:     "2029-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually offset percentage",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      7500_00,
			ExpectedNextCost:              7500_00 * 7,
			ExpectedEmployeeDeductions:    7500_00 / 6,
			ExpectedNextExecutionDate:     "2029-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "percentage",
				Amount:           75_000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Biannually giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              uint64(250_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/6)),
			ExpectedEmployeeDeductions:    250_00 / 6,
			ExpectedNextExecutionDate:     "2385-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "biannually",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},

		// Yearly
		{
			Description:                   "Yearly simple fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00 / 12,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly simple percentage",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2500_00,
			ExpectedNextCost:              2500_00,
			ExpectedEmployeeDeductions:    2500_00 / 12,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "percentage",
				Amount:           25_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly offset fixed before salary",
			DatabaseTime:                  "2025-06-25",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              uint64(500_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/12)),
			ExpectedEmployeeDeductions:    500_00 / 12,
			ExpectedNextExecutionDate:     "2032-06-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly offset fixed after salary",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              uint64(500_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/12)),
			ExpectedEmployeeDeductions:    500_00 / 12,
			ExpectedNextExecutionDate:     "2032-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly offset percentage",
			DatabaseTime:                  "2025-07-01",
			ExpectedCalculatedAmount:      7500_00,
			ExpectedNextCost:              uint64(7500_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/12)),
			ExpectedEmployeeDeductions:    7500_00 / 12,
			ExpectedNextExecutionDate:     "2032-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "percentage",
				Amount:           75_000,
				DistributionType: "employee",
				RelativeOffset:   7,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Yearly giga offset fixed",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              uint64(250_00 * math.Ceil(utils.GetTotalMonthsForMaxForecastYears()/12)),
			ExpectedEmployeeDeductions:    250_00 / 12,
			ExpectedNextExecutionDate:     "2745-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "yearly",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   720,
				TargetDate:       nil,
				LabelID:          nil,
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

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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

func TestMonthlySalaryInBetweenMonthWithToDate(t *testing.T) {
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

	label, err := CreateSalaryCostLabel(apiService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salary, err := apiService.CreateSalary(models.CreateSalary{
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
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple percentage before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2000_00,
			ExpectedNextCost:              2000_00,
			ExpectedEmployeeDeductions:    2000_00,
			ExpectedNextExecutionDate:     "2025-02-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           20_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple fixed passed salary end but in range",
			DatabaseTime:                  "2025-08-25",
			ExpectedCalculatedAmount:      250_00,
			ExpectedNextCost:              250_00,
			ExpectedEmployeeDeductions:    250_00,
			ExpectedNextExecutionDate:     "2025-08-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           250_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly simple percentage passed salary end but in range",
			DatabaseTime:                  "2025-08-25",
			ExpectedCalculatedAmount:      3500_00,
			ExpectedNextCost:              3500_00,
			ExpectedEmployeeDeductions:    3500_00,
			ExpectedNextExecutionDate:     "2025-08-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           35_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly simple fixed after salary out of range",
			// Mind one day after a month after the salary ended
			DatabaseTime:                  "2025-09-01",
			ExpectedCalculatedAmount:      0,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           300_00,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly simple percentage after salary out of range",
			// Mind one day after a month after the salary ended
			DatabaseTime:                  "2025-09-01",
			ExpectedCalculatedAmount:      0,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           15_000,
				DistributionType: "employee",
				RelativeOffset:   1,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly offset fixed before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00 * 5,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-06-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   5,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:                   "Monthly percentage fixed before salary ends in range",
			DatabaseTime:                  "2025-01-01",
			ExpectedCalculatedAmount:      2000_00,
			ExpectedNextCost:              2000_00 * 5,
			ExpectedEmployeeDeductions:    2000_00,
			ExpectedNextExecutionDate:     "2025-06-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           20_000,
				DistributionType: "employee",
				RelativeOffset:   5,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:              "Monthly offset fixed passed salary end but in range",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 500_00,
			// Because salary ends on 26.07 we have a maximum of 7 months
			ExpectedNextCost:              500_00 * 7,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-11-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description:              "Monthly offset percentage passed salary end but in range",
			DatabaseTime:             "2025-01-01",
			ExpectedCalculatedAmount: 3500_00,
			// Because salary ends on 26.07 we have a maximum of 7 months
			ExpectedNextCost:              3500_00 * 7,
			ExpectedEmployeeDeductions:    3500_00,
			ExpectedNextExecutionDate:     "2025-11-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           35_000,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly offset fixed passed salary end with current date but in range",
			// Because salary ends on 26.07 adding 10 months would be the last on 26.05 so we are in range
			DatabaseTime:                  "2026-05-25",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2026-05-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly offset percentage passed salary end with current date but in range",
			// Because salary ends on 26.07 adding 10 months would be the last on 26.05 so we are in range
			DatabaseTime:                  "2026-05-25",
			ExpectedCalculatedAmount:      1500_00,
			ExpectedNextCost:              1500_00,
			ExpectedEmployeeDeductions:    1500_00,
			ExpectedNextExecutionDate:     "2026-05-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           15_000,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly offset fixed passed salary end with current date out of range",
			// Because salary ends on 26.07 adding 10 months would be the last on 26.05
			DatabaseTime:                  "2026-06-01",
			ExpectedCalculatedAmount:      0,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
			},
		},
		{
			Description: "Monthly offset percentage passed salary end with current date out of range",
			// Because salary ends on 26.07 adding 10 months would be the last on 26.05
			DatabaseTime:                  "2026-06-01",
			ExpectedCalculatedAmount:      0,
			ExpectedNextCost:              0,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "percentage",
				Amount:           15_000,
				DistributionType: "employee",
				RelativeOffset:   10,
				TargetDate:       nil,
				LabelID:          nil,
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

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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

func TestMultipleSalaryInBetweenMonthCases(t *testing.T) {
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

	salaryCostLabel, err := CreateSalaryCostLabel(apiService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salary1, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              7_500 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 29,
		FromDate:            "2025-01-26",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
	assert.NoError(t, err)

	salary2, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10_000 * 100,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 29,
		FromDate:            "2025-09-26",
		ToDate:              nil,
		WithSeparateCosts:   true,
	}, user.ID, employee.ID)
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

	testCases1 := []TestCase{
		{
			Description:                   "BVG 2. Quartal",
			DatabaseTime:                  "2025-07-09",
			ExpectedCalculatedAmount:      200_00,
			ExpectedNextCost:              600_00,
			ExpectedEmployeeDeductions:    200_00,
			ExpectedNextExecutionDate:     "2025-07-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           200_00,
				DistributionType: "employee",
				RelativeOffset:   3,
				TargetDate:       utils.StringAsPointer("2025-01-26"),
				LabelID:          &salaryCostLabel.ID,
			},
		},
		{
			Description:                   "BVG 3. Quartal",
			DatabaseTime:                  "2025-08-01",
			ExpectedCalculatedAmount:      200_00,
			ExpectedNextCost:              400_00,
			ExpectedEmployeeDeductions:    200_00,
			ExpectedNextExecutionDate:     "2025-10-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           200_00,
				DistributionType: "employee",
				RelativeOffset:   3,
				TargetDate:       utils.StringAsPointer("2025-01-26"),
				LabelID:          &salaryCostLabel.ID,
			},
		},
	}

	testCases2 := []TestCase{
		{
			Description:                   "BVG 3. Quartal Neues Gehalt",
			DatabaseTime:                  "2025-07-25",
			ExpectedCalculatedAmount:      500_00,
			ExpectedNextCost:              500_00,
			ExpectedEmployeeDeductions:    500_00,
			ExpectedNextExecutionDate:     "2025-10-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           500_00,
				DistributionType: "employee",
				RelativeOffset:   3,
				TargetDate:       utils.StringAsPointer("2025-01-26"),
				LabelID:          &salaryCostLabel.ID,
			},
		},
	}

	for _, testCase := range testCases1 {
		t.Run(testCase.Description, func(t *testing.T) {
			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, testCase.DatabaseTime)
			assert.NoError(t, err)

			utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)
			defer func() {
				utils.DefaultClock.SetFixedTime(nil)
			}()

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary1.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary1.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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

	for _, testCase := range testCases2 {
		t.Run(testCase.Description, func(t *testing.T) {
			err = SetDatabaseTime(conn, testCase.DatabaseTime)
			assert.NoError(t, err)

			parsedDatabaseTime, err := time.Parse(utils.InternalDateFormat, testCase.DatabaseTime)
			assert.NoError(t, err)

			utils.DefaultClock.SetFixedTime(&parsedDatabaseTime)
			defer func() {
				utils.DefaultClock.SetFixedTime(nil)
			}()

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary2.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary2.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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

func TestLongOffsetScenariosInBetweenMonth(t *testing.T) {
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

	salaryCostLabel, err := CreateSalaryCostLabel(apiService, user.ID, "Test Label")
	assert.NoError(t, err)

	// Tests
	salary, err := apiService.CreateSalary(models.CreateSalary{
		HoursPerMonth:       160,
		Amount:              10000_00,
		Cycle:               "monthly",
		CurrencyID:          *currency.ID,
		VacationDaysPerYear: 25,
		FromDate:            "2025-01-26",
		ToDate:              nil,
		// We want to test separate costs
		WithSeparateCosts: true,
	}, user.ID, employee.ID)
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

	testCases := []TestCase{
		{
			Description:                   "Monthly with 12 months relative offset",
			DatabaseTime:                  "2025-01-26",
			ExpectedCalculatedAmount:      15_00,
			ExpectedNextCost:              180_00,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           15_00,
				DistributionType: "employer",
				RelativeOffset:   12,
				TargetDate:       utils.StringAsPointer("2025-01-26"),
				LabelID:          nil,
			},
		},
		{
			Description:              "Monthly with 12 months relative offset",
			DatabaseTime:             "2025-11-26",
			ExpectedCalculatedAmount: 15_00,
			// Should be the same independent of the date since it still counts as 1 year
			ExpectedNextCost:              180_00,
			ExpectedEmployeeDeductions:    0,
			ExpectedNextExecutionDate:     "2026-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           15_00,
				DistributionType: "employer",
				RelativeOffset:   12,
				TargetDate:       utils.StringAsPointer("2025-01-26"),
				LabelID:          nil,
			},
		},
		{
			Description:                "Monthly with 12 months relative offset",
			DatabaseTime:               "2026-02-01",
			ExpectedCalculatedAmount:   15_00,
			ExpectedNextCost:           180_00,
			ExpectedEmployeeDeductions: 0,
			// It should shift one year further
			ExpectedNextExecutionDate:     "2027-01-01",
			ExpectedPreviousExecutionDate: "",
			CreateData: models.CreateSalaryCost{
				Cycle:            "monthly",
				AmountType:       "fixed",
				Amount:           15_00,
				DistributionType: "employer",
				RelativeOffset:   12,
				TargetDate:       utils.StringAsPointer("2025-01-26"),
				LabelID:          nil,
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

			salaryCost, err := apiService.CreateSalaryCost(testCase.CreateData, user.ID, salary.ID)
			assert.NoError(t, err)

			// Check deduction
			updatedSalary, err := apiService.GetSalary(user.ID, salary.ID)
			assert.NoError(t, err)

			assert.Equal(t, int64(testCase.ExpectedEmployeeDeductions), int64(updatedSalary.EmployeeDeductions), "updatedSalary.EmployeeDeductions")

			err = apiService.DeleteSalaryCost(user.ID, salaryCost.ID)
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
