package api_service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"liquiswiss/internal/mocks"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
)

type stubClock struct {
	fixed time.Time
}

func (c *stubClock) SetFixedTime(t *time.Time) {
	if t != nil {
		c.fixed = *t
	}
}

func (c *stubClock) Today() time.Time {
	return c.fixed
}

func TestCalculateForecast_SkipsDisabledTransactions(t *testing.T) {
	utils.InitValidator()

	userID := int64(99)
	fixedToday := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	originalClock := utils.DefaultClock
	utils.DefaultClock = &stubClock{fixed: fixedToday}
	defer func() {
		utils.DefaultClock = originalClock
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockIDatabaseAdapter(ctrl)
	service := api_service.NewAPIService(mockDB, nil)

	baseCode := "CHF"
	localeCode := "de-CH"
	orgCurrency := models.Currency{
		Code:       &baseCode,
		LocaleCode: &localeCode,
	}
	user := models.User{
		ID:                    userID,
		Name:                  "Test User",
		Email:                 "test@example.com",
		CurrentOrganisationID: 500,
		Currency:              orgCurrency,
	}
	organisation := models.Organisation{
		ID:       user.CurrentOrganisationID,
		Name:     "Org",
		Currency: orgCurrency,
	}
	mockDB.EXPECT().
		GetProfile(userID).
		Return(&user, nil)
	mockDB.EXPECT().
		GetOrganisation(userID, user.CurrentOrganisationID).
		Return(&organisation, nil)

	enabledStart := time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC)
	enabledStartDate := types.AsDate(enabledStart)

	transactions := []models.Transaction{
		{
			ID:          1,
			Name:        "Enabled Transaction",
			Amount:      100_00,
			VatIncluded: true,
			Type:        "single",
			StartDate:   enabledStartDate,
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  false,
		},
		{
			ID:          2,
			Name:        "Disabled Transaction",
			Amount:      500_00,
			VatIncluded: true,
			Type:        "single",
			StartDate:   enabledStartDate,
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  true,
		},
	}

	mockDB.EXPECT().
		ListTransactions(userID, int64(1), int64(100000), "name", "ASC").
		Return(transactions, int64(len(transactions)), nil)

	mockDB.EXPECT().
		ListFiatRates(baseCode).
		Return([]models.FiatRate{}, nil)

	mockDB.EXPECT().
		ListForecastExclusions(userID, int64(1), utils.TransactionsTableName).
		Return(map[string]bool{}, nil)

	mockDB.EXPECT().
		ListEmployees(userID, int64(1), int64(100000), "name", "ASC").
		Return([]models.Employee{}, int64(0), nil)

	mockDB.EXPECT().
		GetVatSetting(userID).
		Return(nil, nil)

	mockDB.EXPECT().
		ClearForecasts(userID).
		Return(int64(0), nil)

	var capturedForecast models.CreateForecast
	mockDB.EXPECT().
		UpsertForecast(gomock.Any(), userID).
		DoAndReturn(func(payload models.CreateForecast, _ int64) (int64, error) {
			capturedForecast = payload
			return 1, nil
		})

	mockDB.EXPECT().
		UpsertForecastDetail(gomock.Any(), userID, int64(1)).
		Return(int64(0), nil)

	mockDB.EXPECT().
		ListForecasts(userID, int64(utils.GetTotalMonthsForMaxForecastYears())).
		DoAndReturn(func(_ int64, _ int64) ([]models.Forecast, error) {
			return []models.Forecast{
				{
					Data: models.ForecastData{
						Month:    capturedForecast.Month,
						Revenue:  capturedForecast.Revenue,
						Expense:  capturedForecast.Expense,
						Cashflow: capturedForecast.Cashflow,
					},
				},
			}, nil
		})

	results, err := service.CalculateForecast(userID)
	require.NoError(t, err)

	require.Equal(t, enabledStart.Format("2006-01"), capturedForecast.Month)
	require.EqualValues(t, 100_00, capturedForecast.Revenue)
	require.EqualValues(t, 0, capturedForecast.Expense)
	require.EqualValues(t, capturedForecast.Revenue+capturedForecast.Expense, capturedForecast.Cashflow)

	require.Len(t, results, 1)
	require.Equal(t, capturedForecast.Month, results[0].Data.Month)
	require.EqualValues(t, capturedForecast.Revenue, results[0].Data.Revenue)
	require.EqualValues(t, capturedForecast.Expense, results[0].Data.Expense)
}

