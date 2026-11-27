package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func currentScenario(t *testing.T, env *CrossOrgTestEnv, userID int64) models.Scenario {
	t.Helper()
	scenarios, err := env.APIService.ListScenarios(context.Background(), userID)
	require.NoError(t, err)
	for _, s := range scenarios {
		if s.IsCurrent {
			return s
		}
	}
	t.Fatalf("no current scenario for user %d", userID)
	return models.Scenario{}
}

func TestScenario_NewOrganisationGetsBaseScenario(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()

	scenarios, err := env.APIService.ListScenarios(context.Background(), env.UserA.ID)
	require.NoError(t, err)
	require.Len(t, scenarios, 1)
	require.True(t, scenarios[0].IsBase)
	require.True(t, scenarios[0].IsCurrent)
	require.Nil(t, scenarios[0].ParentID)
}

func TestScenario_CopyIsIndependentFromSource(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()
	ctx := context.Background()
	user := env.UserA.ID
	base := currentScenario(t, env, user)

	// Build a base scenario with every scenario-bound entity
	employee, err := CreateEmployee(env.APIService, user, "Anna")
	require.NoError(t, err)
	salary, err := env.APIService.CreateSalary(ctx, models.CreateSalary{
		HoursPerMonth: 160, Amount: 5000_00, Cycle: utils.CycleMonthly, CurrencyID: *env.Currency.ID,
		VacationDaysPerYear: 25, FromDate: "2025-01-01",
	}, user, employee.ID)
	require.NoError(t, err)
	label, err := CreateSalaryCostLabel(env.APIService, user, "AHV")
	require.NoError(t, err)
	baseCost, err := env.APIService.CreateSalaryCost(ctx, models.CreateSalaryCost{
		Cycle: utils.CycleMonthly, AmountType: "fixed", Amount: 100_00, DistributionType: "employee", RelativeOffset: 1,
		LabelID: &label.ID,
	}, user, salary.ID)
	require.NoError(t, err)
	_, err = env.APIService.CreateSalaryCost(ctx, models.CreateSalaryCost{
		Cycle: utils.CycleMonthly, AmountType: "percentage", Amount: 5_000, DistributionType: "employer", RelativeOffset: 1,
		BaseSalaryCostIDs: []int64{baseCost.ID},
	}, user, salary.ID)
	require.NoError(t, err)
	category, err := env.APIService.CreateCategory(ctx, models.CreateCategory{Name: "Löhne"}, &user)
	require.NoError(t, err)
	transaction, err := env.APIService.CreateTransaction(ctx, models.CreateTransaction{
		Name: "Bonus", Amount: -1000_00, Type: "single", StartDate: "2025-06-01",
		Category: category.ID, Currency: *env.Currency.ID, Employee: &employee.ID,
	}, user)
	require.NoError(t, err)
	_, err = env.APIService.CreateForecastExclusion(ctx, models.CreateForecastExclusion{
		RelatedID: transaction.ID, RelatedTable: "transactions", Month: "2025-06",
	}, user)
	require.NoError(t, err)
	bankAccount, err := env.APIService.CreateBankAccount(ctx, models.CreateBankAccount{
		Name: "UBS", Amount: 50_000_00, Currency: *env.Currency.ID,
	}, user)
	require.NoError(t, err)

	// Creating copies the current scenario and switches into the copy
	copied, err := env.APIService.CreateScenario(ctx, models.CreateScenario{Name: "Pessimistisch"}, user)
	require.NoError(t, err)
	require.False(t, copied.IsBase)
	require.True(t, copied.IsCurrent)
	require.NotNil(t, copied.ParentID)
	require.Equal(t, base.ID, *copied.ParentID)

	employees, _, err := env.APIService.ListEmployees(ctx, user, 1, 100, "name", "ASC", "", false)
	require.NoError(t, err)
	require.Len(t, employees, 1)
	require.Equal(t, "Anna", employees[0].Name)
	require.NotEqual(t, employee.ID, employees[0].ID)
	copiedEmployeeID := employees[0].ID

	salaries, _, err := env.APIService.ListSalaries(ctx, user, copiedEmployeeID, 1, 100)
	require.NoError(t, err)
	require.Len(t, salaries, 1)
	require.NotEqual(t, salary.ID, salaries[0].ID)
	costs, _, err := env.APIService.ListSalaryCosts(ctx, user, salaries[0].ID, 1, 100, false)
	require.NoError(t, err)
	require.Len(t, costs, 2)

	transactions, _, err := env.APIService.ListTransactions(ctx, user, 1, 100, "name", "ASC", "", false, false)
	require.NoError(t, err)
	require.Len(t, transactions, 1)
	require.NotEqual(t, transaction.ID, transactions[0].ID)
	require.NotNil(t, transactions[0].Employee)
	require.Equal(t, copiedEmployeeID, transactions[0].Employee.ID, "transaction must point to the copied employee")
	exclusions, err := env.APIService.ListForecastExclusions(ctx, user, transactions[0].ID, "transactions")
	require.NoError(t, err)
	require.True(t, exclusions["2025-06"], "exclusions are copied with the transaction")

	accounts, _, err := env.APIService.ListBankAccounts(ctx, user, 1, 100, "name", "ASC", "")
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	require.NotEqual(t, bankAccount.ID, accounts[0].ID)

	// The source rows are invisible and untouchable from the copy
	_, err = env.APIService.GetBankAccount(ctx, user, bankAccount.ID)
	require.Error(t, err)
	_, err = env.APIService.GetEmployee(ctx, user, employee.ID)
	require.Error(t, err)

	// Changing the copy leaves the base untouched
	newAmount := int64(10_00)
	_, err = env.APIService.UpdateBankAccount(ctx, models.UpdateBankAccount{Amount: &newAmount}, user, accounts[0].ID)
	require.NoError(t, err)
	err = env.APIService.DeleteTransaction(ctx, user, transactions[0].ID)
	require.NoError(t, err)

	_, err = env.APIService.SwitchScenario(ctx, user, base.ID)
	require.NoError(t, err)
	baseAccount, err := env.APIService.GetBankAccount(ctx, user, bankAccount.ID)
	require.NoError(t, err)
	require.Equal(t, int64(50_000_00), baseAccount.Amount)
	_, err = env.APIService.GetTransaction(ctx, user, transaction.ID)
	require.NoError(t, err)
}

