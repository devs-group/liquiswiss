package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) GetVatSetting(userID int64) (*models.VatSetting, error) {
	vatSetting, err := a.dbService.GetVatSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if vatSetting == nil {
		return nil, nil
	}
	validator := utils.GetValidator()
	if err := validator.Struct(vatSetting); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return vatSetting, nil
}

func (a *APIService) CreateVatSetting(payload models.CreateVatSetting, userID int64) (*models.VatSetting, error) {
	// Check if a VAT setting already exists for this organization
	existingSetting, err := a.dbService.GetVatSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if existingSetting != nil {
		// If setting exists, update it instead
		updatePayload := models.UpdateVatSetting{
			Enabled:                &payload.Enabled,
			BillingDate:            &payload.BillingDate,
			TransactionMonthOffset: &payload.TransactionMonthOffset,
			Interval:               &payload.Interval,
		}
		return a.UpdateVatSetting(updatePayload, userID)
	}

	_, err = a.dbService.CreateVatSetting(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	vatSetting, err := a.dbService.GetVatSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(vatSetting); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return vatSetting, nil
}

func (a *APIService) UpdateVatSetting(payload models.UpdateVatSetting, userID int64) (*models.VatSetting, error) {
	_, err := a.dbService.GetVatSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.UpdateVatSetting(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	vatSetting, err := a.dbService.GetVatSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(vatSetting); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return vatSetting, nil
}

func (a *APIService) DeleteVatSetting(userID int64) error {
	_, err := a.dbService.GetVatSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.DeleteVatSetting(userID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}