func TestCalculateForecast_SkipsDisabledSalariesOnly(t *testing.T) {
	utils.InitValidator()

	userID := int64(101)
	fixedToday := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	originalClock := utils.DefaultClock
	utils.DefaultClock = &stubClock{fixed: fixedToday}
	defer func() {
		utils.DefaultClock = originalClock
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockIDatabaseAdapter(ctrl)
	service := api_service.NewAPIService(mockDB, nil)

	baseCode := "CHF"
	localeCode := "de-CH"
	orgCurrency := models.Currency{
		Code:       &baseCode,
		LocaleCode: &localeCode,
	}
	user := models.User{
		ID:                    userID,
		Name:                  "Test User",
		Email:                 "test@example.com",
		CurrentOrganisationID: 600,
		Currency:              orgCurrency,
	}
	organisation := models.Organisation{
		ID:       user.CurrentOrganisationID,
		Name:     "Org",
		Currency: orgCurrency,
	}
	mockDB.EXPECT().
		GetProfile(userID).
		Return(&user, nil)
	mockDB.EXPECT().
		GetOrganisation(userID, user.CurrentOrganisationID).
		Return(&organisation, nil)

	// No transactions in this scenario
	mockDB.EXPECT().
		ListTransactions(userID, int64(1), int64(100000), "name", "ASC").
		Return([]models.Transaction{}, int64(0), nil)

	mockDB.EXPECT().
		ListFiatRates(baseCode).
		Return([]models.FiatRate{}, nil)

	employee := models.Employee{
		ID:   55,
		Name: "Employee A",
	}

	mockDB.EXPECT().
		ListEmployees(userID, int64(1), int64(100000), "name", "ASC").
		Return([]models.Employee{employee}, int64(1), nil)

	activeFrom := time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC)
	activeTo := time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC)
	activeFromDate := types.AsDate(activeFrom)
	activeToDate := types.AsDate(activeTo)

	activeSalary := models.Salary{
		ID:                  200,
		EmployeeID:          employee.ID,
		HoursPerMonth:       160,
		Amount:              500_00,
		Cycle:               utils.CycleMonthly,
		Currency:            orgCurrency,
		VacationDaysPerYear: 25,
		FromDate:            activeFromDate,
		ToDate:              &activeToDate,
		IsTermination:       false,
		IsDisabled:          false,
		EmployeeDeductions:  0,
		EmployerCosts:       0,
	}

	disabledSalary := models.Salary{
		ID:                  201,
		EmployeeID:          employee.ID,
		HoursPerMonth:       160,
		Amount:              800_00,
		Cycle:               utils.CycleMonthly,
		Currency:            orgCurrency,
		VacationDaysPerYear: 25,
		FromDate:            activeFromDate,
		ToDate:              &activeToDate,
		IsTermination:       false,
		IsDisabled:          true,
	}

	mockDB.EXPECT().
		ListSalaryCosts(userID, activeSalary.ID, int64(1), int64(1000)).
		Return([]models.SalaryCost{}, int64(0), nil).
		Times(2)

	mockDB.EXPECT().
		ListSalaryCosts(userID, disabledSalary.ID, int64(1), int64(1000)).
		Return([]models.SalaryCost{}, int64(0), nil)

	mockDB.EXPECT().
		ListSalaries(userID, employee.ID, int64(1), int64(100000)).
		Return([]models.Salary{activeSalary, disabledSalary}, int64(2), nil)

	mockDB.EXPECT().
		ListForecastExclusions(userID, activeSalary.ID, utils.SalariesTableName).
		Return(map[string]bool{}, nil)

	mockDB.EXPECT().
		GetVatSetting(userID).
		Return(nil, nil)

	mockDB.EXPECT().
		ClearForecasts(userID).
		Return(int64(0), nil)

	var capturedForecast models.CreateForecast
	mockDB.EXPECT().
		UpsertForecast(gomock.Any(), userID).
		DoAndReturn(func(payload models.CreateForecast, _ int64) (int64, error) {
			capturedForecast = payload
			return 1, nil
		})

	mockDB.EXPECT().
		UpsertForecastDetail(gomock.Any(), userID, int64(1)).
		Return(int64(0), nil)

	mockDB.EXPECT().
		ListForecasts(userID, int64(utils.GetTotalMonthsForMaxForecastYears())).
		DoAndReturn(func(_ int64, _ int64) ([]models.Forecast, error) {
			return []models.Forecast{
				{
					Data: models.ForecastData{
						Month:    capturedForecast.Month,
						Revenue:  capturedForecast.Revenue,
						Expense:  capturedForecast.Expense,
						Cashflow: capturedForecast.Cashflow,
					},
				},
			}, nil
		})

	results, err := service.CalculateForecast(userID)
	require.NoError(t, err)

	require.Equal(t, activeFrom.Format("2006-01"), capturedForecast.Month)
	require.EqualValues(t, 0, capturedForecast.Revenue)
	require.EqualValues(t, -int64(activeSalary.Amount), capturedForecast.Expense)
	require.EqualValues(t, capturedForecast.Revenue+capturedForecast.Expense, capturedForecast.Cashflow)

	require.Len(t, results, 1)
	require.Equal(t, capturedForecast.Month, results[0].Data.Month)
	require.EqualValues(t, capturedForecast.Expense, results[0].Data.Expense)
}

