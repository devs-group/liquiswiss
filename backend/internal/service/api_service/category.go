package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListCategories(userID, page, limit int64) ([]models.Category, int64, error) {
	categories, totalCount, err := a.dbService.ListCategories(userID, page, limit)
	if err != nil {
		logger.Logger.Error(err)
		return categories, totalCount, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(categories, "dive"); err != nil {
		logger.Logger.Error(err)
		return categories, totalCount, err
	}
	return categories, totalCount, nil
}

func (a *APIService) GetCategory(userID int64, categoryID int64) (*models.Category, error) {
	category, err := a.dbService.GetCategory(userID, categoryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(category); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return category, nil
}

func (a *APIService) CreateCategory(payload models.CreateCategory, userID *int64) (*models.Category, error) {
	categoryID, err := a.dbService.CreateCategory(userID, payload)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	category, err := a.dbService.GetCategory(*userID, categoryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(category); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return category, nil
}

func (a *APIService) UpdateCategory(payload models.UpdateCategory, userID int64, categoryID int64) (*models.Category, error) {
	err := a.dbService.UpdateCategory(userID, payload, categoryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	category, err := a.dbService.GetCategory(userID, categoryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(category); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return category, nil
}
