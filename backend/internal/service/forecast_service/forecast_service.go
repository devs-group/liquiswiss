//go:generate mockgen -package=mocks -destination ../mocks/forecast_service.go liquiswiss/internal/service/forecast_service IForecastService
package forecast_service

import (
	"liquiswiss/internal/service/db_service"
	"liquiswiss/internal/service/user_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"sort"
	"time"
)

type EmployeeForecastData struct {
	Data map[models.Employee][]models.EmployeeHistory
}

type IForecastService interface {
	CalculateForecast(userID int64) ([]models.Forecast, error)
}

type ForecastService struct {
	dbService   db_service.IDatabaseService
	userService user_service.IUserService
}

func NewForecastService(dbService *db_service.IDatabaseService, userService *user_service.IUserService) IForecastService {
	return &ForecastService{
		dbService:   *dbService,
		userService: *userService,
	}
}

func (f *ForecastService) CalculateForecast(userID int64) ([]models.Forecast, error) {
	page := int64(1)
	limit := int64(100000)
	sortBy := "name"
	sortOrder := "ASC"

	organisation, err := f.userService.GetCurrentOrganisation(userID)
	if err != nil {
		return nil, err
	}

	// Set the organisation wide default currency as base
	baseCurrency := *organisation.Currency.Code

	transactions, _, err := f.dbService.ListTransactions(userID, page, limit, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	fiatRates, err := f.dbService.ListFiatRates(baseCurrency)
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
		fiatRate := models.GetFiatRateFromCurrency(fiatRates, baseCurrency, *transaction.Currency.Code)
		amount := models.CalculateAmountWithFiatRate(transaction.Amount, fiatRate)
		if transaction.Vat != nil && !transaction.VatIncluded {
			amount = models.CalculateAmountWithFiatRate(transaction.Amount+transaction.VatAmount, fiatRate)
		}
		isRevenue := amount > 0

		exclusions, err := f.dbService.ListForecastExclusions(userID, transaction.ID, utils.TransactionsTableName)
		if err != nil {
			return nil, err
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
				if exclusions[monthKey] {
					forecastMap[monthKey]["revenue"] += 0
					addForecastDetail(
						forecastDetailMap, monthKey, 0, isRevenue, true,
						transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
					)
				} else {
					forecastMap[monthKey]["revenue"] += amount
					addForecastDetail(
						forecastDetailMap, monthKey, amount, isRevenue, false,
						transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
					)
				}
			} else if amount < 0 {
				if exclusions[monthKey] {
					forecastMap[monthKey]["expense"] += 0
					addForecastDetail(
						forecastDetailMap, monthKey, 0, isRevenue, true,
						transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
					)
				} else {
					forecastMap[monthKey]["expense"] += amount
					addForecastDetail(forecastDetailMap, monthKey, amount, isRevenue, false,
						transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
					)
				}
			}
		} else {
			startDate := time.Time(transaction.StartDate)
			endDate := lastDayOfMaxEndDate
			if transaction.EndDate != nil {
				endDate = time.Time(*transaction.EndDate)
			}
			switch *transaction.Cycle {
			case utils.CycleMonthly:
				for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 1) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						if exclusions[monthKey] {
							forecastMap[monthKey]["revenue"] += 0
							addForecastDetail(
								forecastDetailMap, monthKey, 0, isRevenue, true,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						} else {
							forecastMap[monthKey]["revenue"] += amount
							addForecastDetail(forecastDetailMap, monthKey, amount, isRevenue, false,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						}
					} else if amount < 0 {
						if exclusions[monthKey] {
							forecastMap[monthKey]["expense"] += 0
							addForecastDetail(forecastDetailMap, monthKey, 0, isRevenue, true,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						} else {
							forecastMap[monthKey]["expense"] += amount
							addForecastDetail(forecastDetailMap, monthKey, amount, isRevenue, false,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						}
					}
				}
			case utils.CycleQuarterly:
				for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 3) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						if exclusions[monthKey] {
							forecastMap[monthKey]["revenue"] += 0
							addForecastDetail(
								forecastDetailMap, monthKey, 0, isRevenue, true,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						} else {
							forecastMap[monthKey]["revenue"] += amount
							addForecastDetail(
								forecastDetailMap, monthKey, amount, isRevenue, false,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						}
					} else if amount < 0 {
						if exclusions[monthKey] {
							forecastMap[monthKey]["expense"] += 0
							addForecastDetail(
								forecastDetailMap, monthKey, 0, isRevenue, true,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						} else {
							forecastMap[monthKey]["expense"] += amount
							addForecastDetail(
								forecastDetailMap, monthKey, amount, isRevenue, false,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						}
					}
				}
			case utils.CycleBiannually:
				for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 6) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						if exclusions[monthKey] {
							forecastMap[monthKey]["revenue"] += 0
							addForecastDetail(
								forecastDetailMap, monthKey, 0, isRevenue, true,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						} else {
							forecastMap[monthKey]["revenue"] += amount
							addForecastDetail(
								forecastDetailMap, monthKey, amount, isRevenue, false,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						}
					} else if amount < 0 {
						if exclusions[monthKey] {
							forecastMap[monthKey]["expense"] += 0
							addForecastDetail(
								forecastDetailMap, monthKey, 0, isRevenue, true,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						} else {
							forecastMap[monthKey]["expense"] += amount
							addForecastDetail(
								forecastDetailMap, monthKey, amount, isRevenue, false,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						}
					}
				}
			case utils.CycleYearly:
				for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 12) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						if exclusions[monthKey] {
							forecastMap[monthKey]["revenue"] += 0
							addForecastDetail(
								forecastDetailMap, monthKey, 0, isRevenue, true,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						} else {
							forecastMap[monthKey]["revenue"] += amount
							addForecastDetail(
								forecastDetailMap, monthKey, amount, isRevenue, false,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						}
					} else if amount < 0 {
						if exclusions[monthKey] {
							forecastMap[monthKey]["expense"] += 0
							addForecastDetail(
								forecastDetailMap, monthKey, 0, isRevenue, true,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						} else {
							forecastMap[monthKey]["expense"] += amount
							addForecastDetail(
								forecastDetailMap, monthKey, amount, isRevenue, false,
								transaction.ID, utils.TransactionsTableName, transaction.Category.Name, transaction.Name,
							)
						}
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
		employeeHistories, _, err := f.dbService.ListEmployeeHistory(userID, employee.ID, page, limit)
		if err != nil {
			return nil, err
		}
		for _, history := range employeeHistories {
			fromDate := time.Time(history.FromDate)
			toDate := lastDayOfMaxEndDate
			if history.ToDate != nil {
				toDate = time.Time(*history.ToDate)
			}

			fiatRate := models.GetFiatRateFromCurrency(fiatRates, baseCurrency, *history.Currency.Code)
			// Must be minus here
			netSalary := history.Salary
			if history.WithSeparateCosts {
				netSalary = history.Salary - history.EmployeeDeductions
			}
			salary := -models.CalculateAmountWithFiatRate(int64(netSalary), fiatRate)

			historyExclusions, err := f.dbService.ListForecastExclusions(userID, history.ID, utils.EmployeeHistoriesTableName)
			if err != nil {
				return nil, err
			}

			switch history.Cycle {
			case utils.CycleMonthly:
				for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 1) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if historyExclusions[monthKey] {
						forecastMap[monthKey]["expense"] += 0
						addForecastDetail(
							forecastDetailMap, monthKey, 0, false, true,
							history.ID, utils.EmployeeHistoriesTableName, "Löhne", employee.Name,
						)
					} else {
						forecastMap[monthKey]["expense"] += salary
						addForecastDetail(forecastDetailMap, monthKey, salary, false, false,
							history.ID, utils.EmployeeHistoriesTableName, "Löhne", employee.Name,
						)
					}
				}
			case utils.CycleQuarterly:
				for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 3) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if historyExclusions[monthKey] {
						forecastMap[monthKey]["expense"] += 0
						addForecastDetail(
							forecastDetailMap, monthKey, 0, false, true,
							history.ID, utils.EmployeeHistoriesTableName, "Löhne", employee.Name,
						)
					} else {
						forecastMap[monthKey]["expense"] += salary
						addForecastDetail(forecastDetailMap, monthKey, salary, false, false,
							history.ID, utils.EmployeeHistoriesTableName, "Löhne", employee.Name,
						)
					}
				}
			case utils.CycleBiannually:
				for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 6) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if historyExclusions[monthKey] {
						forecastMap[monthKey]["expense"] += 0
						addForecastDetail(
							forecastDetailMap, monthKey, 0, false, true,
							history.ID, utils.EmployeeHistoriesTableName, "Löhne", employee.Name,
						)
					} else {
						forecastMap[monthKey]["expense"] += salary
						addForecastDetail(forecastDetailMap, monthKey, salary, false, false,
							history.ID, utils.EmployeeHistoriesTableName, "Löhne", employee.Name,
						)
					}
				}
			case utils.CycleYearly:
				for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 12) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if historyExclusions[monthKey] {
						forecastMap[monthKey]["expense"] += 0
						addForecastDetail(
							forecastDetailMap, monthKey, 0, false, true,
							history.ID, utils.EmployeeHistoriesTableName, "Löhne", employee.Name,
						)
					} else {
						forecastMap[monthKey]["expense"] += salary
						addForecastDetail(forecastDetailMap, monthKey, salary, false, false,
							history.ID, utils.EmployeeHistoriesTableName, "Löhne", employee.Name,
						)
					}
				}
			}

			// Calculate the separate costs if wanted
			var historyCosts []models.EmployeeHistoryCost
			if history.WithSeparateCosts {
				historyCosts, _, err = f.dbService.ListEmployeeHistoryCosts(userID, history.ID, page, limit)
				if err != nil {
					return nil, err
				}
			}

			for _, historyCost := range historyCosts {
				historyCostExclusions, err := f.dbService.ListForecastExclusions(userID, historyCost.ID, utils.EmployeeHistoryCostsTableName)
				if err != nil {
					return nil, err
				}

				if historyCost.CalculatedNextExecutionDate != nil {
					costFromDate := time.Time(*historyCost.CalculatedNextExecutionDate)
					nextCost := -models.CalculateAmountWithFiatRate(int64(historyCost.CalculatedNextCost), fiatRate)

					labelName := "<Kein Label>"
					if historyCost.Label != nil {
						labelName = historyCost.Label.Name
					}

					switch historyCost.Cycle {
					case utils.CycleOnce:
						if costFromDate.Before(today) {
							continue
						}
						monthKey := getYearMonth(costFromDate)
						if forecastMap[monthKey] == nil {
							initForecastMapKey(forecastMap, monthKey)
						}
						if historyCostExclusions[monthKey] {
							forecastMap[monthKey]["expense"] += 0
							addForecastDetail(
								forecastDetailMap, monthKey, 0, false, true,
								historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
							)
						} else {
							forecastMap[monthKey]["expense"] += nextCost
							addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
								historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
							)
						}
					case utils.CycleMonthly:
						for current := costFromDate; ; current = addOffset(historyCost.Cycle, costFromDate, current, historyCost.RelativeOffset) {
							var matchingDetail *models.EmployeeHistoryCostDetail
							for _, detail := range historyCost.CalculatedCostDetails {
								if detail.Month == current.Format("2006-01") {
									matchingDetail = &detail
									break
								}
							}
							if matchingDetail == nil {
								break
							}
							nextCost := -models.CalculateAmountWithFiatRate(int64(matchingDetail.Amount), fiatRate)
							monthKey := getYearMonth(current)
							if forecastMap[monthKey] == nil {
								initForecastMapKey(forecastMap, monthKey)
							}
							if historyCostExclusions[monthKey] {
								forecastMap[monthKey]["expense"] += 0
								addForecastDetail(
									forecastDetailMap, monthKey, 0, false, true,
									historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
								)
							} else {
								forecastMap[monthKey]["expense"] += nextCost
								addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
									historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
								)
							}
						}
					case utils.CycleQuarterly:
						lastToDate := utils.GetNextDate(costFromDate, toDate, 3*int(historyCost.RelativeOffset))
						for current := costFromDate; ; current = addOffset(historyCost.Cycle, costFromDate, current, historyCost.RelativeOffset) {
							if current.After(lastToDate) {
								break
							}
							if current.Before(today) {
								continue
							}
							monthKey := getYearMonth(current)
							if forecastMap[monthKey] == nil {
								initForecastMapKey(forecastMap, monthKey)
							}
							if historyCostExclusions[monthKey] {
								forecastMap[monthKey]["expense"] += 0
								addForecastDetail(
									forecastDetailMap, monthKey, 0, false, true,
									historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
								)
							} else {
								forecastMap[monthKey]["expense"] += nextCost
								addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
									historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
								)
							}
						}
					case utils.CycleBiannually:
						lastToDate := utils.GetNextDate(costFromDate, toDate, 6*int(historyCost.RelativeOffset))
						for current := costFromDate; ; current = addOffset(historyCost.Cycle, costFromDate, current, historyCost.RelativeOffset) {
							if current.After(lastToDate) {
								break
							}
							if current.Before(today) {
								continue
							}
							monthKey := getYearMonth(current)
							if forecastMap[monthKey] == nil {
								initForecastMapKey(forecastMap, monthKey)
							}
							if historyCostExclusions[monthKey] {
								forecastMap[monthKey]["expense"] += 0
								addForecastDetail(
									forecastDetailMap, monthKey, 0, false, true,
									historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
								)
							} else {
								forecastMap[monthKey]["expense"] += nextCost
								addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
									historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
								)
							}
						}
					case utils.CycleYearly:
						lastToDate := utils.GetNextDate(costFromDate, toDate, 12*int(historyCost.RelativeOffset))
						for current := costFromDate; ; current = addOffset(historyCost.Cycle, costFromDate, current, historyCost.RelativeOffset) {
							if current.After(lastToDate) {
								break
							}
							if current.Before(today) {
								continue
							}
							monthKey := getYearMonth(current)
							if forecastMap[monthKey] == nil {
								initForecastMapKey(forecastMap, monthKey)
							}
							if historyCostExclusions[monthKey] {
								forecastMap[monthKey]["expense"] += 0
								addForecastDetail(
									forecastDetailMap, monthKey, 0, false, true,
									historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
								)
							} else {
								forecastMap[monthKey]["expense"] += nextCost
								addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
									historyCost.ID, utils.EmployeeHistoryCostsTableName, "Lohnkosten", labelName,
								)
							}
						}
					}
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

		iterateForecastDetails(forecastDetail.Revenue, &revenueList)
		iterateForecastDetails(forecastDetail.Expense, &expenseList)

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