func TestScenario_CopyFromExplicitSource(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()
	ctx := context.Background()
	user := env.UserA.ID
	base := currentScenario(t, env, user)

	first, err := env.APIService.CreateScenario(ctx, models.CreateScenario{Name: "Erster"}, user)
	require.NoError(t, err)
	_, err = env.APIService.CreateBankAccount(ctx, models.CreateBankAccount{Name: "Nur im Ersten", Amount: 1_00, Currency: *env.Currency.ID}, user)
	require.NoError(t, err)

	// Copy from the base while the first scenario is active
	second, err := env.APIService.CreateScenario(ctx, models.CreateScenario{Name: "Zweiter", SourceScenarioID: &base.ID}, user)
	require.NoError(t, err)
	require.Equal(t, base.ID, *second.ParentID)
	accounts, _, err := env.APIService.ListBankAccounts(ctx, user, 1, 100, "name", "ASC", "")
	require.NoError(t, err)
	require.Empty(t, accounts)

	// And from the first one
	third, err := env.APIService.CreateScenario(ctx, models.CreateScenario{Name: "Dritter", SourceScenarioID: &first.ID}, user)
	require.NoError(t, err)
	require.Equal(t, first.ID, *third.ParentID)
	accounts, _, err = env.APIService.ListBankAccounts(ctx, user, 1, 100, "name", "ASC", "")
	require.NoError(t, err)
	require.Len(t, accounts, 1)
}

func TestScenario_DeleteRules(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()
	ctx := context.Background()
	user := env.UserA.ID
	base := currentScenario(t, env, user)

	err := env.APIService.DeleteScenario(ctx, user, base.ID)
	require.ErrorIs(t, err, api_service.ErrBaseScenario)

	copied, err := env.APIService.CreateScenario(ctx, models.CreateScenario{Name: "Weg damit"}, user)
	require.NoError(t, err)
	_, err = env.APIService.CreateBankAccount(ctx, models.CreateBankAccount{Name: "Temp", Amount: 1_00, Currency: *env.Currency.ID}, user)
	require.NoError(t, err)

	// Deleting the active scenario falls back to the base and removes its data
	err = env.APIService.DeleteScenario(ctx, user, copied.ID)
	require.NoError(t, err)
	require.Equal(t, base.ID, currentScenario(t, env, user).ID)
	accounts, _, err := env.APIService.ListBankAccounts(ctx, user, 1, 100, "name", "ASC", "")
	require.NoError(t, err)
	require.Empty(t, accounts)
}

