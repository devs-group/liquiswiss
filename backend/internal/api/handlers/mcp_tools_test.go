package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/pkg/models"
)

// mcpAccessToken runs the full OAuth flow and returns a valid MCP access token
func (env *oauthTestEnv) mcpAccessToken(t *testing.T) string {
	t.Helper()
	redirectURI := "http://localhost:33418/callback"
	clientID := env.registerClient(t, redirectURI)
	verifier, challenge := pkcePair()
	code := env.approve(t, clientID, redirectURI, challenge)

	status, tokens := env.tokenRequest(t, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
	})
	require.Equal(t, http.StatusOK, status)
	accessToken, _ := tokens["access_token"].(string)
	require.NotEmpty(t, accessToken)
	return accessToken
}

type mcpToolResult struct {
	IsError    bool
	Text       string
	Structured map[string]any
}

// mcpCall performs a raw JSON-RPC tools/call against /api/mcp. args must be a
// JSON object literal so type-level cases (strings, floats, scientific
// notation) reach the server exactly as written.
func (env *oauthTestEnv) mcpCall(t *testing.T, token, tool, args string) mcpToolResult {
	t.Helper()
	body := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":%q,"arguments":%s}}`, tool, args)
	req, _ := http.NewRequest(http.MethodPost, "/api/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	env.API.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var rpc struct {
		Result struct {
			IsError bool `json:"isError"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			StructuredContent map[string]any `json:"structuredContent"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rpc), w.Body.String())

	result := mcpToolResult{IsError: rpc.Result.IsError, Structured: rpc.Result.StructuredContent}
	if len(rpc.Result.Content) > 0 {
		result.Text = rpc.Result.Content[0].Text
	}
	return result
}

// allNullUpdateTransactionArgs returns update_transaction arguments with every
// nullable field null except the given overrides (raw JSON per field)
func allNullUpdateTransactionArgs(id int64, overrides map[string]string) string {
	fields := map[string]string{
		"name": "null", "link": "null", "amount": "null", "cycle": "null",
		"type": "null", "startDate": "null", "endDate": "null", "category": "null",
		"currency": "null", "employee": "null", "vat": "null", "vatIncluded": "null",
		"isDisabled": "null",
	}
	for k, v := range overrides {
		fields[k] = v
	}
	parts := []string{fmt.Sprintf(`"id":%d`, id)}
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%q:%s", k, v))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func TestMCPBankAccountCRUD(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	created := env.mcpCall(t, token, "create_bank_account", `{"name":"MCP Konto","amount":5000000,"currency":1}`)
	require.False(t, created.IsError, created.Text)
	accountID := int64(created.Structured["id"].(float64))

	updated := env.mcpCall(t, token, "update_bank_account",
		fmt.Sprintf(`{"id":%d,"name":"MCP Konto neu","amount":-250000,"currency":null}`, accountID))
	require.False(t, updated.IsError, updated.Text)
	require.Equal(t, "MCP Konto neu", updated.Structured["name"])
	require.Equal(t, float64(-250000), updated.Structured["amount"], "negative balances (overdraft) are allowed")

	listed := env.mcpCall(t, token, "list_bank_accounts", `{}`)
	require.False(t, listed.IsError, listed.Text)
	require.Equal(t, float64(1), listed.Structured["total"])

	deleted := env.mcpCall(t, token, "delete_bank_account", fmt.Sprintf(`{"id":%d}`, accountID))
	require.False(t, deleted.IsError, deleted.Text)

	listed = env.mcpCall(t, token, "list_bank_accounts", `{}`)
	require.Equal(t, float64(0), listed.Structured["total"])
}