func addForecastDetail(detailMap map[string]*models.ForecastDetails, monthKey string, amount int64, isRevenue, isExcluded bool, relatedID int64, relatedTable string, categories ...string) {
	if detailMap[monthKey] == nil {
		detailMap[monthKey] = &models.ForecastDetails{
			Revenue: make(map[string]interface{}),
			Expense: make(map[string]interface{}),
		}
	}

	var currentMap map[string]interface{}

	if isRevenue {
		currentMap = detailMap[monthKey].Revenue
	} else {
		currentMap = detailMap[monthKey].Expense
	}

	// Traverse through the categories to create or navigate nested maps
	for i, category := range categories {
		if i == len(categories)-1 {
			// If this is the last category, add the amount
			if _, exists := currentMap[category]; !exists {
				currentMap[category] = models.ForecastDetail{
					Amount:       0,
					RelatedID:    relatedID,
					RelatedTable: relatedTable,
					IsExcluded:   isExcluded,
				}
			}
			existingDetail := currentMap[category].(models.ForecastDetail)
			existingDetail.Amount += amount
			existingDetail.IsExcluded = isExcluded
			currentMap[category] = existingDetail
		} else {
			// Otherwise, ensure the nested map exists and navigate deeper
			if _, exists := currentMap[category]; !exists {
				currentMap[category] = make(map[string]interface{})
			}
			currentMap = currentMap[category].(map[string]interface{})
		}
	}
}

