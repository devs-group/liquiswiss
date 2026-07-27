package api_service

import (
	"context"
	"liquiswiss/internal/events"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListBankAccounts(ctx context.Context, userID int64, page int64, limit int64, sortBy string, sortOrder string, search string) ([]models.BankAccount, int64, error) {
	bankAccounts, totalCount, err := a.dbService.ListBankAccounts(userID, page, limit, sortBy, sortOrder, search)
	if err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(bankAccounts, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	return bankAccounts, totalCount, nil
}

func (a *APIService) GetBankAccount(ctx context.Context, userID int64, bankAccountID int64) (*models.BankAccount, error) {
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

func (a *APIService) CreateBankAccount(ctx context.Context, payload models.CreateBankAccount, userID int64) (*models.BankAccount, error) {
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
	a.notifyChange(ctx, userID, "bank_account", events.ActionCreated, bankAccountID)
	return bankAccount, nil
}

func (a *APIService) UpdateBankAccount(ctx context.Context, payload models.UpdateBankAccount, userID int64, bankAccountID int64) (*models.BankAccount, error) {
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
	a.notifyChange(ctx, userID, "bank_account", events.ActionUpdated, bankAccountID)
	return bankAccount, nil
}

func (a *APIService) DeleteBankAccount(ctx context.Context, userID int64, bankAccountID int64) error {
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
	a.notifyChange(ctx, userID, "bank_account", events.ActionDeleted, bankAccountID)
	return nil
}
