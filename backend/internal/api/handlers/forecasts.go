package handlers

import (
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

func ListForecasts(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	forecasts, err := dbService.ListForecasts(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(forecasts, "dive"); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, forecasts)
}

func ListForecastDetails(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	forecastDetails, err := dbService.ListForecastDetails(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(forecastDetails, "dive"); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, forecastDetails)
}

func CalculateForecasts(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	page := int64(1)
	limit := int64(100000)
	sortBy := "name"
	sortOrder := "ASC"

	transactions, _, err := dbService.ListTransactions(userID, page, limit, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fiatRates, err := dbService.ListFiatRates(utils.BaseCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
				addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
			} else if amount < 0 {
				forecastMap[monthKey]["expense"] += amount
				addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
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
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
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
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
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
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
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
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
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
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
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
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
						addForecastDetail(forecastDetailMap, monthKey, transaction.Category.Name, transaction.Amount)
					}
				}
			}
		}
	}

	// Collect the employee expenses now
	employees, _, err := dbService.ListEmployees(userID, page, limit, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, employee := range employees {
		employeeHistories, _, err := dbService.ListEmployeeHistory(userID, strconv.FormatInt(employee.ID, 10), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, history := range employeeHistories {
			fromDate := time.Time(history.FromDate)
			toDate := lastDayOfMaxEndDate
			if history.ToDate != nil {
				toDate = time.Time(*history.ToDate)
			}

			for current := fromDate; !current.After(toDate); current = utils.GetNextDate(fromDate, current, 1) {
				if current.Before(today) {
					continue
				}
				monthKey := getYearMonth(current)
				if forecastMap[monthKey] == nil {
					initForecastMapKey(forecastMap, monthKey)
				}
				// Must be minus here
				salaryPerMonth := -int64(history.SalaryPerMonth)
				forecastMap[monthKey]["expense"] += salaryPerMonth
				addForecastDetail(forecastDetailMap, monthKey, "Gehälter", salaryPerMonth)
			}
		}
	}

	_, err = dbService.ClearForecasts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for monthKey, forecast := range forecastMap {
		revenue := forecast["revenue"]
		expense := forecast["expense"]
		forecastID, err := dbService.UpsertForecast(models.CreateForecast{
			Month:    monthKey,
			Revenue:  revenue,
			Expense:  expense,
			Cashflow: revenue + expense,
		}, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
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
		_, err = dbService.UpsertForecastDetail(models.CreateForecastDetail{
			Month:      monthKey,
			Revenue:    revenueList,
			Expense:    expenseList,
			ForecastID: forecastID,
		}, userID, forecastID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	forecasts, err := dbService.ListForecasts(userID, 37)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validator := utils.GetValidator()
	if err := validator.Var(forecasts, "dive"); err != nil {
		// Return validation errors
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Daten", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"forecast": forecasts,
		"details":  forecastDetailMap,
	})
}

func initForecastMapKey(forecastMap map[string]map[string]int64, monthKey string) {
	forecastMap[monthKey] = make(map[string]int64)
	forecastMap[monthKey]["revenue"] = 0
	forecastMap[monthKey]["expense"] = 0
}

func initForecastDetailMapKey(forecastDetailMap map[string]map[string]map[string]int64, monthKey, subkey string) {
	forecastDetailMap[monthKey] = make(map[string]map[string]int64)
	forecastDetailMap[monthKey]["revenue"] = make(map[string]int64)
	forecastDetailMap[monthKey]["expense"] = make(map[string]int64)
	forecastDetailMap[monthKey]["revenue"][subkey] = 0
	forecastDetailMap[monthKey]["expense"][subkey] = 0
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
