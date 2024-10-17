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

func CalculateForecasts(dbService service.IDatabaseService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	page := int64(1)
	limit := int64(100000)

	transactions, _, err := dbService.ListTransactions(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fiatRates, err := dbService.ListFiatRates(utils.BaseCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// We need a map for revenues and expenses
	forecastMap := make(map[string]map[string]int64)
	for _, transaction := range transactions {
		fiatRate := utils.GetFiatRateFromCurrency(fiatRates, *transaction.Currency.Code)
		amount := utils.CalculateAmountWithFiatRate(transaction.Amount, fiatRate)

		if transaction.Type == "single" {
			startDate := time.Time(transaction.StartDate)
			monthKey := getYearMonth(startDate)
			if forecastMap[monthKey] == nil {
				initForecastMapKey(forecastMap, monthKey)
			}
			if amount > 0 {
				forecastMap[monthKey]["revenue"] += amount
			} else if amount < 0 {
				forecastMap[monthKey]["expense"] += amount
			}
		} else {
			startDate := time.Time(transaction.StartDate)
			endDate := startDate.AddDate(3, 0, 0)
			if transaction.EndDate != nil {
				endDate = time.Time(*transaction.EndDate)
			}
			switch *transaction.Cycle {
			case "daily":
				for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
					}
				}
			case "weekly":
				for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 7) {
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
					}
				}
			case "monthly":
				for current := startDate; !current.After(endDate); current = getNextDate(startDate, current, 1) {
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
					}
				}
			case "quarterly":
				for current := startDate; !current.After(endDate); current = getNextDate(startDate, current, 3) {
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
					}
				}
			case "biannually":
				for current := startDate; !current.After(endDate); current = getNextDate(startDate, current, 6) {
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
					}
				}
			case "yearly":
				for current := startDate; !current.After(endDate); current = getNextDate(startDate, current, 12) {
					monthKey := getYearMonth(current)
					if forecastMap[monthKey] == nil {
						initForecastMapKey(forecastMap, monthKey)
					}
					if amount > 0 {
						forecastMap[monthKey]["revenue"] += amount
					} else if amount < 0 {
						forecastMap[monthKey]["expense"] += amount
					}
				}
			}
		}
	}

	// Collect the employee expenses now
	employees, _, err := dbService.ListEmployees(userID, page, limit)
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
			toDate := time.Now().AddDate(3, 0, 0)
			if history.ToDate != nil {
				toDate = time.Time(*history.ToDate)
			}

			for current := fromDate; !current.After(toDate); current = getNextDate(fromDate, current, 1) {
				monthKey := getYearMonth(current)
				if forecastMap[monthKey] == nil {
					initForecastMapKey(forecastMap, monthKey)
				}
				// Must be minus here
				forecastMap[monthKey]["expense"] -= int64(history.SalaryPerMonth)
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
		_, err = dbService.UpsertForecast(models.CreateForecast{
			Month:    monthKey,
			Revenue:  revenue,
			Expense:  expense,
			Cashflow: revenue + expense,
		}, userID)
	}

	forecasts, err := dbService.ListForecasts(userID, 12)
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

func initForecastMapKey(forecastMap map[string]map[string]int64, monthKey string) {
	forecastMap[monthKey] = make(map[string]int64)
	forecastMap[monthKey]["revenue"] = 0
	forecastMap[monthKey]["expense"] = 0
}

func getNextDate(referenceDate, currentDate time.Time, months int) time.Time {
	dayDiff := referenceDate.Day() - currentDate.Day()

	nextDate := currentDate.AddDate(0, months, dayDiff)
	if currentDate.Day() > nextDate.Day() {
		nextDate = time.Date(nextDate.Year(), nextDate.Month(), 0, referenceDate.Hour(), referenceDate.Minute(), referenceDate.Second(), referenceDate.Nanosecond(), referenceDate.Location())
	}

	return nextDate
}

func getYearMonth(date time.Time) string {
	return date.Format("2006-01")
}