func TestCalculateForecast_CountsBothSalaryCostsTwice(t *testing.T) {
	logger.Logger = zap.NewNop().Sugar()
	utils.InitValidator()

	fixedToday := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	originalClock := utils.DefaultClock
	utils.DefaultClock = &stubClock{fixed: fixedToday}
	defer func() {
		utils.DefaultClock = originalClock
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockIDatabaseAdapter(ctrl)
	service := api_service.NewAPIService(mockDB, nil)

	userID := int64(303)
	baseCode := "CHF"
	localeCode := "de-CH"

	orgCurrency := models.Currency{
		Code:       &baseCode,
		LocaleCode: &localeCode,
	}
	user := models.User{
		ID:                    userID,
		Name:                  "Test User",
		Email:                 "test@example.com",
		CurrentOrganisationID: 808,
		Currency:              orgCurrency,
	}
	organisation := models.Organisation{
		ID:       user.CurrentOrganisationID,
		Name:     "Org",
		Currency: orgCurrency,
	}

	mockDB.EXPECT().
		GetProfile(userID).
		Return(&user, nil)
	mockDB.EXPECT().
		GetOrganisation(userID, user.CurrentOrganisationID).
		Return(&organisation, nil)

	mockDB.EXPECT().
		ListTransactions(userID, int64(1), int64(100000), "name", "ASC").
		Return([]models.Transaction{}, int64(0), nil)

	mockDB.EXPECT().
		ListFiatRates(baseCode).
		Return([]models.FiatRate{}, nil)

	employee := models.Employee{
		ID:   55,
		Name: "Employee Both",
	}

	mockDB.EXPECT().
		ListEmployees(userID, int64(1), int64(100000), "name", "ASC").
		Return([]models.Employee{employee}, int64(1), nil)

	activeFrom := time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC)
	activeFromDate := types.AsDate(activeFrom)
	activeToDate := types.AsDate(activeFrom)

	grossAmount := uint64(10_000_00)
	bothShare := uint64(510_00)

	activeSalary := models.Salary{
		ID:                  700,
		EmployeeID:          employee.ID,
		Amount:              grossAmount,
		Cycle:               utils.CycleMonthly,
		Currency:            orgCurrency,
		FromDate:            activeFromDate,
		ToDate:              &activeToDate,
		IsTermination:       false,
		IsDisabled:          false,
		EmployeeDeductions:  bothShare,
		EmployerCosts:       bothShare,
		VacationDaysPerYear: 25,
	}

	mockDB.EXPECT().
		ListSalaries(userID, employee.ID, int64(1), int64(100000)).
		Return([]models.Salary{activeSalary}, int64(1), nil)

	costFromDate := activeFrom
	costFromDateAsDate := types.AsDate(costFromDate)
	salaryCost := models.SalaryCost{
		ID:                          900,
		Cycle:                       utils.CycleMonthly,
		AmountType:                  "percentage",
		DistributionType:            "both",
		RelativeOffset:              1,
		SalaryID:                    activeSalary.ID,
		CalculatedNextExecutionDate: &costFromDateAsDate,
		CalculatedNextCost:          bothShare,
		DBDate:                      costFromDateAsDate,
		CalculatedCostDetails: []models.SalaryCostDetail{
			{
				Month:  activeFrom.Format("2006-01"),
				Amount: bothShare,
			},
		},
	}

	mockDB.EXPECT().
		ListForecastExclusions(userID, activeSalary.ID, utils.SalariesTableName).
		Return(map[string]bool{}, nil)

	mockDB.EXPECT().
		ListSalaryCosts(userID, activeSalary.ID, int64(1), int64(1000)).
		Return([]models.SalaryCost{salaryCost}, int64(1), nil).
		AnyTimes()

	mockDB.EXPECT().
		ListSalaryCostDetails(salaryCost.ID).
		Return([]models.SalaryCostDetail{
			{
				Month:  activeFrom.Format("2006-01"),
				Amount: bothShare,
				CostID: salaryCost.ID,
			},
		}, nil).
		AnyTimes()

	mockDB.EXPECT().
		ListForecastExclusions(userID, salaryCost.ID, utils.SalaryCostsTableName).
		Return(map[string]bool{}, nil)

	mockDB.EXPECT().
		GetVatSetting(userID).
		Return(nil, nil)

	mockDB.EXPECT().
		ClearForecasts(userID).
		Return(int64(0), nil)

	var capturedForecast models.CreateForecast
	mockDB.EXPECT().
		UpsertForecast(gomock.Any(), userID).
		DoAndReturn(func(payload models.CreateForecast, _ int64) (int64, error) {
			capturedForecast = payload
			return 1, nil
		})

	mockDB.EXPECT().
		UpsertForecastDetail(gomock.Any(), userID, int64(1)).
		Return(int64(0), nil)

	mockDB.EXPECT().
		ListForecasts(userID, int64(utils.GetTotalMonthsForMaxForecastYears())).
		DoAndReturn(func(_ int64, _ int64) ([]models.Forecast, error) {
			return []models.Forecast{
				{
					Data: models.ForecastData{
						Month:    capturedForecast.Month,
						Revenue:  capturedForecast.Revenue,
						Expense:  capturedForecast.Expense,
						Cashflow: capturedForecast.Cashflow,
					},
				},
			}, nil
		})

	results, err := service.CalculateForecast(userID)
	require.NoError(t, err)

	expectedExpense := -int64(grossAmount) - int64(bothShare*2)
	require.Equal(t, activeFrom.Format("2006-01"), capturedForecast.Month)
	require.EqualValues(t, expectedExpense, capturedForecast.Expense)
	require.Len(t, results, 1)
	require.EqualValues(t, expectedExpense, results[0].Data.Expense)
}

