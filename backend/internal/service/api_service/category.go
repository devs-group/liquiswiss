package api_service

import (
	"context"
	"errors"
	"fmt"
	"liquiswiss/internal/events"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

var (
	// ErrCategoryGlobal marks attempts to delete a global preset category
	ErrCategoryGlobal = errors.New("global categories cannot be deleted")
	// ErrCategoryInUse marks deletes blocked by transactions still using the category
	ErrCategoryInUse = errors.New("category is still used by transactions")
)

func (a *APIService) ListCategories(ctx context.Context, userID, page, limit int64) ([]models.Category, int64, error) {
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

func (a *APIService) GetCategory(ctx context.Context, userID int64, categoryID int64) (*models.Category, error) {
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

func (a *APIService) CreateCategory(ctx context.Context, payload models.CreateCategory, userID *int64) (*models.Category, error) {
	categoryID, err := a.dbService.CreateCategory(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Global categories are created without a user; userID 0 still resolves
	// them because the query matches organisation_id IS NULL
	var scopeUserID int64
	if userID != nil {
		scopeUserID = *userID
	}
	category, err := a.dbService.GetCategory(scopeUserID, categoryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(category); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if userID != nil {
		a.notifyChange(ctx, *userID, "category", events.ActionCreated, categoryID)
	}
	return category, nil
}

func (a *APIService) UpdateCategory(ctx context.Context, payload models.UpdateCategory, userID int64, categoryID int64) (*models.Category, error) {
	err := a.dbService.UpdateCategory(payload, userID, categoryID)
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
	a.notifyChange(ctx, userID, "category", events.ActionUpdated, categoryID)
	return category, nil
}

// ReassignCategoryTransactions moves every transaction of the user's current
// organisation from one category to another, typically right before the
// source category gets deleted.
func (a *APIService) ReassignCategoryTransactions(ctx context.Context, userID int64, fromCategoryID int64, toCategoryID int64) (int64, error) {
	if fromCategoryID == toCategoryID {
		return 0, fmt.Errorf("source and target category must differ")
	}
	// Both categories must be visible to the user (own org or global preset)
	if _, err := a.dbService.GetCategory(userID, fromCategoryID); err != nil {
		logger.Logger.Error(err)
		return 0, err
	}
	if _, err := a.dbService.GetCategory(userID, toCategoryID); err != nil {
		logger.Logger.Error(err)
		return 0, err
	}
	affected, err := a.dbService.ReassignTransactionsCategory(userID, fromCategoryID, toCategoryID)
	if err != nil {
		logger.Logger.Error(err)
		return 0, err
	}
	if affected > 0 {
		if _, err := a.CalculateForecast(ctx, userID); err != nil {
			logger.Logger.Error(err)
		}
		a.notifyChange(ctx, userID, "transaction", events.ActionUpdated, 0)
	}
	return affected, nil
}

func (a *APIService) DeleteCategory(ctx context.Context, userID int64, categoryID int64) error {
	category, err := a.dbService.GetCategory(userID, categoryID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	if !category.CanEdit {
		return ErrCategoryGlobal
	}
	inUse, err := a.dbService.CountTransactionsWithCategory(userID, categoryID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	if inUse > 0 {
		return fmt.Errorf("%w: %d transaction(s)", ErrCategoryInUse, inUse)
	}
	err = a.dbService.DeleteCategory(userID, categoryID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	a.notifyChange(ctx, userID, "category", events.ActionDeleted, categoryID)
	return nil
}