func TestScenario_UsersSwitchIndependently(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()
	ctx := context.Background()

	// A second member of organisation A
	memberID, err := env.DBAdapter.CreateUser("member@scenario-test.com", "test")
	require.NoError(t, err)
	err = env.DBAdapter.AssignUserToOrganisation(memberID, env.OrgA.ID, "editor", false)
	require.NoError(t, err)
	err = env.DBAdapter.SetUserCurrentOrganisation(memberID, env.OrgA.ID)
	require.NoError(t, err)
	base := currentScenario(t, env, memberID)

	copied, err := env.APIService.CreateScenario(ctx, models.CreateScenario{Name: "Nur für A"}, env.UserA.ID)
	require.NoError(t, err)

	// The member still works in the base and sees the new scenario in the list
	require.Equal(t, base.ID, currentScenario(t, env, memberID).ID)
	_, err = env.APIService.SwitchScenario(ctx, memberID, copied.ID)
	require.NoError(t, err)
	require.Equal(t, copied.ID, currentScenario(t, env, memberID).ID)
}

func TestScenario_ForecastOfAnotherScenario(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()
	ctx := context.Background()
	user := env.UserA.ID
	base := currentScenario(t, env, user)

	category, err := env.APIService.CreateCategory(ctx, models.CreateCategory{Name: "Projekte"}, &user)
	require.NoError(t, err)
	inTwoMonths := time.Now().AddDate(0, 2, 0).Format("2006-01-02")

	// The copy gets an extra revenue and its own bank account
	copied, err := env.APIService.CreateScenario(ctx, models.CreateScenario{Name: "Mehr Umsatz"}, user)
	require.NoError(t, err)
	_, err = env.APIService.CreateTransaction(ctx, models.CreateTransaction{
		Name: "Zusatzauftrag", Amount: 5000_00, Type: "single", StartDate: inTwoMonths,
		Category: category.ID, Currency: *env.Currency.ID,
	}, user)
	require.NoError(t, err)
	_, err = env.APIService.CreateBankAccount(ctx, models.CreateBankAccount{Name: "Konto Kopie", Amount: 100_00, Currency: *env.Currency.ID}, user)
	require.NoError(t, err)

	// Back in the base, the copy can be read without switching to it
	_, err = env.APIService.SwitchScenario(ctx, user, base.ID)
	require.NoError(t, err)
	compared, err := env.APIService.GetScenarioForecast(ctx, user, copied.ID, 4)
	require.NoError(t, err)
	require.Equal(t, copied.ID, compared.Scenario.ID)
	require.Len(t, compared.Forecasts, 4)
	require.Len(t, compared.ForecastDetails, 4)
	require.Len(t, compared.BankAccounts, 1)
	revenue := int64(0)
	for _, f := range compared.Forecasts {
		revenue += f.Data.Revenue
	}
	require.Equal(t, int64(5000_00), revenue)
	require.Equal(t, base.ID, currentScenario(t, env, user).ID, "reading a scenario must not switch to it")

	own, err := env.APIService.ListForecasts(ctx, user, 4)
	require.NoError(t, err)
	for _, f := range own {
		require.Equal(t, int64(0), f.Data.Revenue)
	}

	// Another organisation cannot read it
	_, err = env.APIService.GetScenarioForecast(ctx, env.UserB.ID, copied.ID, 4)
	require.Error(t, err)
}

func TestScenario_CrossOrgIsolation(t *testing.T) {
	env := SetupCrossOrgTestEnvironment(t)
	defer env.Conn.Close()
	ctx := context.Background()

	baseA := currentScenario(t, env, env.UserA.ID)
	copiedA, err := env.APIService.CreateScenario(ctx, models.CreateScenario{Name: "Geheim"}, env.UserA.ID)
	require.NoError(t, err)

	scenariosB, err := env.APIService.ListScenarios(ctx, env.UserB.ID)
	require.NoError(t, err)
	require.Len(t, scenariosB, 1)
	require.NotEqual(t, baseA.ID, scenariosB[0].ID)

	_, err = env.APIService.GetScenario(ctx, env.UserB.ID, copiedA.ID)
	require.Error(t, err)
	_, err = env.APIService.SwitchScenario(ctx, env.UserB.ID, copiedA.ID)
	require.Error(t, err)
	_, err = env.APIService.CreateScenario(ctx, models.CreateScenario{Name: "Klau", SourceScenarioID: &baseA.ID}, env.UserB.ID)
	require.Error(t, err)
	newName := "Übernommen"
	_, err = env.APIService.UpdateScenario(ctx, models.UpdateScenario{Name: &newName}, env.UserB.ID, copiedA.ID)
	require.Error(t, err)
	err = env.APIService.DeleteScenario(ctx, env.UserB.ID, copiedA.ID)
	require.Error(t, err)

	stillThere, err := env.APIService.GetScenario(ctx, env.UserA.ID, copiedA.ID)
	require.NoError(t, err)
	require.Equal(t, "Geheim", stillThere.Name)
}
