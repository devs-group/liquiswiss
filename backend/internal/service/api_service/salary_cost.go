package api_service

import (
	"fmt"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
	"time"
)

func (a *APIService) ListSalaryCosts(userID int64, salaryID int64, page int64, limit int64, skipPrevious bool) ([]models.SalaryCost, int64, error) {
	salaryCosts, totalCount, err := a.dbService.ListSalaryCosts(userID, salaryID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range salaryCosts {
		salaryCost := salaryCosts[i]
		updatedSalaryCost, err := a.applyCostCalculation(&salaryCost, skipPrevious)
		if err != nil {
			return nil, 0, err
		}
		salaryCosts[i] = *updatedSalaryCost
	}
	validator := utils.GetValidator()
	if err := validator.Var(salaryCosts, "dive"); err != nil {
		return nil, 0, err
	}
	return salaryCosts, totalCount, err
}

func (a *APIService) GetSalaryCost(userID int64, salaryCostID int64, skipPrevious bool) (*models.SalaryCost, error) {
	salaryCost, err := a.dbService.GetSalaryCost(userID, salaryCostID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	salaryCost, err = a.applyCostCalculation(salaryCost, skipPrevious)
	if err != nil {
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(salaryCost); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return salaryCost, nil
}

func (a *APIService) CreateSalaryCost(payload models.CreateSalaryCost, userID int64, salaryID int64) (*models.SalaryCost, error) {
	salary, err := a.dbService.GetSalary(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if salary.IsTermination {
		return nil, fmt.Errorf("cannot attach costs to a termination salary")
	}

	salaryCostID, err := a.dbService.CreateSalaryCost(payload, userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.CalculateSalaryCostDetails(userID, salaryCostID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	salaryCost, err := a.GetSalaryCost(userID, salaryCostID, true)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(salaryCost); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Recalculate Forecast
	_, err = a.CalculateForecast(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return salaryCost, nil
}

func (a *APIService) UpdateSalaryCost(payload models.CreateSalaryCost, userID int64, salaryCostID int64) (*models.SalaryCost, error) {
	_, err := a.GetSalaryCost(userID, salaryCostID, true)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.UpdateSalaryCost(payload, userID, salaryCostID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.CalculateSalaryCostDetails(userID, salaryCostID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	salaryCost, err := a.GetSalaryCost(userID, salaryCostID, true)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(salaryCost); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Recalculate Forecast
	_, err = a.CalculateForecast(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return salaryCost, nil
}

func (a *APIService) DeleteSalaryCost(userID int64, salaryCostID int64) error {
	existingSalaryCost, err := a.GetSalaryCost(userID, salaryCostID, true)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.DeleteSalaryCost(userID, existingSalaryCost.ID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	// Recalculate Forecast
	_, err = a.CalculateForecast(userID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a *APIService) CopySalaryCosts(payload models.CopySalaryCosts, userID int64, salaryID int64) error {
	err := a.dbService.CopySalaryCosts(payload, userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.RefreshSalaryCostDetails(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	// Recalculate Forecast
	_, err = a.CalculateForecast(userID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a *APIService) applyCostCalculation(salaryCost *models.SalaryCost, skipPrevious bool) (*models.SalaryCost, error) {
	costDetails, err := a.dbService.ListSalaryCostDetails(salaryCost.ID)
	if err != nil {
		return nil, err
	}

	salaryCost.CalculatedCostDetails = costDetails

	currDate := time.Time(salaryCost.DBDate)

	var nextDetail *models.SalaryCostDetail
	var previousDetail *models.SalaryCostDetail
	var lastValidDetail *models.SalaryCostDetail
	var secondLastValidDetail *models.SalaryCostDetail

	for i := range costDetails {
		detail := &costDetails[i]
		if detail.IsExtraMonth && skipPrevious {
			continue
		}
		dt, err := time.Parse("2006-01", detail.Month)
		if err != nil {
			continue
		}
		currDateAsMonth := currDate.Format("2006-01")
		if currDateAsMonth == detail.Month || dt.After(currDate) {
			// Found the first future (or current) detail.
			if !skipPrevious && lastValidDetail != nil {
				// When including previous executions, surface the most recent
				// already accounted cost instead of jumping ahead.
				nextDetail = lastValidDetail
				previousDetail = secondLastValidDetail
			} else {
				nextDetail = detail
				previousDetail = lastValidDetail
			}
			break
		}
		// Track history to be able to surface the latest executed cost when requested.
		secondLastValidDetail = lastValidDetail
		lastValidDetail = detail
	}

	if nextDetail == nil {
		// We have not found any detail in the future.
		if !skipPrevious && lastValidDetail != nil {
			// Use the most recent executed cost.
			nextDetail = lastValidDetail
			previousDetail = secondLastValidDetail
		}
		// If skipPrevious is true we intentionally keep nextDetail nil here.
	}

	if previousDetail != nil {
		prevDt, err := time.Parse("2006-01", previousDetail.Month)
		if err == nil {
			prevAsDate := types.AsDate(prevDt)
			salaryCost.CalculatedPreviousExecutionDate = &prevAsDate
		}
	}

	if nextDetail != nil {
		dt, err := time.Parse("2006-01", nextDetail.Month)
		if err != nil {
			return nil, err
		}
		dtAsDate := types.AsDate(dt)
		salaryCost.CalculatedNextExecutionDate = &dtAsDate
		salaryCost.CalculatedNextCost = nextDetail.Amount
		if nextDetail.Divider > 0 {
			salaryCost.CalculatedAmount = nextDetail.Amount / uint64(nextDetail.Divider)
		}
	}

	return salaryCost, nil
}
