//go:generate mockgen -package=mocks -destination ../mocks/forecast_service.go liquiswiss/internal/service/forecast_service IForecastService
package forecast_service

import (
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"strconv"
	"time"
)

type EmployeeForecastData struct {
	Data map[models.Employee][]models.EmployeeHistory
}

type IForecastService interface {
	CalculateForecast(userID int64) ([]models.Forecast, error)
}

type ForecastService struct {
	dbService db_service.IDatabaseService
}

func NewForecastService(dbService *db_service.IDatabaseService) IForecastService {
	return &ForecastService{
		dbService: *dbService,
	}
}

func (f *ForecastService) CalculateForecast(userID int64) ([]models.Forecast, error) {
	page := int64(1)
	limit := int64(100000)
	sortBy := "name"
	sortOrder := "ASC"

	transactions, _, err := f.dbService.ListTransactions(userID, page, limit, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	fiatRates, err := f.dbService.ListFiatRates(utils.BaseCurrency)
	if err != nil {
		return nil, err
	}

	today := utils.GetTodayAsUTC()
	maxEndDate := today.AddDate(3, 0, 0)
	// We include the whole final month, otherwise the results might be confusing
	lastDayOfMaxEndDate := time.Date(maxEndDate.Year(), maxEndDate.Month()+1, 0, 23, 59, 59, 999999999, maxEndDate.Location())

	// We need a map for revenues and expenses
	forecastMap := make(map[string]map[string]int64)
	forecastDetailMap := make(map[string]*models.ForecastDetails)
	for _, transaction := range transactions {
		fiatRate := models.GetFiatRateFromCurrency(fiatRates, *transaction.Currency.Code)
		amount := models.CalculateAmountWithFiatRate(transaction.Amount, fiatRate)
		if transaction.Vat != nil && !transaction.VatIncluded {
			amount = models.CalculateAmountWithFiatRate(transaction.Amount+transaction.VatAmount, fiatRate)
		}

		if transaction.Type == "single" {
			startDate := time.Time(transaction.StartDate)
			if startDate.Before(today) {
				continue
			}
			monthKey := getYearMonth(startDate)
			if forecastMap[monthKey] == nil {
				initForecastMapKey(forecastMap, monthKey)
			}
			if amount > 0 {
				forecastMap[monthKey]["revenue"] += amount
				addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
			} else if amount < 0 {
				forecastMap[monthKey]["expense"] += amount
				addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
			}
		} else {
			startDate := time.Time(transaction.StartDate)
			endDate := lastDayOfMaxEndDate
			if transaction.EndDate != nil {
				endDate = time.Time(*transaction.EndDate)
			}
			switch *transaction.Cycle {
			case "daily":
				for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					}
				}
			case "weekly":
				for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 7) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					}
				}
			case "monthly":
				for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 1) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					}
				}
			case "quarterly":
				for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 3) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					}
				}
			case "biannually":
				for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 6) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					}
				}
			case "yearly":
				for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 12) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, amount)
					}
				}
			}
		}
	}

	// Collect the employee expenses now
	employees, _, err := f.dbService.ListEmployees(userID, page, limit, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}
	for _, employee := range employees {
		employeeHistories, _, err := f.dbService.ListEmployeeHistory(userID, strconv.FormatInt(employee.ID, 10), page, limit)
		if err != nil {
			return nil, err
		}
		for _, history := range employeeHistories {
			fromDate := time.Time(history.FromDate)
			toDate := lastDayOfMaxEndDate
			if history.ToDate != nil {
				toDate = time.Time(*history.ToDate)
			}

			fiatRate := models.GetFiatRateFromCurrency(fiatRates, *history.Currency.Code)
			// Must be minus here
			salary := -models.CalculateAmountWithFiatRate(int64(history.Salary), fiatRate)

			switch history.Cycle {
			case "daily":
				for current := fromDate; !current.After(toDate); current = current.AddDate(0, 0, 1) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					forecastMap[monthKey]["expense"] += salary
					addForecastDetail(forecastDetailMap, monthKey, "Gehälter", salary)
				}
			case "weekly":
				for current := fromDate; !current.After(toDate); current = current.AddDate(0, 0, 7) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					forecastMap[monthKey]["expense"] += salary
					addForecastDetail(forecastDetailMap, monthKey, "Gehälter", salary)
				}
			case "monthly":
				for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 1) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					forecastMap[monthKey]["expense"] += salary
					addForecastDetail(forecastDetailMap, monthKey, "Gehälter", salary)
				}
			case "quarterly":
				for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 3) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					forecastMap[monthKey]["expense"] += salary
					addForecastDetail(forecastDetailMap, monthKey, "Gehälter", salary)
				}
			case "biannually":
				for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 6) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					forecastMap[monthKey]["expense"] += salary
					addForecastDetail(forecastDetailMap, monthKey, "Gehälter", salary)
				}
			case "yearly":
				for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 12) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					forecastMap[monthKey]["expense"] += salary
					addForecastDetail(forecastDetailMap, monthKey, "Gehälter", salary)
				}
			}
		}
	}

	_, err = f.dbService.ClearForecasts(userID)
	if err != nil {
		return nil, err
	}

	for monthKey, forecast := range forecastMap {
		revenue := forecast["revenue"]
		expense := forecast["expense"]
		forecastID, err := f.dbService.UpsertForecast(models.CreateForecast{
			Month:    monthKey,
			Revenue:  revenue,
			Expense:  expense,
			Cashflow: revenue + expense,
		}, userID)
		if err != nil {
			return nil, err
		}

		// Upsert the details along with the forecast
		forecastDetail := forecastDetailMap[monthKey]
		revenueList := make([]models.ForecastDetailRevenueExpense, 0)
		expenseList := make([]models.ForecastDetailRevenueExpense, 0)
		for name, amount := range forecastDetail.Revenue {
			revenueList = append(revenueList, models.ForecastDetailRevenueExpense{
				Name:   name,
				Amount: amount,
			})
		}
		for name, amount := range forecastDetail.Expense {
			expenseList = append(expenseList, models.ForecastDetailRevenueExpense{
				Name:   name,
				Amount: amount,
			})
		}
		_, err = f.dbService.UpsertForecastDetail(models.CreateForecastDetail{
			Month:      monthKey,
			Revenue:    revenueList,
			Expense:    expenseList,
			ForecastID: forecastID,
		}, userID, forecastID)
		if err != nil {
			return nil, err
		}
	}

	forecasts, err := f.dbService.ListForecasts(userID, 37)
	if err != nil {
		return nil, err
	}

	validator := utils.GetValidator()
	if err := validator.Var(forecasts, "dive"); err != nil {
		// Return validation errors
		return nil, err
	}

	return forecasts, nil
}

