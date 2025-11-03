package api_service

import (
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/types"
	"liquiswiss/pkg/utils"
)

func (a *APIService) ListSalaries(userID int64, employeeID int64, page int64, limit int64) ([]models.Salary, int64, error) {
	salaries, totalCount, err := a.dbService.ListSalaries(userID, employeeID, page, limit)
	if err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	for i := range salaries {
		salary := salaries[i]
		updatedSalary, err := a.applySalaryCalculations(userID, &salary)
		if err != nil {
			return nil, 0, err
		}
		salaries[i] = *updatedSalary
	}
	validator := utils.GetValidator()
	if err := validator.Var(salaries, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	return salaries, totalCount, nil
}

func (a *APIService) GetSalary(userID int64, salaryID int64) (*models.Salary, error) {
	salary, err := a.dbService.GetSalary(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(salary); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	salary, err = a.applySalaryCalculations(userID, salary)
	if err != nil {
		return nil, err
	}
	return salary, nil
}

func (a *APIService) CreateSalary(payload models.CreateSalary, userID int64, employeeID int64) (*models.Salary, error) {
	salaryID, previousSalaryID, nextSalaryID, err := a.dbService.CreateSalary(payload, userID, employeeID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Refresh all cost details
	err = a.dbService.RefreshSalaryCostDetails(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if previousSalaryID != 0 {
		err = a.dbService.RefreshSalaryCostDetails(userID, previousSalaryID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
	}
	if nextSalaryID != 0 {
		err = a.dbService.RefreshSalaryCostDetails(userID, nextSalaryID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
	}
	salary, err := a.GetSalary(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(salary); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Recalculate Forecast
	_, err = a.CalculateForecast(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return salary, nil
}

func (a *APIService) UpdateSalary(payload models.UpdateSalary, userID int64, salaryID int64) (*models.Salary, error) {
	existingSalary, err := a.GetSalary(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if payload.HoursPerMonth == nil {
		hoursPerMonth := existingSalary.HoursPerMonth
		payload.HoursPerMonth = &hoursPerMonth
	}
	if payload.Amount == nil {
		salary := existingSalary.Amount
		payload.Amount = &salary
	}
	if payload.CurrencyID == nil {
		payload.CurrencyID = existingSalary.Currency.ID
	}
	if payload.VacationDaysPerYear == nil {
		vacationDaysPerYear := existingSalary.VacationDaysPerYear
		payload.VacationDaysPerYear = &vacationDaysPerYear
	}
	if payload.FromDate == nil {
		fromDate := existingSalary.FromDate.ToString()
		payload.FromDate = &fromDate
	}
	previousSalaryID, nextSalaryID, err := a.dbService.UpdateSalary(payload, existingSalary.EmployeeID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Refresh all cost details
	err = a.dbService.RefreshSalaryCostDetails(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	if previousSalaryID != 0 {
		err = a.dbService.RefreshSalaryCostDetails(userID, previousSalaryID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
	}
	if nextSalaryID != 0 {
		err = a.dbService.RefreshSalaryCostDetails(userID, nextSalaryID)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
	}
	salary, err := a.GetSalary(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(salary); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Recalculate Forecast
	_, err = a.CalculateForecast(userID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return salary, nil
}

func (a *APIService) DeleteSalary(userID int64, salaryID int64) error {
	existingSalary, err := a.GetSalary(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	previousSalaryID, nextSalaryID, err := a.dbService.DeleteSalary(existingSalary, userID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	// Refresh all cost details
	err = a.dbService.RefreshSalaryCostDetails(userID, salaryID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	if previousSalaryID != 0 {
		err = a.dbService.RefreshSalaryCostDetails(userID, previousSalaryID)
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
	}
	if nextSalaryID != 0 {
		err = a.dbService.RefreshSalaryCostDetails(userID, nextSalaryID)
		if err != nil {
			logger.Logger.Error(err)
			return err
		}
	}
	// Recalculate Forecast
	_, err = a.CalculateForecast(userID)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a *APIService) applySalaryCalculations(userID int64, salary *models.Salary) (*models.Salary, error) {
	// Determine whether separate salary costs exist.
	salaryCosts, _, err := a.ListSalaryCosts(userID, salary.ID, 1, 1000, true)
	if err != nil {
		return nil, err
	}
	salary.HasSeparateCostsDefined = len(salaryCosts) > 0

	if salary.IsDisabled {
		salary.EmployeeDeductions = 0
		salary.EmployerCosts = 0
		salary.NextExecutionDate = nil
		return salary, nil
	}

	if salary.IsTermination {
		salary.Amount = 0
		salary.HoursPerMonth = 0
		salary.VacationDaysPerYear = 0
		salary.EmployeeDeductions = 0
		salary.EmployerCosts = 0
		salary.HasSeparateCostsDefined = false
		salary.NextExecutionDate = nil
		return salary, nil
	}

	employeeDeductions := a.CalculateSalaryAdjustments(
		salary.Cycle,
		models.SalaryCostDistributionEmployee,
		salaryCosts,
	)
	salary.EmployeeDeductions = employeeDeductions
	employerCosts := a.CalculateSalaryAdjustments(
		salary.Cycle,
		models.SalaryCostDistributionEmployer,
		salaryCosts,
	)
	salary.EmployerCosts = employerCosts
	nextExecutionDate := a.CalculateSalaryExecutionDate(salary.FromDate, salary.ToDate, &salary.Cycle, salary.DBDate, 1, true)
	if nextExecutionDate != nil {
		nextSalaryExecutionDateAsDate := types.AsDate(*nextExecutionDate)
		salary.NextExecutionDate = &nextSalaryExecutionDateAsDate
	}
	return salary, nil
}