func TestUpdateForecastExclusions_Success(t *testing.T) {
	logger.Logger = zap.NewNop().Sugar()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockIDatabaseAdapter(ctrl)
	service := api_service.NewAPIService(mockDB, nil)

	userID := int64(42)
	payload := models.UpdateForecastExclusions{
		Updates: []models.ForecastExclusionUpdate{
			{
				Month:        "2024-01",
				RelatedID:    1,
				RelatedTable: utils.TransactionsTableName,
				IsExcluded:   true,
			},
			{
				Month:        "2024-02",
				RelatedID:    2,
				RelatedTable: utils.TransactionsTableName,
				IsExcluded:   false,
			},
		},
	}

	gomock.InOrder(
		mockDB.EXPECT().
			CreateForecastExclusion(models.CreateForecastExclusion{
				Month:        "2024-01",
				RelatedID:    1,
				RelatedTable: utils.TransactionsTableName,
			}, userID).
			Return(int64(10), nil),
		mockDB.EXPECT().
			DeleteForecastExclusion(models.CreateForecastExclusion{
				Month:        "2024-02",
				RelatedID:    2,
				RelatedTable: utils.TransactionsTableName,
			}, userID).
			Return(int64(1), nil),
	)

	err := service.UpdateForecastExclusions(payload, userID)
	require.NoError(t, err)
}

