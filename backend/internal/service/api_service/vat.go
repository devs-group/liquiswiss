package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListVats(userID int64) ([]models.Vat, error) {
	vats, err := a.dbService.ListVats(userID)
	if err != nil {
		logger.Logger.Error(err)
		return vats, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(vats, "dive"); err != nil {
		logger.Logger.Error(err)
		return vats, err
	}
	return vats, nil
}

func (a *APIService) GetVat(userID int64, vatID int64) (*models.Vat, error) {
	vat, err := a.dbService.GetVat(userID, vatID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(vat); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return vat, nil
}

func (a *APIService) CreateVat(payload models.CreateVat, userID int64) (*models.Vat, error) {
	vatID, err := a.dbService.CreateVat(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	vat, err := a.dbService.GetVat(userID, vatID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(vat); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return vat, nil
}

func (a *APIService) UpdateVat(payload models.UpdateVat, userID int64, vatID int64) (*models.Vat, error) {
	_, err := a.dbService.GetVat(userID, vatID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.UpdateVat(payload, userID, vatID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	vat, err := a.dbService.GetVat(userID, vatID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(vat); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return vat, nil
}

func (a *APIService) DeleteVat(userID int64, vatID int64) error {
	_, err := a.dbService.GetVat(userID, vatID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.DeleteVat(userID, vatID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}
