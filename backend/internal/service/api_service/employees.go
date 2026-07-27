package api_service

import (
	"context"
	"liquiswiss/internal/events"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListEmployees(ctx context.Context, userID int64, page int64, limit int64, sortBy string, sortOrder string, search string, hideTerminated bool) ([]models.Employee, int64, error) {
	employees, totalCount, err := a.dbService.ListEmployees(userID, page, limit, sortBy, sortOrder, search, hideTerminated)
	if err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(employees, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	return employees, totalCount, nil
}

func (a *APIService) GetEmployee(ctx context.Context, userID int64, employeeID int64) (*models.Employee, error) {
	employee, err := a.dbService.GetEmployee(userID, employeeID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(employee); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return employee, nil
}

func (a *APIService) CreateEmployee(ctx context.Context, payload models.CreateEmployee, userID int64) (*models.Employee, error) {
	employeeID, err := a.dbService.CreateEmployee(payload, userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	employee, err := a.dbService.GetEmployee(userID, employeeID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(employee); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	a.notifyChange(ctx, userID, "employee", events.ActionCreated, employeeID)
	return employee, nil
}

func (a *APIService) UpdateEmployee(ctx context.Context, payload models.UpdateEmployee, userID int64, employeeID int64) (*models.Employee, error) {
	existingEmployee, err := a.dbService.GetEmployee(userID, employeeID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if payload.Name == nil {
		name := existingEmployee.Name
		payload.Name = &name
	}
	err = a.dbService.UpdateEmployee(payload, userID, employeeID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	employee, err := a.dbService.GetEmployee(userID, employeeID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(employee); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	a.notifyChange(ctx, userID, "employee", events.ActionUpdated, employeeID)
	return employee, nil
}

func (a *APIService) DeleteEmployee(ctx context.Context, userID int64, employeeID int64) error {
	existingEmployee, err := a.dbService.GetEmployee(userID, employeeID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.DeleteEmployee(userID, existingEmployee.ID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	a.notifyChange(ctx, userID, "employee", events.ActionDeleted, existingEmployee.ID)
	return nil
}

func (a *APIService) CountEmployees(ctx context.Context, userID int64, page int64, limit int64) (int64, error) {
	totalCount, err := a.dbService.CountEmployees(userID, page, limit)
	if err != nil {
		logger.Logger.Error(err)
		return 0, err
	}
	return totalCount, nil
}
