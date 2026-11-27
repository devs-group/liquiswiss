package handlers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func TestDeleteEmployee_RemovesSalaryFromForecast(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()
	ctx := context.Background()
	user := env.UserA.ID

	employee, err := CreateEmployee(env.APIService, user, "Anna")
	require.NoError(t, err)
	_, err = env.APIService.CreateSalary(ctx, models.CreateSalary{
		HoursPerMonth: 160, Amount: 5000_00, Cycle: utils.CycleMonthly, CurrencyID: *env.Currency.ID,
		VacationDaysPerYear: 25, FromDate: "2025-01-01",
	}, user, employee.ID)
	require.NoError(t, err)

	totalExpense := func() int64 {
		forecasts, err := env.APIService.ListForecasts(ctx, user, 13)
		require.NoError(t, err)
		sum := int64(0)
		for _, f := range forecasts {
			sum += f.Data.Expense
		}
		return sum
	}
	require.Less(t, totalExpense(), int64(0), "salary must show up in the forecast")

	require.NoError(t, env.APIService.DeleteEmployee(ctx, user, employee.ID))
	require.Equal(t, int64(0), totalExpense(), "deleted employee's salary must leave the forecast")
}
