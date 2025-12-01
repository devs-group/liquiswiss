package handlers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
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

func TestCalculateForecast_VATSettlement_DBIntegration(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	// Fix "today" so all test transactions are in the future for revenue, but VAT collection includes them.
	fixedToday := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	originalClock := utils.DefaultClock
	utils.DefaultClock = &stubClock{fixed: fixedToday}
	defer func() {
		utils.DefaultClock = originalClock
	}()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currencyCHF, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	require.NoError(t, err)

	user, organisation, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "vat.integration@example.com", "test", "VAT Integration Org",
	)
	require.NoError(t, err)

	// Ensure organisation uses CHF to match test data.
	require.Equal(t, *currencyCHF.Code, *organisation.Currency.Code)

	category, err := apiService.CreateCategory(models.CreateCategory{Name: "Sales"}, &user.ID)
	require.NoError(t, err)

	vat, err := apiService.CreateVat(models.CreateVat{Value: 810}, user.ID)
	require.NoError(t, err)

	vatSetting, err := apiService.CreateVatSetting(models.CreateVatSetting{
		Enabled:                true,
		BillingDate:            "2026-01-01",
		TransactionMonthOffset: 1, // 1 month after billing date (January 2026 -> February 2026)
		Interval:               "biannually",
	}, user.ID)
	require.NoError(t, err)
	require.NotNil(t, vatSetting)

	monthly := utils.CycleMonthly
	vatID := vat.ID
	makeTx := func(name string, amount int64, startDate time.Time, opts ...func(*models.CreateTransaction)) *models.Transaction {
		return createTransaction(t, apiService, user.ID, category.ID, *currencyCHF.ID, nil,
			func(payload *models.CreateTransaction) {
				payload.Name = name
				payload.Amount = amount
				payload.StartDate = startDate.Format(utils.InternalDateFormat)
				payload.Vat = &vatID
				payload.VatIncluded = false
				payload.Cycle = &monthly
				payload.Type = "repeating"
				for _, opt := range opts {
					opt(payload)
				}
			},
		)
	}

	// Real transaction set from MariaDB fixture.
	txKevin := makeTx("Kevin Heim", 1920000, time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC),
		func(payload *models.CreateTransaction) {
			end := "2025-11-30"
			payload.EndDate = &end
		},
	)
	txMatthias := makeTx("Matthias Hillert-Wernicke", 1440000, time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC))
	txMarkus := makeTx("Markus Bauermeister", 1920000, time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC),
		func(payload *models.CreateTransaction) {
			end := "2026-01-31"
			payload.EndDate = &end
		},
	)
	txFlorian := makeTx("Florian Müller", 1920000, time.Date(2025, time.May, 31, 0, 0, 0, 0, time.UTC))
	makeTx("Ralph", 600000, time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC))
	makeTx("Robert", 600000, time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC))
	makeTx("Sebastian Fekete", 1920000, time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC))
	makeTx("Adrian Fedoreanu", 1920000, time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC))

	// Mirror production exclusions: Sep and Oct 2025 excluded for Kevin, Matthias, Markus, Florian
	exclusionMonths := []string{"2025-09", "2025-10"}
	exclusionTxs := []int64{txKevin.ID, txMatthias.ID, txMarkus.ID, txFlorian.ID}
	for _, month := range exclusionMonths {
		for _, txID := range exclusionTxs {
			_, err := dbAdapter.CreateForecastExclusion(models.CreateForecastExclusion{
				Month:        month,
				RelatedID:    txID,
				RelatedTable: utils.TransactionsTableName,
			}, user.ID)
			require.NoError(t, err)
		}
	}

	scenarios := []struct {
		name      string
		interval  string
		expectVAT map[string]int64
	}{
		{
			name:     "monthly",
			interval: "monthly",
			expectVAT: map[string]int64{
				"2026-02": -427_680, // Dec 2025 VAT (Matthias + Markus + Florian, Kevin ended Nov 30)
			},
		},
		{
			name:     "quarterly",
			interval: "quarterly",
			expectVAT: map[string]int64{
				"2026-02": -1_594_080, // Oct-Dec 2025 VAT
			},
		},
		{
			name:     "biannually",
			interval: "biannually",
			expectVAT: map[string]int64{
				"2026-02": -3_343_680, // Jul-Dec 2025 VAT
				"2026-08": -4_237_920, // Jan-Jun 2026 VAT
			},
		},
		{
			name:     "yearly",
			interval: "yearly",
			expectVAT: map[string]int64{
				"2026-02": -5_909_760, // Jan-Dec 2025 VAT
				"2027-02": -8_320_320, // Jan-Dec 2026 VAT
			},
		},
	}

	for _, sc := range scenarios {
		// Update interval per scenario
		_, err := apiService.UpdateVatSetting(models.UpdateVatSetting{
			Interval: &sc.interval,
		}, user.ID)
		require.NoError(t, err, sc.name)

		results, err := apiService.CalculateForecast(user.ID)
		require.NoError(t, err, sc.name)
		require.NotEmpty(t, results, sc.name)

		resultMap := make(map[string]models.ForecastData)
		for _, r := range results {
			resultMap[r.Data.Month] = r.Data
		}

		for month, exp := range sc.expectVAT {
			require.Contains(t, resultMap, month, sc.name+" month presence")
			require.EqualValues(t, exp, resultMap[month].Expense, sc.name+" VAT settlement amount mismatch")
		}
	}
}