func initForecastMapKey(forecastMap map[string]map[string]int64, monthKey string) {
	forecastMap[monthKey] = make(map[string]int64)
	forecastMap[monthKey]["revenue"] = 0
	forecastMap[monthKey]["expense"] = 0
}

func addForecastDetail(detailMap map[string]*models.ForecastDetails, monthKey, category string, amount int64) {
	// Make sure the map is prepared
	if detailMap[monthKey] == nil {
		detailMap[monthKey] = &models.ForecastDetails{
			Revenue: make(map[string]int64),
			Expense: make(map[string]int64),
		}
	}
	if amount > 0 {
		// Make sure the inner map is prepared
		if detailMap[monthKey].Revenue[category] == 0 {
			detailMap[monthKey].Revenue[category] = 0
		}

		detailMap[monthKey].Revenue[category] += amount
	} else if amount < 0 {
		// Make sure the inner map is prepared
		if detailMap[monthKey].Expense[category] == 0 {
			detailMap[monthKey].Expense[category] = 0
		}

		detailMap[monthKey].Expense[category] += amount
	}
}

func getYearMonth(date time.Time) string {
	return date.Format("2006-01")
}

func getTodayAndLastDayOfMaxEndDate() (time.Time, time.Time) {
	today := utils.GetTodayAsUTC()
	maxEndDate := today.AddDate(3, 0, 0)
	// We include the whole final month, otherwise the results might be confusing
	lastDayOfMaxEndDate := time.Date(maxEndDate.Year(), maxEndDate.Month()+1, 0, 23, 59, 59, 999999999, maxEndDate.Location())
	return today, lastDayOfMaxEndDate
}