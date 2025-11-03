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
		WithSeparateCosts:   false,
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
		WithSeparateCosts:   false,
		IsTermination:       false,
		IsDisabled:          true,
	}

	mockDB.EXPECT().
		ListSalaryCosts(userID, activeSalary.ID, int64(1), int64(1000)).
		Return([]models.SalaryCost{}, int64(0), nil)

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