func iterateForecastDetails(data map[string]interface{}, result *[]models.ForecastDetailRevenueExpense) {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := data[key]
		switch v := value.(type) {
		case models.ForecastDetail:
			// Leaf node with an amount
			*result = append(*result, models.ForecastDetailRevenueExpense{
				Name:         key,
				Amount:       v.Amount,
				RelatedID:    v.RelatedID,
				RelatedTable: v.RelatedTable,
				IsExcluded:   v.IsExcluded,
			})
		case map[string]interface{}:
			// Nested node with children
			children := []models.ForecastDetailRevenueExpense{}
			iterateForecastDetails(v, &children)
			*result = append(*result, models.ForecastDetailRevenueExpense{
				Name:     key,
				Children: children,
			})
		default:
			// Handle unexpected types gracefully
		}
	}
}

func getYearMonth(date time.Time) string {
	return date.Format("2006-01")
}

func addOffset(costCycle string, fromDate time.Time, current time.Time, relativeOffset int64) time.Time {
	offset := int(relativeOffset)
	switch costCycle {
	case utils.CycleMonthly:
		return utils.GetNextDate(fromDate, current, offset*1)
	case utils.CycleQuarterly:
		return utils.GetNextDate(fromDate, current, offset*3)
	case utils.CycleBiannually:
		return utils.GetNextDate(fromDate, current, offset*6)
	case utils.CycleYearly:
		return utils.GetNextDate(fromDate, current, offset*12)
	}
	return current
}