func TestCalculateForecast_VATSettlement_MonthEndDates(t *testing.T) {
	conn := SetupTestEnvironment(t)
	defer conn.Close()

	// Fix "today" so transactions are in the future
	fixedToday := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	originalClock := utils.DefaultClock
	utils.DefaultClock = &stubClock{fixed: fixedToday}
	defer func() {
		utils.DefaultClock = originalClock
	}()

	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
	sendgridService := sendgrid_adapter.NewSendgridAdapter("")
	apiService := api_service.NewAPIService(dbAdapter, sendgridService)

	currencyCHF, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
	require.NoError(t, err)

	user, organisation, err := CreateUserWithOrganisation(
		apiService, dbAdapter, "vat.monthend@example.com", "test", "VAT MonthEnd Org",
	)
	require.NoError(t, err)
	require.Equal(t, *currencyCHF.Code, *organisation.Currency.Code)

	category, err := apiService.CreateCategory(models.CreateCategory{Name: "Sales"}, &user.ID)
	require.NoError(t, err)

	vat, err := apiService.CreateVat(models.CreateVat{Value: 810}, user.ID)
	require.NoError(t, err)

	// Test scenarios with different billing dates at month-end
	scenarios := []struct {
		name        string
		billingDate string
		offset      int
		expectMonth string // Expected settlement month
	}{
		{
			name:        "Jan 31 + 1 month",
			billingDate: "2026-01-31",
			offset:      1,
			expectMonth: "2026-02", // Should be Feb, not March!
		},
		{
			name:        "Jan 30 + 1 month",
			billingDate: "2026-01-30",
			offset:      1,
			expectMonth: "2026-02", // Should be Feb, not March!
		},
		{
			name:        "Jan 29 + 1 month",
			billingDate: "2026-01-29",
			offset:      1,
			expectMonth: "2026-02", // Should be Feb, not March!
		},
		{
			name:        "Mar 31 + 1 month",
			billingDate: "2026-03-31",
			offset:      1,
			expectMonth: "2026-04", // Should be April (30 days)
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			// Create VAT setting with specific billing date and offset
			vatSetting, err := apiService.CreateVatSetting(models.CreateVatSetting{
				Enabled:                true,
				BillingDate:            sc.billingDate,
				TransactionMonthOffset: sc.offset,
				Interval:               "monthly",
			}, user.ID)
			require.NoError(t, err)
			require.NotNil(t, vatSetting)

			// Create a simple transaction with VAT
			// Start in December 2025 to ensure VAT is collected in the period before billing
			monthly := utils.CycleMonthly
			vatID := vat.ID
			createTransaction(t, apiService, user.ID, category.ID, *currencyCHF.ID, nil,
				func(payload *models.CreateTransaction) {
					payload.Name = "Test Transaction"
					payload.Amount = 1000000 // 10,000 CHF
					payload.StartDate = "2025-12-01"
					payload.Vat = &vatID
					payload.VatIncluded = false
					payload.Cycle = &monthly
					payload.Type = "repeating"
				},
			)

			// Calculate forecast
			results, err := apiService.CalculateForecast(user.ID)
			require.NoError(t, err)
			require.NotEmpty(t, results)

			// Find VAT settlement
			resultMap := make(map[string]models.ForecastData)
			for _, r := range results {
				resultMap[r.Data.Month] = r.Data
			}

			// Verify settlement is in the expected month
			// The key assertion: with our fix, a billing date of Jan 31 + 1 month offset
			// should produce a settlement in February (not March, which would indicate overflow)
			require.Contains(t, resultMap, sc.expectMonth, "Settlement should be in "+sc.expectMonth)
			require.Less(t, resultMap[sc.expectMonth].Expense, int64(0), "Settlement should be negative (expense)")

			// Clean up for next test
			err = apiService.DeleteVatSetting(user.ID)
			require.NoError(t, err)
		})
	}
}

// func TestCalculateForecast_VATSettlement_PastDatesIgnored(t *testing.T) {
// 	conn := SetupTestEnvironment(t)
// 	defer conn.Close()
//
// 	fixedToday := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
// 	orig := utils.DefaultClock
// 	utils.DefaultClock = &stubClock{fixed: fixedToday}
// 	defer func() { utils.DefaultClock = orig }()
//
// 	dbAdapter := db_adapter.NewDatabaseAdapter(conn)
// 	apiService := api_service.NewAPIService(dbAdapter, sendgrid_adapter.NewSendgridAdapter(""))
//
// 	currencyCHF, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
// 	require.NoError(t, err)
// 	user, _, err := CreateUserWithOrganisation(apiService, dbAdapter, "vat.past@example.com", "test", "VAT Past Org")
// 	require.NoError(t, err)
// 	category, err := apiService.CreateCategory(models.CreateCategory{Name: "Sales"}, &user.ID)
// 	require.NoError(t, err)
//
// 	vat, err := apiService.CreateVat(models.CreateVat{Value: 1000}, user.ID)
// 	require.NoError(t, err)
// 	_, err = apiService.CreateVatSetting(models.CreateVatSetting{
// 		Enabled:         true,
// 		BillingDate:     "2024-01-01",
// 		TransactionDate: "2024-01-15",
// 		Interval:        "monthly",
// 	}, user.ID)
// 	require.NoError(t, err)
//
// 	monthly := utils.CycleMonthly
// 	createTransaction(t, apiService, user.ID, category.ID, *currencyCHF.ID, nil, func(p *models.CreateTransaction) {
// 		p.Name = "Past VAT Tx"
// 		p.Amount = 100000
// 		p.StartDate = "2025-01-01"
// 		p.Type = "repeating"
// 		p.Cycle = &monthly
// 		p.Vat = &vat.ID
// 	})
//
// 	results, err := apiService.CalculateForecast(user.ID)
// 	require.NoError(t, err)
// 	require.Empty(t, results, "Past billing/transaction dates should not create future VAT settlements")
// }