func TestUpdateForecastExclusions_PropagatesError(t *testing.T) {
	logger.Logger = zap.NewNop().Sugar()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockIDatabaseAdapter(ctrl)
	service := api_service.NewAPIService(mockDB, nil)

	userID := int64(7)
	payload := models.UpdateForecastExclusions{
		Updates: []models.ForecastExclusionUpdate{
			{
				Month:        "2024-03",
				RelatedID:    3,
				RelatedTable: utils.TransactionsTableName,
				IsExcluded:   true,
			},
		},
	}

	expectedErr := errors.New("create err")
	mockDB.EXPECT().
		CreateForecastExclusion(models.CreateForecastExclusion{
			Month:        "2024-03",
			RelatedID:    3,
			RelatedTable: utils.TransactionsTableName,
		}, userID).
		Return(int64(0), expectedErr)

	err := service.UpdateForecastExclusions(payload, userID)
	require.ErrorIs(t, err, expectedErr)
}

func TestCalculateForecast_VATSettlement_BiannuallyInterval(t *testing.T) {
	logger.Logger = zap.NewNop().Sugar()
	utils.InitValidator()

	// Today is 29.11.2025
	// VAT billing date (Rechnungszeitpunkt): 27.02.2026
	// VAT transaction date (Transaktionszeitpunkt): 28.02.2026
	// Interval: biannually (6 months)
	// Expected settlements:
	// - 28.02.2026: collects VAT until Jan 2026 (months BEFORE billing date 27.02.2026)
	// - 28.08.2026: collects VAT from Feb-Jul 2026 (next 6 months)
	fixedToday := time.Date(2025, time.November, 29, 0, 0, 0, 0, time.UTC)
	originalClock := utils.DefaultClock
	utils.DefaultClock = &stubClock{fixed: fixedToday}
	defer func() {
		utils.DefaultClock = originalClock
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockIDatabaseAdapter(ctrl)
	service := api_service.NewAPIService(mockDB, nil)

	userID := int64(999)
	baseCode := "CHF"
	localeCode := "de-CH"

	orgCurrency := models.Currency{
		Code:       &baseCode,
		LocaleCode: &localeCode,
	}
	user := models.User{
		ID:                    userID,
		Name:                  "VAT Test User",
		Email:                 "vattest@example.com",
		CurrentOrganisationID: 111,
		Currency:              orgCurrency,
	}
	organisation := models.Organisation{
		ID:       user.CurrentOrganisationID,
		Name:     "VAT Test Org",
		Currency: orgCurrency,
	}

	mockDB.EXPECT().
		GetProfile(userID).
		Return(&user, nil)
	mockDB.EXPECT().
		GetOrganisation(userID, user.CurrentOrganisationID).
		Return(&organisation, nil)

	// VAT Settings:
	// - billing_date = 27.02.2026 (Rechnungszeitpunkt - determines collection period)
	// - transaction_date = 28.02.2026 (Transaktionszeitpunkt - when payment appears in forecast)
	// - interval = biannually (6 months)
	// First period: collects until Jan 2026, appears on 28.02.2026
	// Second period: collects Feb-Jul 2026, appears on 28.08.2026
	vatBillingDate := time.Date(2026, time.February, 27, 0, 0, 0, 0, time.UTC)
	vatTransactionDate := time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC)
	vatSetting := &models.VatSetting{
		ID:              1,
		OrganisationID:  user.CurrentOrganisationID,
		Enabled:         true,
		BillingDate:     vatBillingDate,
		TransactionDate: vatTransactionDate,
		Interval:        "biannually",
	}

	mockDB.EXPECT().
		GetVatSetting(userID).
		Return(vatSetting, nil)

	// Create transactions with VAT from December 2025 to July 2026
	// Each transaction has a 7.7% VAT (770 basis points)
	vatValue := int64(770) // 7.7%
	vat := models.Vat{
		ID:    1,
		Value: vatValue,
	}

	transactions := []models.Transaction{
		// December 2025: 1000.00 CHF revenue + 77.00 VAT
		{
			ID:          1,
			Name:        "Dec Revenue",
			Amount:      1000_00,
			VatAmount:   77_00,
			Vat:         &vat,
			VatIncluded: false,
			Type:        "single",
			StartDate:   types.AsDate(time.Date(2025, time.December, 15, 0, 0, 0, 0, time.UTC)),
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  false,
		},
		// January 2026: 2000.00 CHF revenue + 154.00 VAT
		{
			ID:          2,
			Name:        "Jan Revenue",
			Amount:      2000_00,
			VatAmount:   154_00,
			Vat:         &vat,
			VatIncluded: false,
			Type:        "single",
			StartDate:   types.AsDate(time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)),
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  false,
		},
		// February 2026: 1500.00 CHF revenue + 115.50 VAT
		{
			ID:          3,
			Name:        "Feb Revenue",
			Amount:      1500_00,
			VatAmount:   115_50,
			Vat:         &vat,
			VatIncluded: false,
			Type:        "single",
			StartDate:   types.AsDate(time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)),
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  false,
		},
		// March 2026: 3000.00 CHF revenue + 231.00 VAT
		{
			ID:          4,
			Name:        "Mar Revenue",
			Amount:      3000_00,
			VatAmount:   231_00,
			Vat:         &vat,
			VatIncluded: false,
			Type:        "single",
			StartDate:   types.AsDate(time.Date(2026, time.March, 5, 0, 0, 0, 0, time.UTC)),
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  false,
		},
		// April 2026: 2500.00 CHF revenue + 192.50 VAT
		{
			ID:          5,
			Name:        "Apr Revenue",
			Amount:      2500_00,
			VatAmount:   192_50,
			Vat:         &vat,
			VatIncluded: false,
			Type:        "single",
			StartDate:   types.AsDate(time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC)),
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  false,
		},
		// May 2026: 1800.00 CHF revenue + 138.60 VAT
		{
			ID:          6,
			Name:        "May Revenue",
			Amount:      1800_00,
			VatAmount:   138_60,
			Vat:         &vat,
			VatIncluded: false,
			Type:        "single",
			StartDate:   types.AsDate(time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)),
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  false,
		},
		// June 2026: 2200.00 CHF revenue + 169.40 VAT
		{
			ID:          7,
			Name:        "Jun Revenue",
			Amount:      2200_00,
			VatAmount:   169_40,
			Vat:         &vat,
			VatIncluded: false,
			Type:        "single",
			StartDate:   types.AsDate(time.Date(2026, time.June, 8, 0, 0, 0, 0, time.UTC)),
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  false,
		},
		// July 2026: 2800.00 CHF revenue + 215.60 VAT
		{
			ID:          8,
			Name:        "Jul Revenue",
			Amount:      2800_00,
			VatAmount:   215_60,
			Vat:         &vat,
			VatIncluded: false,
			Type:        "single",
			StartDate:   types.AsDate(time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC)),
			Category:    models.Category{Name: "Sales"},
			Currency:    orgCurrency,
			IsDisabled:  false,
		},
	}

	mockDB.EXPECT().
		ListTransactions(userID, int64(1), int64(100000), "name", "ASC").
		Return(transactions, int64(len(transactions)), nil)

	mockDB.EXPECT().
		ListFiatRates(baseCode).
		Return([]models.FiatRate{}, nil)

	// Mock forecast exclusions for each transaction (none excluded)
	// Called twice per transaction: once for revenue, once for VAT collection
	for _, tx := range transactions {
		mockDB.EXPECT().
			ListForecastExclusions(userID, tx.ID, utils.TransactionsTableName).
			Return(map[string]bool{}, nil).
			Times(2)
	}

	mockDB.EXPECT().
		ListEmployees(userID, int64(1), int64(100000), "name", "ASC").
		Return([]models.Employee{}, int64(0), nil)

	mockDB.EXPECT().
		ClearForecasts(userID).
		Return(int64(0), nil)

	// Capture all upserted forecasts
	capturedForecasts := make(map[string]models.CreateForecast)
	mockDB.EXPECT().
		UpsertForecast(gomock.Any(), userID).
		DoAndReturn(func(payload models.CreateForecast, _ int64) (int64, error) {
			capturedForecasts[payload.Month] = payload
			return int64(len(capturedForecasts)), nil
		}).
		AnyTimes()

	mockDB.EXPECT().
		UpsertForecastDetail(gomock.Any(), userID, gomock.Any()).
		Return(int64(0), nil).
		AnyTimes()

	mockDB.EXPECT().
		ListForecasts(userID, int64(utils.GetTotalMonthsForMaxForecastYears())).
		DoAndReturn(func(_ int64, _ int64) ([]models.Forecast, error) {
			forecasts := make([]models.Forecast, 0, len(capturedForecasts))
			for _, cf := range capturedForecasts {
				forecasts = append(forecasts, models.Forecast{
					Data: models.ForecastData{
						Month:    cf.Month,
						Revenue:  cf.Revenue,
						Expense:  cf.Expense,
						Cashflow: cf.Cashflow,
					},
				})
			}
			return forecasts, nil
		})

	results, err := service.CalculateForecast(userID)
	require.NoError(t, err)
	require.NotEmpty(t, results)

	// Verify VAT settlement for 28.02.2026 (transaction date)
	// Billing date 27.02.2026 means we collect VAT UNTIL Jan 2026 (not including Feb)
	// Transaction date 28.02.2026 is when the expense appears in forecast
	feb2026 := "2026-02"
	require.Contains(t, capturedForecasts, feb2026, "Expected VAT settlement in Feb 2026")

	// Expected VAT for first settlement (until Jan 2026):
	// Dec: 77.00 + Jan: 154.00 = 231.00 CHF (Feb NOT included as it's >= billing date)
	expectedFirstVAT := int64(77_00 + 154_00)
	feb2026Forecast := capturedForecasts[feb2026]

	// The VAT should be a negative expense (money we owe)
	// Feb forecast should have:
	// - Revenue: Feb transaction amount + Feb VAT (1500.00 + 115.50)
	// - Expense: VAT settlement payment (collected VAT from Dec, Jan only)
	expectedFebRevenue := int64(1500_00 + 115_50) // Feb revenue + its VAT
	expectedVATExpense := -expectedFirstVAT       // VAT to pay as expense

	require.EqualValues(t, expectedFebRevenue, feb2026Forecast.Revenue,
		"Feb 2026 revenue should be Feb transaction + its VAT (expected: %d, got: %d)",
		expectedFebRevenue, feb2026Forecast.Revenue)

	// The expense should include the VAT settlement
	require.LessOrEqual(t, feb2026Forecast.Expense, expectedVATExpense,
		"Feb 2026 expense should include VAT settlement (expected at most: %d, got: %d)",
		expectedVATExpense, feb2026Forecast.Expense)

	// Verify VAT settlement for 28.08.2026 (transaction date)
	// Billing date 27.08.2026 means we collect VAT from Feb - Jul 2026 (6 months)
	aug2026 := "2026-08"
	require.Contains(t, capturedForecasts, aug2026, "Expected VAT settlement in Aug 2026")

	// Expected VAT for second settlement (Feb - Jul 2026):
	// Feb: 115.50 + Mar: 231.00 + Apr: 192.50 + May: 138.60 + Jun: 169.40 + Jul: 215.60 = 1062.60 CHF
	expectedSecondVAT := int64(115_50 + 231_00 + 192_50 + 138_60 + 169_40 + 215_60)
	aug2026Forecast := capturedForecasts[aug2026]

	// Aug should have VAT expense
	require.LessOrEqual(t, aug2026Forecast.Expense, -expectedSecondVAT,
		"Aug 2026 should have VAT settlement as expense (expected: %d, got: %d)",
		-expectedSecondVAT, aug2026Forecast.Expense)

	// Verify that the VAT settlements are in the future (after today)
	require.True(t, vatTransactionDate.After(fixedToday),
		"First VAT transaction date should be in the future")
	require.True(t, time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC).After(fixedToday),
		"Second VAT transaction date should be in the future")
}
