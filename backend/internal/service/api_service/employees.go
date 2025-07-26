package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListEmployees(userID int64, page int64, limit int64, sortBy string, sortOrder string) ([]models.Employee, int64, error) {
	employees, totalCount, err := a.dbService.ListEmployees(userID, page, limit, sortBy, sortOrder)
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

func (a *APIService) GetEmployee(userID int64, employeeID int64) (*models.Employee, error) {
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

func (a *APIService) CreateEmployee(payload models.CreateEmployee, userID int64) (*models.Employee, error) {
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
	return employee, nil
}

func (a *APIService) UpdateEmployee(payload models.UpdateEmployee, userID int64, employeeID int64) (*models.Employee, error) {
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
	return employee, nil
}

func (a *APIService) DeleteEmployee(employeeID int64, userID int64) error {
	existingEmployee, err := a.dbService.GetEmployee(userID, employeeID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	err = a.dbService.DeleteEmployee(existingEmployee.ID, userID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a *APIService) CountEmployees(userID int64, page int64, limit int64) (int64, error) {
	totalCount, err := a.dbService.CountEmployees(userID, page, limit)
	if err != nil {
		logger.Logger.Error(err)
		return 0, err
	}
	return totalCount, nil
}
