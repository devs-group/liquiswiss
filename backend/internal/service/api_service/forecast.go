package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"sort"
	"time"
)

// addMonthsStable adds months to a date while keeping the day within the target month's valid range.
// For example: Jan 31 + 1 month = Feb 28/29 (last day of Feb), not March 3.
func addMonthsStable(date time.Time, months int) time.Time {
	targetYear := date.Year()
	targetMonth := date.Month() + time.Month(months)

	// Normalize year and month
	for targetMonth > 12 {
		targetYear++
		targetMonth -= 12
	}
	for targetMonth < 1 {
		targetYear--
		targetMonth += 12
	}

	// Get the last day of the target month
	firstOfNextMonth := time.Date(targetYear, targetMonth+1, 1, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location())
	lastDayOfTargetMonth := firstOfNextMonth.AddDate(0, 0, -1).Day()

	// Use the original day or the last day of the target month, whichever is smaller
	targetDay := min(date.Day(), lastDayOfTargetMonth)

	return time.Date(targetYear, targetMonth, targetDay, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location())
}

func (a *APIService) ListForecasts(userID int64, limit int64) ([]models.Forecast, error) {
	forecasts, err := a.dbService.ListForecasts(userID, limit)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(forecasts, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return forecasts, nil
}

func (a *APIService) ListForecastDetails(userID int64, limit int64) ([]models.ForecastDatabaseDetails, error) {
	forecastDetails, err := a.dbService.ListForecastDetails(userID, limit)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(forecastDetails, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return forecastDetails, nil
}

func (a *APIService) ListForecastExclusions(userID int64, relatedID int64, relatedTable string) (map[string]bool, error) {
	forecastExclusions, err := a.dbService.ListForecastExclusions(userID, relatedID, relatedTable)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(forecastExclusions, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return forecastExclusions, nil
}

func (a *APIService) CreateForecastExclusion(payload models.CreateForecastExclusion, userID int64) (int64, error) {
	excludeID, err := a.dbService.CreateForecastExclusion(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return 0, err
	}
	if excludeID == 0 {
		logger.Logger.Error(err)
		return 0, err
	}
	return excludeID, nil
}

func (a *APIService) DeleteForecastExclusion(payload models.CreateForecastExclusion, userID int64) (int64, error) {
	affected, err := a.dbService.DeleteForecastExclusion(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return 0, err
	}
	if affected == 0 {
		logger.Logger.Error(err)
		return 0, err
	}
	return affected, nil
}

func (a *APIService) UpdateForecastExclusions(payload models.UpdateForecastExclusions, userID int64) error {
	for _, update := range payload.Updates {
		request := models.CreateForecastExclusion{
			Month:        update.Month,
			RelatedID:    update.RelatedID,
			RelatedTable: update.RelatedTable,
		}

		var err error
		if update.IsExcluded {
			_, err = a.dbService.CreateForecastExclusion(request, userID)
		} else {
			_, err = a.dbService.DeleteForecastExclusion(request, userID)
		}

		if err != nil {
			logger.Logger.Error(err)
			return err
		}
	}

	return nil
}

func (a *APIService) CalculateForecast(userID int64) ([]models.Forecast, error) {
	page := int64(1)
	limit := int64(100000)
	sortBy := "name"
	sortOrder := "ASC"

	organisation, err := a.GetCurrentOrganisation(userID)
	if err != nil {
		return nil, err
	}

	// Set the organisation wide default currency as base
	baseCurrency := *organisation.Currency.Code

	transactions, _, err := a.ListTransactions(userID, page, limit, sortBy, sortOrder, "")
	if err != nil {
		return nil, err
	}

	fiatRates, err := a.ListFiatRates(baseCurrency)
	if err != nil {
		return nil, err
	}

	today := utils.GetTodayAsUTC()
	maxEndDate := today.AddDate(utils.MaxForecastYears, 0, 0)
	// We include the whole final month, otherwise the results might be confusing
	lastDayOfMaxEndDate := time.Date(maxEndDate.Year(), maxEndDate.Month()+1, 0, 23, 59, 59, 999999999, maxEndDate.Location())

	// We need a map for revenues and expenses
	forecastMap := make(map[string]map[string]int64)
	forecastDetailMap := make(map[string]*models.ForecastDetails)
	for _, transaction := range transactions {
		if transaction.IsDisabled {
			continue
		}
		fiatRate := models.GetFiatRateFromCurrency(fiatRates, baseCurrency, *transaction.Currency.Code)
		amount := models.CalculateAmountWithFiatRate(transaction.Amount, fiatRate)
		if transaction.Vat != nil && !transaction.VatIncluded {
			amount = models.CalculateAmountWithFiatRate(transaction.Amount+transaction.VatAmount, fiatRate)
		}
		isRevenue := amount > 0

		exclusions, err := a.ListForecastExclusions(userID, transaction.ID, utils.TransactionsTableName)
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
	employees, _, err := a.ListEmployees(userID, page, limit, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}
	for _, employee := range employees {
		salaries, _, err := a.ListSalaries(userID, employee.ID, page, limit)
		if err != nil {
			return nil, err
		}
		for _, salary := range salaries {
			if salary.IsDisabled {
				continue
			}
			fromDate := time.Time(salary.FromDate)
			toDate := lastDayOfMaxEndDate
			if salary.ToDate != nil {
				toDate = time.Time(*salary.ToDate)
			}

			fiatRate := models.GetFiatRateFromCurrency(fiatRates, baseCurrency, *salary.Currency.Code)
			// Must be minus here
			netAmount := salary.Amount - salary.EmployeeDeductions
			amount := -models.CalculateAmountWithFiatRate(int64(netAmount), fiatRate)

			salaryExclusions, err := a.ListForecastExclusions(userID, salary.ID, utils.SalariesTableName)
			if err != nil {
				return nil, err
			}

			switch salary.Cycle {
			case utils.CycleMonthly:
				for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 1) {
					if current.Before(today) {
						continue
					}
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if salaryExclusions[monthKey] {
						forecastMap[monthKey]["expense"] += 0
						addForecastDetail(
							forecastDetailMap, monthKey, 0, false, true,
							salary.ID, utils.SalariesTableName, "Löhne", employee.Name,
						)
					} else {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, amount, false, false,
							salary.ID, utils.SalariesTableName, "Löhne", employee.Name,
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
					if salaryExclusions[monthKey] {
						forecastMap[monthKey]["expense"] += 0
						addForecastDetail(
							forecastDetailMap, monthKey, 0, false, true,
							salary.ID, utils.SalariesTableName, "Löhne", employee.Name,
						)
					} else {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, amount, false, false,
							salary.ID, utils.SalariesTableName, "Löhne", employee.Name,
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
					if salaryExclusions[monthKey] {
						forecastMap[monthKey]["expense"] += 0
						addForecastDetail(
							forecastDetailMap, monthKey, 0, false, true,
							salary.ID, utils.SalariesTableName, "Löhne", employee.Name,
						)
					} else {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, amount, false, false,
							salary.ID, utils.SalariesTableName, "Löhne", employee.Name,
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
					if salaryExclusions[monthKey] {
						forecastMap[monthKey]["expense"] += 0
						addForecastDetail(
							forecastDetailMap, monthKey, 0, false, true,
							salary.ID, utils.SalariesTableName, "Löhne", employee.Name,
						)
					} else {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, amount, false, false,
							salary.ID, utils.SalariesTableName, "Löhne", employee.Name,
						)
					}
				}
			}

			// Always calculate the separate costs; salaries without definitions return an empty list.
			salaryCosts, _, err := a.ListSalaryCosts(userID, salary.ID, 1, 1000, false)
			if err != nil {
				return nil, err
			}

			for _, salaryCost := range salaryCosts {
				salaryCostExclusions, err := a.ListForecastExclusions(userID, salaryCost.ID, utils.SalaryCostsTableName)
				if err != nil {
					return nil, err
				}

				if salaryCost.CalculatedNextExecutionDate != nil {
					costFromDate := time.Time(*salaryCost.CalculatedNextExecutionDate)
					distributionMultiplier := int64(models.SalaryCostDistributionMultiplier(salaryCost.DistributionType))
					nextCost := -models.CalculateAmountWithFiatRate(int64(salaryCost.CalculatedNextCost)*distributionMultiplier, fiatRate)

					labelName := "<Kein Label>"
					if salaryCost.Label != nil {
						labelName = salaryCost.Label.Name
					}

					switch salaryCost.Cycle {
					case utils.CycleOnce:
						if costFromDate.Before(today) {
							continue
						}
						monthKey := getYearMonth(costFromDate)
						if forecastMap[monthKey] == nil {
							initForecastMapKey(forecastMap, monthKey)
						}
						if salaryCostExclusions[monthKey] {
							forecastMap[monthKey]["expense"] += 0
							addForecastDetail(
								forecastDetailMap, monthKey, 0, false, true,
								salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
							)
						} else {
							forecastMap[monthKey]["expense"] += nextCost
							addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
								salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
							)
						}
					case utils.CycleMonthly:
						for current := costFromDate; ; current = addOffset(salaryCost.Cycle, costFromDate, current, salaryCost.RelativeOffset) {
							var matchingDetail *models.SalaryCostDetail
							for _, detail := range salaryCost.CalculatedCostDetails {
								if detail.Month == current.Format("2006-01") {
									matchingDetail = &detail
									break
								}
							}
							if matchingDetail == nil {
								break
							}
							distributionMultiplier := int64(models.SalaryCostDistributionMultiplier(salaryCost.DistributionType))
							nextCost := -models.CalculateAmountWithFiatRate(int64(matchingDetail.Amount)*distributionMultiplier, fiatRate)
							monthKey := getYearMonth(current)
							if forecastMap[monthKey] == nil {
								initForecastMapKey(forecastMap, monthKey)
							}
							if salaryCostExclusions[monthKey] {
								forecastMap[monthKey]["expense"] += 0
								addForecastDetail(
									forecastDetailMap, monthKey, 0, false, true,
									salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
								)
							} else {
								forecastMap[monthKey]["expense"] += nextCost
								addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
									salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
								)
							}
						}
					case utils.CycleQuarterly:
						lastToDate := utils.GetNextDate(costFromDate, toDate, 3*int(salaryCost.RelativeOffset))
						for current := costFromDate; ; current = addOffset(salaryCost.Cycle, costFromDate, current, salaryCost.RelativeOffset) {
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
							if salaryCostExclusions[monthKey] {
								forecastMap[monthKey]["expense"] += 0
								addForecastDetail(
									forecastDetailMap, monthKey, 0, false, true,
									salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
								)
							} else {
								forecastMap[monthKey]["expense"] += nextCost
								addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
									salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
								)
							}
						}
					case utils.CycleBiannually:
						lastToDate := utils.GetNextDate(costFromDate, toDate, 6*int(salaryCost.RelativeOffset))
						for current := costFromDate; ; current = addOffset(salaryCost.Cycle, costFromDate, current, salaryCost.RelativeOffset) {
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
							if salaryCostExclusions[monthKey] {
								forecastMap[monthKey]["expense"] += 0
								addForecastDetail(
									forecastDetailMap, monthKey, 0, false, true,
									salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
								)
							} else {
								forecastMap[monthKey]["expense"] += nextCost
								addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
									salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
								)
							}
						}
					case utils.CycleYearly:
						lastToDate := utils.GetNextDate(costFromDate, toDate, 12*int(salaryCost.RelativeOffset))
						for current := costFromDate; ; current = addOffset(salaryCost.Cycle, costFromDate, current, salaryCost.RelativeOffset) {
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
							if salaryCostExclusions[monthKey] {
								forecastMap[monthKey]["expense"] += 0
								addForecastDetail(
									forecastDetailMap, monthKey, 0, false, true,
									salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
								)
							} else {
								forecastMap[monthKey]["expense"] += nextCost
								addForecastDetail(forecastDetailMap, monthKey, nextCost, false, false,
									salaryCost.ID, utils.SalaryCostsTableName, "Lohnkosten", labelName,
								)
							}
						}
					}
				}
			}
		}
	}

	// VAT Settlement Calculation
	vatSetting, err := a.GetVatSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		// Don't fail the forecast if VAT settings can't be retrieved
		vatSetting = nil
	}

	if vatSetting != nil && vatSetting.Enabled {
		// Collect VAT amounts from positive transactions per month
		vatCollectionMap := make(map[string]int64) // month -> total VAT amount

		for _, transaction := range transactions {
			if transaction.IsDisabled {
				continue
			}

			// Only collect VAT from positive (revenue) transactions
			fiatRate := models.GetFiatRateFromCurrency(fiatRates, baseCurrency, *transaction.Currency.Code)
			amount := models.CalculateAmountWithFiatRate(transaction.Amount, fiatRate)

			if amount <= 0 || transaction.Vat == nil || transaction.VatAmount == 0 {
				continue
			}

			vatAmount := models.CalculateAmountWithFiatRate(transaction.VatAmount, fiatRate)

			if transaction.Type == "single" {
				startDate := time.Time(transaction.StartDate)
				// For VAT collection, we INCLUDE past transactions
				// because we need to collect historical VAT for future settlement
				monthKey := getYearMonth(startDate)

				vatCollectionMap[monthKey] += vatAmount
			} else {
				startDate := time.Time(transaction.StartDate)
				endDate := lastDayOfMaxEndDate
				if transaction.EndDate != nil {
					endDate = time.Time(*transaction.EndDate)
				}

				switch *transaction.Cycle {
				case utils.CycleMonthly:
					for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 1) {
						// For VAT collection, we INCLUDE past transactions
						monthKey := getYearMonth(current)
						vatCollectionMap[monthKey] += vatAmount
					}
				case utils.CycleQuarterly:
					for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 3) {
						// For VAT collection, we INCLUDE past transactions
						monthKey := getYearMonth(current)
						vatCollectionMap[monthKey] += vatAmount
					}
				case utils.CycleBiannually:
					for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 6) {
						// For VAT collection, we INCLUDE past transactions
						monthKey := getYearMonth(current)
						vatCollectionMap[monthKey] += vatAmount
					}
				case utils.CycleYearly:
					for current := startDate; !current.After(endDate); current = utils.GetNextDate(startDate, current, 12) {
						// For VAT collection, we INCLUDE past transactions
						monthKey := getYearMonth(current)
						vatCollectionMap[monthKey] += vatAmount
					}
				}
			}
		}

		// Calculate VAT settlement periods based on interval
		var intervalMonths int
		switch vatSetting.Interval {
		case "monthly":
			intervalMonths = 1
		case "quarterly":
			intervalMonths = 3
		case "biannually":
			intervalMonths = 6
		case "yearly":
			intervalMonths = 12
		default:
			intervalMonths = 3 // default to quarterly
		}

		// Group VAT amounts by settlement period and add to forecast
		settlementPeriods := make(map[string]int64) // settlement month -> total VAT

		billingDate := vatSetting.BillingDate
		transactionMonthOffset := vatSetting.TransactionMonthOffset

		// Period start is the first day of the billing month minus the interval (billing marks end of period)
		firstPeriodStart := time.Date(billingDate.Year(), billingDate.Month(), 1, 0, 0, 0, 0, billingDate.Location()).AddDate(0, -intervalMonths, 0)

		for monthKey, vatAmount := range vatCollectionMap {
			monthTime, err := time.Parse("2006-01", monthKey)
			if err != nil {
				continue
			}

			monthsSinceFirstPeriodStart := (monthTime.Year()-firstPeriodStart.Year())*12 + int(monthTime.Month()-firstPeriodStart.Month())
			if monthsSinceFirstPeriodStart < 0 {
				continue
			}

			periodIndex := monthsSinceFirstPeriodStart / intervalMonths
			// Calculate the billing date for this period, then add the month offset for the transaction
			periodBillingDate := addMonthsStable(billingDate, periodIndex*intervalMonths)
			settlementTransactionDate := addMonthsStable(periodBillingDate, transactionMonthOffset)

			if settlementTransactionDate.After(today) {
				settlementKey := getYearMonth(settlementTransactionDate)
				settlementPeriods[settlementKey] += vatAmount
			}
		}

		// Add VAT settlements as expenses
		for settlementKey, totalVat := range settlementPeriods {
			if forecastMap[settlementKey] == nil {
				initForecastMapKey(forecastMap, settlementKey)
			}

			// Add as negative expense
			forecastMap[settlementKey]["expense"] += -totalVat

			// Add to forecast details
			addForecastDetail(
				forecastDetailMap, settlementKey, -totalVat, false, false,
				0, "vat_settlement", "Mwst.", "Mwst.",
			)
		}
	}

	_, err = a.dbService.ClearForecasts(userID)
	if err != nil {
		return nil, err
	}

	for monthKey, forecast := range forecastMap {
		revenue := forecast["revenue"]
		expense := forecast["expense"]
		forecastID, err := a.dbService.UpsertForecast(models.CreateForecast{
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

		_, err = a.dbService.UpsertForecastDetail(models.CreateForecastDetail{
			Month:      monthKey,
			Revenue:    revenueList,
			Expense:    expenseList,
			ForecastID: forecastID,
		}, userID, forecastID)
		if err != nil {
			return nil, err
		}
	}

	forecasts, err := a.ListForecasts(userID, int64(utils.GetTotalMonthsForMaxForecastYears()))
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
			Revenue: make(map[string]any),
			Expense: make(map[string]any),
		}
	}

	var currentMap map[string]any

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
				currentMap[category] = make(map[string]any)
			}
			currentMap = currentMap[category].(map[string]any)
		}
	}
}

func iterateForecastDetails(data map[string]any, result *[]models.ForecastDetailRevenueExpense) {
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
		case map[string]any:
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
