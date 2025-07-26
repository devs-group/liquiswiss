package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListSalaryCostLabels(userID int64, page int64, limit int64) ([]models.SalaryCostLabel, int64, error) {
	salaryCostLabels, totalCount, err := a.dbService.ListSalaryCostLabels(userID, page, limit)
	if err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(salaryCostLabels, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	return salaryCostLabels, totalCount, nil
}

func (a *APIService) GetSalaryCostLabel(userID int64, salaryCostLabelID int64) (*models.SalaryCostLabel, error) {
	salaryCostLabel, err := a.dbService.GetSalaryCostLabel(userID, salaryCostLabelID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(salaryCostLabel); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return salaryCostLabel, nil
}

func (a *APIService) CreateSalaryCostLabel(payload models.CreateSalaryCostLabel, userID int64) (*models.SalaryCostLabel, error) {
	salaryCostLabelID, err := a.dbService.CreateSalaryCostLabel(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	salaryCostLabel, err := a.dbService.GetSalaryCostLabel(userID, salaryCostLabelID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(salaryCostLabel); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return salaryCostLabel, nil
}

func (a *APIService) UpdateSalaryCostLabel(payload models.CreateSalaryCostLabel, userID int64, salaryCostLabelID int64) (*models.SalaryCostLabel, error) {
	_, err := a.dbService.GetSalaryCostLabel(userID, salaryCostLabelID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.UpdateSalaryCostLabel(payload, userID, salaryCostLabelID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	salaryCostLabel, err := a.dbService.GetSalaryCostLabel(userID, salaryCostLabelID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(salaryCostLabel); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return salaryCostLabel, nil
}

func (a *APIService) DeleteSalaryCostLabel(userID int64, salaryCostLabelID int64) error {
	existingSalaryCostLabel, err := a.dbService.GetSalaryCostLabel(userID, salaryCostLabelID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.DeleteSalaryCostLabel(userID, existingSalaryCostLabel.ID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}
