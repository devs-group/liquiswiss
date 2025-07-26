package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListBankAccounts(userID int64) ([]models.BankAccount, error) {
	bankAccounts, err := a.dbService.ListBankAccounts(userID)
	if err != nil {
		logger.Logger.Error(err)
		return bankAccounts, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(bankAccounts, "dive"); err != nil {
		logger.Logger.Error(err)
		return bankAccounts, err
	}
	return bankAccounts, nil
}

func (a *APIService) GetBankAccount(userID int64, bankAccountID int64) (*models.BankAccount, error) {
	bankAccount, err := a.dbService.GetBankAccount(userID, bankAccountID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(bankAccount); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return bankAccount, nil
}

func (a *APIService) CreateBankAccount(payload models.CreateBankAccount, userID int64) (*models.BankAccount, error) {
	bankAccountID, err := a.dbService.CreateBankAccount(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	bankAccount, err := a.dbService.GetBankAccount(userID, bankAccountID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(bankAccount); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return bankAccount, nil
}

func (a *APIService) UpdateBankAccount(payload models.UpdateBankAccount, userID int64, bankAccountID int64) (*models.BankAccount, error) {
	_, err := a.dbService.GetBankAccount(userID, bankAccountID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.UpdateBankAccount(payload, userID, bankAccountID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	bankAccount, err := a.dbService.GetBankAccount(userID, bankAccountID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(bankAccount); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return bankAccount, nil
}

func (a *APIService) DeleteBankAccount(userID int64, bankAccountID int64) error {
	_, err := a.dbService.GetBankAccount(userID, bankAccountID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.DeleteBankAccount(userID, bankAccountID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}
