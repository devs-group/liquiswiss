package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
	"time"
)

func (a *APIService) ListSalaryCosts(userID int64, salaryID int64, page int64, limit int64) ([]models.SalaryCost, int64, error) {
	salaryCosts, totalCount, err := a.dbService.ListSalaryCosts(userID, salaryID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range salaryCosts {
		salaryCost := salaryCosts[i]
		updatedSalaryCost, err := a.applyCostCalculation(&salaryCost)
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

func (a *APIService) GetSalaryCost(userID int64, salaryCostID int64) (*models.SalaryCost, error) {
	salaryCost, err := a.dbService.GetSalaryCost(userID, salaryCostID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	salaryCost, err = a.applyCostCalculation(salaryCost)
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
	salaryCost, err := a.GetSalaryCost(userID, salaryCostID)
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
	_, err := a.GetSalaryCost(userID, salaryCostID)
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
	salaryCost, err := a.GetSalaryCost(userID, salaryCostID)
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
	existingSalaryCost, err := a.GetSalaryCost(userID, salaryCostID)
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

func (a *APIService) applyCostCalculation(salaryCost *models.SalaryCost) (*models.SalaryCost, error) {
	costDetails, err := a.dbService.ListSalaryCostDetails(salaryCost.ID)
	if err != nil {
		return nil, err
	}

	salaryCost.CalculatedCostDetails = costDetails

	currDate := time.Time(salaryCost.DBDate)

	var next *models.SalaryCostDetail
	for i := range costDetails {
		costDetail := costDetails[i]
		dt, err := time.Parse("2006-01", costDetail.Month)
		if err != nil {
			continue
		}
		if currDate.Format("2006-01") == costDetail.Month || dt.After(currDate) {
			next = &costDetail
			break
		}
	}

	if next != nil {
		dt, err := time.Parse("2006-01", next.Month)
		if err != nil {
			return nil, err
		}
		dtAsDate := types.AsDate(dt)
		//salaryCost.CalculatedPreviousExecutionDate = ??
		salaryCost.CalculatedNextExecutionDate = &dtAsDate
		salaryCost.CalculatedNextCost = next.Amount
		salaryCost.CalculatedAmount = next.Amount / uint64(next.Divider)
	}
	return salaryCost, nil
}