func TestMCPTransactionCRUDAndPatchSemantics(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	employee := env.mcpCall(t, token, "create_employee", `{"name":"MCP Angestellter"}`)
	require.False(t, employee.IsError, employee.Text)
	employeeID := int64(employee.Structured["id"].(float64))

	created := env.mcpCall(t, token, "create_transaction", fmt.Sprintf(
		`{"name":"MCP Umsatz","link":"https://example.com/x","amount":150000,"cycle":"monthly","type":"repeating","startDate":"2026-08-01","endDate":"2027-08-01","category":1,"currency":1,"employee":%d,"vat":1,"VatIncluded":false}`,
		employeeID))
	require.False(t, created.IsError, created.Text)
	transactionID := int64(created.Structured["id"].(float64))
	require.NotNil(t, created.Structured["employee"])
	require.NotNil(t, created.Structured["vat"])

	// PATCH semantics 1: isDisabled provided -> nil nullable fields are PRESERVED
	preserved := env.mcpCall(t, token, "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{
		"amount":     "175000",
		"isDisabled": "false",
	}))
	require.False(t, preserved.IsError, preserved.Text)
	require.Equal(t, float64(175000), preserved.Structured["amount"])
	require.NotNil(t, preserved.Structured["employee"], "employee must survive when isDisabled is provided")
	require.NotNil(t, preserved.Structured["vat"], "vat must survive when isDisabled is provided")
	require.NotNil(t, preserved.Structured["link"], "link must survive when isDisabled is provided")
	require.NotNil(t, preserved.Structured["endDate"], "endDate must survive when isDisabled is provided")

	// PATCH semantics 2: isDisabled absent -> nil nullable fields are CLEARED
	cleared := env.mcpCall(t, token, "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{
		"type": `"single"`,
	}))
	require.False(t, cleared.IsError, cleared.Text)
	require.Equal(t, "single", cleared.Structured["type"])
	require.Nil(t, cleared.Structured["employee"], "employee must be cleared when isDisabled is absent")
	require.Nil(t, cleared.Structured["vat"], "vat must be cleared when isDisabled is absent")
	require.Nil(t, cleared.Structured["link"], "link must be cleared when isDisabled is absent")
	require.Nil(t, cleared.Structured["endDate"], "endDate must be cleared when isDisabled is absent")

	deleted := env.mcpCall(t, token, "delete_transaction", fmt.Sprintf(`{"id":%d}`, transactionID))
	require.False(t, deleted.IsError, deleted.Text)

	notFound := env.mcpCall(t, token, "get_transaction", fmt.Sprintf(`{"id":%d}`, transactionID))
	require.True(t, notFound.IsError)
}

