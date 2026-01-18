package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListTransactions(userID int64, page int64, limit int64, sortBy string, sortOrder string, search string) ([]models.Transaction, int64, error) {
	transactions, totalCount, err := a.dbService.ListTransactions(userID, page, limit, sortBy, sortOrder, search)
	if err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(transactions, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	return transactions, totalCount, nil
}

func (a *APIService) GetTransaction(userID int64, transactionID int64) (*models.Transaction, error) {
	transaction, err := a.dbService.GetTransaction(userID, transactionID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(transaction, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return transaction, nil
}

func (a *APIService) CreateTransaction(payload models.CreateTransaction, userID int64) (*models.Transaction, error) {
	transactionID, err := a.dbService.CreateTransaction(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	transaction, err := a.dbService.GetTransaction(userID, transactionID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(transaction, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Recalculate Forecast
	_, err = a.CalculateForecast(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return transaction, nil
}

func (a *APIService) UpdateTransaction(payload models.UpdateTransaction, userID int64, transactionID int64) (*models.Transaction, error) {
	existingTransaction, err := a.dbService.GetTransaction(userID, transactionID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Set the value of StartDate to be able to make comparisons in case it is not already set for the update
	if payload.StartDate == nil {
		startDate := existingTransaction.StartDate.ToString()
		payload.StartDate = &startDate
	}
	if payload.Cycle == nil {
		if existingTransaction.Cycle != nil {
			cycle := *existingTransaction.Cycle
			payload.Cycle = &cycle
		}
	} else if *payload.Cycle == "" {
		payload.Cycle = nil
	}
	err = a.dbService.UpdateTransaction(payload, userID, transactionID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	transaction, err := a.dbService.GetTransaction(userID, transactionID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(transaction, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Recalculate Forecast
	_, err = a.CalculateForecast(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return transaction, nil
}

func (a *APIService) DeleteTransaction(userID int64, transactionID int64) error {
	_, err := a.dbService.GetTransaction(userID, transactionID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.DeleteTransaction(userID, transactionID)
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