func TestMCPTransactionValidationRejections(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	created := env.mcpCall(t, token, "create_transaction",
		`{"name":"Basis","link":null,"amount":-10000,"cycle":"monthly","type":"repeating","startDate":"2026-08-01","endDate":null,"category":1,"currency":1,"employee":null,"vat":null,"VatIncluded":false}`)
	require.False(t, created.IsError, created.Text)
	transactionID := int64(created.Structured["id"].(float64))

	cases := []struct {
		name     string
		tool     string
		args     string
		errraint string
	}{
		{"daily cycle invalid", "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{"cycle": `"daily"`, "type": `"repeating"`}), "allowedCycles"},
		{"invalid type", "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{"type": `"sometimes"`}), "oneof"},
		{"endDate before startDate", "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{"startDate": `"2026-10-01"`, "endDate": `"2026-09-01"`}), "endDateGTEStartDate"},
		{"unknown category", "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{"category": "99999"}), "invalid category"},
		{"unknown employee", "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{"employee": "99999"}), "invalid employee"},
		{"missing cycle for repeating create", "create_transaction",
			`{"name":"Fehlender Cycle","link":null,"amount":-10000,"cycle":null,"type":"repeating","startDate":"2026-08-01","endDate":null,"category":1,"currency":1,"employee":null,"vat":null,"VatIncluded":false}`,
			"cycleRequiredIfRepeating"},
		{"bad date format", "create_transaction",
			`{"name":"Kaputtes Datum","link":null,"amount":-10000,"cycle":null,"type":"single","startDate":"01.08.2026","endDate":null,"category":1,"currency":1,"employee":null,"vat":null,"VatIncluded":false}`,
			"Incorrect date value"},
		{"amount as string", "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{"amount": `"abc"`}), "type"},
		{"amount as float", "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{"amount": "12.5"}), "type"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := env.mcpCall(t, token, testCase.tool, testCase.args)
			require.True(t, result.IsError, "expected error, got: %v", result.Structured)
			require.Contains(t, result.Text, testCase.errraint)
		})
	}

	// Scientific notation for integral values is valid JSON and must be accepted
	scientific := env.mcpCall(t, token, "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{
		"amount": "-45e3", "cycle": `"monthly"`, "isDisabled": "false",
	}))
	require.False(t, scientific.IsError, scientific.Text)
	require.Equal(t, float64(-45000), scientific.Structured["amount"])
}

func TestMCPVatLifecycleAndGlobalProtection(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	// Invalid values rejected
	negative := env.mcpCall(t, token, "create_vat", `{"value":-500}`)
	require.True(t, negative.IsError)
	zero := env.mcpCall(t, token, "create_vat", `{"value":0}`)
	require.True(t, zero.IsError)

	// Global VAT (id 1, organisation NULL) must not be touchable
	globalUpdate := env.mcpCall(t, token, "update_vat", `{"id":1,"value":999}`)
	require.True(t, globalUpdate.IsError)
	require.Contains(t, globalUpdate.Text, "global")
	globalDelete := env.mcpCall(t, token, "delete_vat", `{"id":1}`)
	require.True(t, globalDelete.IsError)
	require.Contains(t, globalDelete.Text, "global")

	// Own VAT lifecycle
	created := env.mcpCall(t, token, "create_vat", `{"value":380}`)
	require.False(t, created.IsError, created.Text)
	vatID := int64(created.Structured["id"].(float64))
	require.Equal(t, true, created.Structured["canEdit"])

	updated := env.mcpCall(t, token, "update_vat", fmt.Sprintf(`{"id":%d,"value":260}`, vatID))
	require.False(t, updated.IsError, updated.Text)
	require.Equal(t, "2.6%", updated.Structured["formattedValue"])

	// Assign to a transaction, then delete the VAT -> transaction must be unlinked
	transaction := env.mcpCall(t, token, "create_transaction", fmt.Sprintf(
		`{"name":"Mit eigener MwSt","link":null,"amount":-45000,"cycle":"monthly","type":"repeating","startDate":"2026-08-01","endDate":null,"category":1,"currency":1,"employee":null,"vat":%d,"VatIncluded":true}`,
		vatID))
	require.False(t, transaction.IsError, transaction.Text)
	transactionID := int64(transaction.Structured["id"].(float64))
	require.NotNil(t, transaction.Structured["vat"])

	deleted := env.mcpCall(t, token, "delete_vat", fmt.Sprintf(`{"id":%d}`, vatID))
	require.False(t, deleted.IsError, deleted.Text)

	after := env.mcpCall(t, token, "get_transaction", fmt.Sprintf(`{"id":%d}`, transactionID))
	require.False(t, after.IsError, after.Text)
	require.Nil(t, after.Structured["vat"], "deleting an assigned VAT must unlink it from the transaction")
	require.Equal(t, float64(0), after.Structured["vatAmount"])

	// Global VAT still present
	vats := env.mcpCall(t, token, "list_vats", `{}`)
	require.False(t, vats.IsError)
	items := vats.Structured["items"].([]any)
	require.Len(t, items, 1)
}

func TestMCPForecastReflectsTransactions(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	created := env.mcpCall(t, token, "create_transaction",
		`{"name":"Wiederkehrender Umsatz","link":null,"amount":100000,"cycle":"monthly","type":"repeating","startDate":"2026-08-01","endDate":null,"category":1,"currency":1,"employee":null,"vat":null,"VatIncluded":false}`)
	require.False(t, created.IsError, created.Text)

	forecast := env.mcpCall(t, token, "get_forecast", `{"months":3,"includeDetails":true}`)
	require.False(t, forecast.IsError, forecast.Text)

	months := forecast.Structured["months"].([]any)
	require.NotEmpty(t, months)

	var foundRevenue bool
	for _, monthEntry := range months {
		data := monthEntry.(map[string]any)["data"].(map[string]any)
		if data["revenue"].(float64) >= 100000 {
			foundRevenue = true
		}
	}
	require.True(t, foundRevenue, "monthly revenue transaction must appear in the forecast")
	require.NotNil(t, forecast.Structured["details"])
}

func TestMCPSalaryTimelineSemantics(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	employee := env.mcpCall(t, token, "create_employee", `{"name":"Timeline Mitarbeiter"}`)
	require.False(t, employee.IsError, employee.Text)
	employeeID := int64(employee.Structured["id"].(float64))

	createSalary := func(fromDate string, amount int64, isTermination bool) mcpToolResult {
		hours, vacation := 160, 25
		if isTermination {
			hours, vacation = 0, 0
		}
		return env.mcpCall(t, token, "create_salary", fmt.Sprintf(
			`{"employeeId":%d,"hoursPerMonth":%d,"amount":%d,"cycle":"monthly","currencyID":1,"vacationDaysPerYear":%d,"fromDate":%q,"toDate":null,"isTermination":%t}`,
			employeeID, hours, amount, vacation, fromDate, isTermination))
	}
	salaryByID := func(id int64) map[string]any {
		result := env.mcpCall(t, token, "get_salary", fmt.Sprintf(`{"id":%d}`, id))
		require.False(t, result.IsError, result.Text)
		return result.Structured
	}

	// January open-ended
	january := createSalary("2026-01-01", 800000, false)
	require.False(t, january.IsError, january.Text)
	januaryID := int64(january.Structured["id"].(float64))
	require.Nil(t, january.Structured["toDate"])

	// August insertion caps January automatically
	august := createSalary("2026-08-01", 900000, false)
	require.False(t, august.IsError, august.Text)
	augustID := int64(august.Structured["id"].(float64))
	require.Equal(t, "2026-07-01", salaryByID(januaryID)["toDate"])

	// April insertion re-caps January and gets capped itself by August
	april := createSalary("2026-04-01", 850000, false)
	require.False(t, april.IsError, april.Text)
	aprilID := int64(april.Structured["id"].(float64))
	require.Equal(t, "2026-07-01", april.Structured["toDate"])
	require.Equal(t, "2026-03-01", salaryByID(januaryID)["toDate"])

	// Deleting April re-expands January up to August
	deleted := env.mcpCall(t, token, "delete_salary", fmt.Sprintf(`{"id":%d}`, aprilID))
	require.False(t, deleted.IsError, deleted.Text)
	require.Equal(t, "2026-07-01", salaryByID(januaryID)["toDate"])

	// Termination caps the previous salary and flags the employee
	termination := createSalary("2026-10-01", 0, true)
	require.False(t, termination.IsError, termination.Text)
	terminationID := int64(termination.Structured["id"].(float64))
	require.Equal(t, "2026-09-01", salaryByID(augustID)["toDate"])

	employeeState := env.mcpCall(t, token, "get_employee", fmt.Sprintf(`{"id":%d}`, employeeID))
	require.False(t, employeeState.IsError, employeeState.Text)
	require.Equal(t, true, employeeState.Structured["willBeTerminated"])

	// Rehire after termination caps the termination entry (multiple exits supported)
	rehire := createSalary("2027-01-01", 950000, false)
	require.False(t, rehire.IsError, rehire.Text)
	rehireID := int64(rehire.Structured["id"].(float64))
	require.Equal(t, "2026-12-01", salaryByID(terminationID)["toDate"])

	secondExit := createSalary("2027-04-01", 0, true)
	require.False(t, secondExit.IsError, secondExit.Text)
	require.Equal(t, "2027-03-01", salaryByID(rehireID)["toDate"])

	// Same fromDate is rejected (unique per employee), for salaries and exits alike
	duplicate := createSalary("2026-08-01", 777000, false)
	require.True(t, duplicate.IsError)
	duplicateExit := createSalary("2027-04-01", 0, true)
	require.True(t, duplicateExit.IsError)

	// Deleting the employee cascades all salaries
	removed := env.mcpCall(t, token, "delete_employee", fmt.Sprintf(`{"id":%d}`, employeeID))
	require.False(t, removed.IsError, removed.Text)
	salaries := env.mcpCall(t, token, "list_salaries", fmt.Sprintf(`{"employeeId":%d}`, employeeID))
	require.False(t, salaries.IsError)
	require.Equal(t, float64(0), salaries.Structured["total"])
}

func TestMCPCategoryCRUDAndGlobalProtection(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	// Global preset (id 1 from fixtures) is protected
	globalUpdate := env.mcpCall(t, token, "update_category", `{"id":1,"name":"Umbenannt"}`)
	require.True(t, globalUpdate.IsError)
	require.Contains(t, globalUpdate.Text, "global")
	globalDelete := env.mcpCall(t, token, "delete_category", `{"id":1}`)
	require.True(t, globalDelete.IsError)
	require.Contains(t, globalDelete.Text, "global")

	// Own category lifecycle
	created := env.mcpCall(t, token, "create_category", `{"name":"Marketing"}`)
	require.False(t, created.IsError, created.Text)
	categoryID := int64(created.Structured["id"].(float64))
	require.Equal(t, true, created.Structured["canEdit"])

	renamed := env.mcpCall(t, token, "update_category", fmt.Sprintf(`{"id":%d,"name":"Marketing & Sales"}`, categoryID))
	require.False(t, renamed.IsError, renamed.Text)
	require.Equal(t, "Marketing & Sales", renamed.Structured["name"])

	listed := env.mcpCall(t, token, "list_categories", `{}`)
	require.False(t, listed.IsError)
	require.Equal(t, float64(2), listed.Structured["total"], "global preset + own category")

	// A category still used by a transaction must not be deletable (transactions
	// INNER JOIN categories, an orphaned category_id would hide the transaction)
	transaction := env.mcpCall(t, token, "create_transaction", fmt.Sprintf(
		`{"name":"Werbebudget","link":null,"amount":-50000,"cycle":"monthly","type":"repeating","startDate":"2026-08-01","endDate":null,"category":%d,"currency":1,"employee":null,"vat":null,"VatIncluded":false}`,
		categoryID))
	require.False(t, transaction.IsError, transaction.Text)
	transactionID := int64(transaction.Structured["id"].(float64))

	blocked := env.mcpCall(t, token, "delete_category", fmt.Sprintf(`{"id":%d}`, categoryID))
	require.True(t, blocked.IsError)
	require.Contains(t, blocked.Text, "used by")

	// Reassign the transaction to the global category, then deletion works
	reassigned := env.mcpCall(t, token, "update_transaction", allNullUpdateTransactionArgs(transactionID, map[string]string{
		"category": "1", "cycle": `"monthly"`, "isDisabled": "false",
	}))
	require.False(t, reassigned.IsError, reassigned.Text)

	deleted := env.mcpCall(t, token, "delete_category", fmt.Sprintf(`{"id":%d}`, categoryID))
	require.False(t, deleted.IsError, deleted.Text)

	after := env.mcpCall(t, token, "get_transaction", fmt.Sprintf(`{"id":%d}`, transactionID))
	require.False(t, after.IsError, after.Text)

	listedAfter := env.mcpCall(t, token, "list_categories", `{}`)
	require.Equal(t, float64(1), listedAfter.Structured["total"], "only the global preset remains")
}

func TestMCPForecastExclusionsAndSettings(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	created := env.mcpCall(t, token, "create_transaction",
		`{"name":"Ausschlusstest","link":null,"amount":-30000,"cycle":"monthly","type":"repeating","startDate":"2026-08-01","endDate":null,"category":1,"currency":1,"employee":null,"vat":null,"VatIncluded":false}`)
	require.False(t, created.IsError, created.Text)
	transactionID := int64(created.Structured["id"].(float64))

	// Exclude one month, verify via list + forecast details
	set := env.mcpCall(t, token, "set_forecast_exclusions", fmt.Sprintf(
		`{"updates":[{"relatedID":%d,"relatedTable":"transactions","month":"2026-09","isExcluded":true}]}`, transactionID))
	require.False(t, set.IsError, set.Text)
	require.Equal(t, float64(1), set.Structured["updated"])

	listed := env.mcpCall(t, token, "list_forecast_exclusions", `{}`)
	require.False(t, listed.IsError, listed.Text)
	require.Equal(t, float64(1), listed.Structured["total"])
	row := listed.Structured["items"].([]any)[0].(map[string]any)
	require.Equal(t, "2026-09", row["month"])
	require.Equal(t, "transactions", row["relatedTable"])
	require.Equal(t, float64(transactionID), row["relatedId"])
	require.Equal(t, "Ausschlusstest", row["name"])
	require.Equal(t, float64(-30000), row["amount"])

	// Re-include
	unset := env.mcpCall(t, token, "set_forecast_exclusions", fmt.Sprintf(
		`{"updates":[{"relatedID":%d,"relatedTable":"transactions","month":"2026-09","isExcluded":false}]}`, transactionID))
	require.False(t, unset.IsError, unset.Text)

	listed = env.mcpCall(t, token, "list_forecast_exclusions", `{}`)
	require.Equal(t, float64(0), listed.Structured["total"])

	// Forecast settings readable
	settings := env.mcpCall(t, token, "get_forecast_settings", `{}`)
	require.False(t, settings.IsError, settings.Text)
	require.NotNil(t, settings.Structured["forecastMonths"])
	require.NotNil(t, settings.Structured["forecastPerformance"])
}

func TestMCPSalaryCostsAndLabels(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	employee := env.mcpCall(t, token, "create_employee", `{"name":"Kosten Mitarbeiter"}`)
	require.False(t, employee.IsError, employee.Text)
	employeeID := int64(employee.Structured["id"].(float64))

	salary := env.mcpCall(t, token, "create_salary", fmt.Sprintf(
		`{"employeeId":%d,"hoursPerMonth":160,"amount":1000000,"cycle":"monthly","currencyID":1,"vacationDaysPerYear":25,"fromDate":"2026-01-01","toDate":null,"isTermination":false}`,
		employeeID))
	require.False(t, salary.IsError, salary.Text)
	salaryID := int64(salary.Structured["id"].(float64))

	// Label lifecycle
	label := env.mcpCall(t, token, "create_salary_cost_label", `{"name":"AHV/IV/EO"}`)
	require.False(t, label.IsError, label.Text)
	labelID := int64(label.Structured["id"].(float64))

	renamed := env.mcpCall(t, token, "update_salary_cost_label", fmt.Sprintf(`{"id":%d,"name":"AHV"}`, labelID))
	require.False(t, renamed.IsError, renamed.Text)
	require.Equal(t, "AHV", renamed.Structured["name"])

	labels := env.mcpCall(t, token, "list_salary_cost_labels", `{}`)
	require.False(t, labels.IsError)
	require.Equal(t, float64(1), labels.Structured["total"])

	// Percentage cost split between employer and employee: 5.3% AHV (unit: 5300 = 5.3%)
	cost := env.mcpCall(t, token, "create_salary_cost", fmt.Sprintf(
		`{"salaryId":%d,"cycle":"monthly","amountType":"percentage","amount":5300,"distributionType":"both","relativeOffset":1,"targetDate":null,"labelID":%d,"baseSalaryCostIDs":[]}`,
		salaryID, labelID))
	require.False(t, cost.IsError, cost.Text)
	costID := int64(cost.Structured["id"].(float64))

	// Salary reflects deductions and employer costs (5.3% of 10'000.00 = 530.00 each side)
	salaryAfter := env.mcpCall(t, token, "get_salary", fmt.Sprintf(`{"id":%d}`, salaryID))
	require.False(t, salaryAfter.IsError, salaryAfter.Text)
	require.Equal(t, float64(53000), salaryAfter.Structured["employeeDeductions"])
	require.Equal(t, float64(53000), salaryAfter.Structured["employerCosts"])

	// Update to employer-only fixed cost
	updated := env.mcpCall(t, token, "update_salary_cost", fmt.Sprintf(
		`{"id":%d,"cycle":"monthly","amountType":"fixed","amount":20000,"distributionType":"employer","relativeOffset":1,"targetDate":null,"labelID":%d,"baseSalaryCostIDs":[]}`,
		costID, labelID))
	require.False(t, updated.IsError, updated.Text)
	require.Equal(t, "fixed", updated.Structured["amountType"])

	salaryAfter = env.mcpCall(t, token, "get_salary", fmt.Sprintf(`{"id":%d}`, salaryID))
	require.Equal(t, float64(0), salaryAfter.Structured["employeeDeductions"])
	require.Equal(t, float64(20000), salaryAfter.Structured["employerCosts"])

	// Copy costs to a second salary
	secondSalary := env.mcpCall(t, token, "create_salary", fmt.Sprintf(
		`{"employeeId":%d,"hoursPerMonth":160,"amount":1100000,"cycle":"monthly","currencyID":1,"vacationDaysPerYear":25,"fromDate":"2027-01-01","toDate":null,"isTermination":false}`,
		employeeID))
	require.False(t, secondSalary.IsError, secondSalary.Text)
	secondSalaryID := int64(secondSalary.Structured["id"].(float64))

	copied := env.mcpCall(t, token, "copy_salary_costs", fmt.Sprintf(
		`{"salaryId":%d,"sourceSalaryID":%d,"ids":[]}`, secondSalaryID, salaryID))
	require.False(t, copied.IsError, copied.Text)
	require.Equal(t, float64(1), copied.Structured["total"])

	// Delete cost, salary costs empty again
	deleted := env.mcpCall(t, token, "delete_salary_cost", fmt.Sprintf(`{"id":%d}`, costID))
	require.False(t, deleted.IsError, deleted.Text)
	costs := env.mcpCall(t, token, "list_salary_costs", fmt.Sprintf(`{"salaryId":%d}`, salaryID))
	require.Equal(t, float64(0), costs.Structured["total"])

	// Label deletion
	labelDeleted := env.mcpCall(t, token, "delete_salary_cost_label", fmt.Sprintf(`{"id":%d}`, labelID))
	require.False(t, labelDeleted.IsError, labelDeleted.Text)
}

func TestMCPFiatRates(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	require.NoError(t, env.DBAdapter.UpsertFiatRate(models.CreateFiatRate{Base: "CHF", Target: "EUR", Rate: 1.08}))

	rates := env.mcpCall(t, token, "get_fiat_rates", `{"base":"chf"}`)
	require.False(t, rates.IsError, rates.Text)
	items := rates.Structured["items"].([]any)
	require.NotEmpty(t, items)
}

func TestMCPFriendlyErrors(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	notFoundCases := []struct {
		tool string
		args string
		want string
	}{
		{"get_transaction", `{"id":99999}`, "transaction not found"},
		{"get_employee", `{"id":99999}`, "employee not found"},
		{"get_salary", `{"id":99999}`, "salary not found"},
		{"delete_transaction", `{"id":99999}`, "transaction not found"},
	}
	for _, testCase := range notFoundCases {
		result := env.mcpCall(t, token, testCase.tool, testCase.args)
		require.True(t, result.IsError, testCase.tool)
		require.Equal(t, testCase.want, result.Text, testCase.tool)
	}

	// once-cycle targetDate pairing validated before the DB constraint
	employee := env.mcpCall(t, token, "create_employee", `{"name":"Fehler Mitarbeiter"}`)
	employeeID := int64(employee.Structured["id"].(float64))
	salary := env.mcpCall(t, token, "create_salary", fmt.Sprintf(
		`{"employeeId":%d,"hoursPerMonth":160,"amount":1000000,"cycle":"monthly","currencyID":1,"vacationDaysPerYear":25,"fromDate":"2026-01-01","toDate":null,"isTermination":false}`, employeeID))
	salaryID := int64(salary.Structured["id"].(float64))

	onceMissing := env.mcpCall(t, token, "create_salary_cost", fmt.Sprintf(
		`{"salaryId":%d,"cycle":"once","amountType":"fixed","amount":50000,"distributionType":"employer","relativeOffset":1,"targetDate":null,"labelID":null,"baseSalaryCostIDs":[]}`, salaryID))
	require.True(t, onceMissing.IsError)
	require.Contains(t, onceMissing.Text, "requires a targetDate")

	// Recurring cycles may carry a targetDate as recurrence anchor
	recurringWithDate := env.mcpCall(t, token, "create_salary_cost", fmt.Sprintf(
		`{"salaryId":%d,"cycle":"monthly","amountType":"fixed","amount":50000,"distributionType":"employer","relativeOffset":1,"targetDate":"2026-08-01","labelID":null,"baseSalaryCostIDs":[]}`, salaryID))
	require.False(t, recurringWithDate.IsError, recurringWithDate.Text)

	// unknown fiat base yields a clear error instead of an empty list
	rates := env.mcpCall(t, token, "get_fiat_rates", `{"base":"XXX"}`)
	require.True(t, rates.IsError)
	require.Contains(t, rates.Text, "no exchange rates")
}

func TestMCPListFilters(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	employee := env.mcpCall(t, token, "create_employee", `{"name":"Filter Mitarbeiter"}`)
	employeeID := int64(employee.Structured["id"].(float64))

	mkTx := func(name string, amount int64, employeeRef string) {
		result := env.mcpCall(t, token, "create_transaction", fmt.Sprintf(
			`{"name":%q,"link":null,"amount":%d,"cycle":"monthly","type":"repeating","startDate":"2026-08-01","endDate":null,"category":1,"currency":1,"employee":%s,"vat":null,"VatIncluded":false}`,
			name, amount, employeeRef))
		require.False(t, result.IsError, result.Text)
	}
	mkTx("Umsatz A", 100000, fmt.Sprintf("%d", employeeID))
	mkTx("Kosten B", -50000, "null")
	mkTx("Umsatz C", 200000, "null")

	byEmployee := env.mcpCall(t, token, "list_transactions", fmt.Sprintf(`{"employeeId":%d}`, employeeID))
	require.False(t, byEmployee.IsError, byEmployee.Text)
	require.Equal(t, float64(1), byEmployee.Structured["total"])

	expenses := env.mcpCall(t, token, "list_transactions", `{"direction":"expense"}`)
	require.Equal(t, float64(1), expenses.Structured["total"])

	revenue := env.mcpCall(t, token, "list_transactions", `{"direction":"revenue"}`)
	require.Equal(t, float64(2), revenue.Structured["total"])

	categories := env.mcpCall(t, token, "list_categories", `{"search":"gener"}`)
	require.False(t, categories.IsError, categories.Text)
	require.Equal(t, float64(1), categories.Structured["total"])

	currencies := env.mcpCall(t, token, "list_currencies", `{"search":"chf"}`)
	require.False(t, currencies.IsError, currencies.Text)
	require.Equal(t, float64(1), currencies.Structured["total"])
}

func TestMCPDuplicateSalary(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	employee := env.mcpCall(t, token, "create_employee", `{"name":"Duplikat Mitarbeiter"}`)
	require.False(t, employee.IsError, employee.Text)
	employeeID := int64(employee.Structured["id"].(float64))

	source := env.mcpCall(t, token, "create_salary", fmt.Sprintf(
		`{"employeeId":%d,"hoursPerMonth":160,"amount":800000,"cycle":"monthly","currencyID":1,"vacationDaysPerYear":25,"fromDate":"2026-01-01","toDate":null,"isTermination":false}`,
		employeeID))
	require.False(t, source.IsError, source.Text)
	sourceID := int64(source.Structured["id"].(float64))

	for _, cost := range []string{
		fmt.Sprintf(`{"salaryId":%d,"cycle":"monthly","amountType":"percentage","amount":5300,"distributionType":"both","relativeOffset":1,"targetDate":null,"labelID":null,"baseSalaryCostIDs":null}`, sourceID),
		fmt.Sprintf(`{"salaryId":%d,"cycle":"monthly","amountType":"fixed","amount":10000,"distributionType":"employer","relativeOffset":1,"targetDate":null,"labelID":null,"baseSalaryCostIDs":null}`, sourceID),
	} {
		created := env.mcpCall(t, token, "create_salary_cost", cost)
		require.False(t, created.IsError, created.Text)
	}

	duplicated := env.mcpCall(t, token, "duplicate_salary", fmt.Sprintf(
		`{"id":%d,"fromDate":"2027-01-01"}`, sourceID))
	require.False(t, duplicated.IsError, duplicated.Text)
	newID := int64(duplicated.Structured["id"].(float64))
	require.NotEqual(t, sourceID, newID)
	require.Equal(t, float64(800000), duplicated.Structured["amount"])
	require.Equal(t, "2027-01-01", duplicated.Structured["fromDate"])

	// Source got capped by the new period
	capped := env.mcpCall(t, token, "get_salary", fmt.Sprintf(`{"id":%d}`, sourceID))
	require.False(t, capped.IsError, capped.Text)
	require.Equal(t, "2026-12-01", capped.Structured["toDate"])

	// All cost entries were copied
	costs := env.mcpCall(t, token, "list_salary_costs", fmt.Sprintf(`{"salaryId":%d}`, newID))
	require.False(t, costs.IsError, costs.Text)
	require.Equal(t, float64(2), costs.Structured["total"])

	// Termination entries cannot be duplicated
	termination := env.mcpCall(t, token, "create_salary", fmt.Sprintf(
		`{"employeeId":%d,"hoursPerMonth":0,"amount":0,"cycle":"monthly","currencyID":1,"vacationDaysPerYear":0,"fromDate":"2027-06-01","toDate":null,"isTermination":true}`,
		employeeID))
	require.False(t, termination.IsError, termination.Text)
	terminationID := int64(termination.Structured["id"].(float64))
	rejected := env.mcpCall(t, token, "duplicate_salary", fmt.Sprintf(
		`{"id":%d,"fromDate":"2027-09-01"}`, terminationID))
	require.True(t, rejected.IsError)
}

func TestMCPInvitationsAndMembers(t *testing.T) {
	env := setupOAuthTestEnvironment(t)
	token := env.mcpAccessToken(t)

	// Create invitation
	created := env.mcpCall(t, token, "create_invitation", `{"email":"mcp-invite@example.com","role":"editor"}`)
	require.False(t, created.IsError, created.Text)
	invitationID := int64(created.Structured["id"].(float64))
	require.Equal(t, "editor", created.Structured["role"])

	// Invalid role rejected
	invalid := env.mcpCall(t, token, "create_invitation", `{"email":"x@example.com","role":"superadmin"}`)
	require.True(t, invalid.IsError)

	// List contains it
	list := env.mcpCall(t, token, "list_invitations", `{}`)
	require.False(t, list.IsError, list.Text)
	require.Equal(t, float64(1), list.Structured["total"])

	// Resend is rate-limited right after creation
	resent := env.mcpCall(t, token, "resend_invitation", fmt.Sprintf(`{"id":%d}`, invitationID))
	require.True(t, resent.IsError)

	// Delete it
	deleted := env.mcpCall(t, token, "delete_invitation", fmt.Sprintf(`{"id":%d}`, invitationID))
	require.False(t, deleted.IsError, deleted.Text)
	list = env.mcpCall(t, token, "list_invitations", `{}`)
	require.False(t, list.IsError)
	require.Equal(t, float64(0), list.Structured["total"])

	// Member removal: owner cannot remove themselves
	self := env.mcpCall(t, token, "remove_organisation_member", fmt.Sprintf(`{"userId":%d}`, env.User.ID))
	require.True(t, self.IsError)
}
