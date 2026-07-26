package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

type contextKey string

const userIDKey contextKey = "userID"

var errNotAuthenticated = errors.New("not authenticated")
var errInsufficientRole = errors.New("your role in this organisation does not allow modifications")

// GinHandler mounts the MCP server on a gin route. The OAuthBearerMiddleware
// must run before this handler so the user ID is present in the gin context.
func GinHandler(apiService api_service.IAPIService, dbService db_adapter.IDatabaseAdapter) gin.HandlerFunc {
	server := newServer(apiService, dbService)
	httpHandler := sdk.NewStreamableHTTPHandler(func(r *http.Request) *sdk.Server {
		return server
	}, &sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true})

	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		ctx := context.WithValue(c.Request.Context(), userIDKey, userID)
		httpHandler.ServeHTTP(c.Writer, c.Request.WithContext(ctx))
	}
}

func userIDFrom(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(userIDKey).(int64)
	if !ok || userID == 0 {
		return 0, errNotAuthenticated
	}
	return userID, nil
}

type toolDeps struct {
	apiService api_service.IAPIService
	dbService  db_adapter.IDatabaseAdapter
}

// requireEditor ensures the user's role in their current organisation allows mutations
func (d *toolDeps) requireEditor(userID int64) error {
	role, err := d.dbService.GetCurrentUserRole(userID)
	if err != nil {
		return err
	}
	switch role {
	case "owner", "admin", "editor":
		return nil
	default:
		return errInsufficientRole
	}
}

func validate(payload any) error {
	return utils.GetValidator().Struct(payload)
}

// toMapResult wraps toMap for direct use in tool returns
func toMapResult(v any) (*sdk.CallToolResult, map[string]any, error) {
	m, err := toMap(v)
	if err != nil {
		return nil, nil, err
	}
	return nil, m, nil
}

// toMap converts a model to a generic map so custom JSON types (e.g. types.AsDate,
// which marshals as a string) do not clash with the SDK's inferred output schema
func toMap(v any) (map[string]any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func newServer(apiService api_service.IAPIService, dbService db_adapter.IDatabaseAdapter) *sdk.Server {
	server := sdk.NewServer(&sdk.Implementation{Name: "liquiswiss", Version: "1.0.0"}, nil)
	deps := &toolDeps{apiService: apiService, dbService: dbService}

	registerOrganisationTools(server, deps)
	registerBankAccountTools(server, deps)
	registerTransactionTools(server, deps)
	registerEmployeeTools(server, deps)
	registerForecastTools(server, deps)

	return server
}

// --- Shared inputs/outputs ---

type emptyInput struct{}

type idInput struct {
	ID int64 `json:"id" jsonschema:"the ID of the entity"`
}

type listInput struct {
	Page   int64  `json:"page,omitempty" jsonschema:"page number, starts at 1 (default 1)"`
	Limit  int64  `json:"limit,omitempty" jsonschema:"items per page (default 50, max 200)"`
	Search string `json:"search,omitempty" jsonschema:"optional search term"`
}

func (l *listInput) normalize() {
	if l.Page < 1 {
		l.Page = 1
	}
	if l.Limit < 1 {
		l.Limit = 50
	}
	if l.Limit > 200 {
		l.Limit = 200
	}
}

type deleteOutput struct {
	Deleted bool  `json:"deleted"`
	ID      int64 `json:"id"`
}

// --- Organisation ---

func registerOrganisationTools(server *sdk.Server, deps *toolDeps) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_organisation",
		Description: "Get the user's current organisation including its base currency. Amounts across all tools are integers in the smallest currency unit (Rappen/cents), i.e. 150000 = 1500.00.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, *models.Organisation, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		org, err := deps.apiService.GetCurrentOrganisation(userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, org, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_categories",
		Description: "List all transaction categories (needed as category ID when creating transactions).",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		categories, total, err := deps.apiService.ListCategories(userID, 1, 1000)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": categories, "total": total}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_currencies",
		Description: "List all currencies (needed as currency ID when creating bank accounts or transactions).",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		currencies, err := deps.apiService.ListCurrencies(userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": currencies}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_vats",
		Description: "List the organisation's VAT rates (usable as vat ID on transactions).",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		vats, err := deps.apiService.ListVats(userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": vats}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_vat",
		Description: "Create a VAT rate for the organisation. Value in basis points of a percent: 810 = 8.1%. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.CreateVat) (*sdk.CallToolResult, *models.Vat, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in); err != nil {
			return nil, nil, err
		}
		vat, err := deps.apiService.CreateVat(in, userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, vat, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_vat",
		Description: "Update a VAT rate's value (810 = 8.1%). Only organisation-owned rates can be edited (canEdit=true). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID int64 `json:"id" jsonschema:"VAT ID"`
		models.UpdateVat
	}) (*sdk.CallToolResult, *models.Vat, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.UpdateVat); err != nil {
			return nil, nil, err
		}
		existing, err := deps.apiService.GetVat(userID, in.ID)
		if err != nil {
			return nil, nil, errors.New("vat not found")
		}
		if !existing.CanEdit {
			return nil, nil, errors.New("this VAT rate is global and cannot be modified")
		}
		vat, err := deps.apiService.UpdateVat(in.UpdateVat, userID, in.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, vat, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_vat",
		Description: "Delete an organisation-owned VAT rate permanently (canEdit=true only). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		existing, err := deps.apiService.GetVat(userID, in.ID)
		if err != nil {
			return nil, nil, errors.New("vat not found")
		}
		if !existing.CanEdit {
			return nil, nil, errors.New("this VAT rate is global and cannot be deleted")
		}
		if err := deps.apiService.DeleteVat(userID, in.ID); err != nil {
			return nil, nil, err
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})
}

// --- Bank accounts ---

func registerBankAccountTools(server *sdk.Server, deps *toolDeps) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_bank_accounts",
		Description: "List all bank accounts of the current organisation with their balances (amounts in Rappen/cents).",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in listInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		in.normalize()
		accounts, total, err := deps.apiService.ListBankAccounts(userID, in.Page, in.Limit, "name", "ASC", in.Search)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": accounts, "total": total}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_bank_account",
		Description: "Create a bank account. Amount is the current balance in Rappen/cents. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.CreateBankAccount) (*sdk.CallToolResult, *models.BankAccount, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in); err != nil {
			return nil, nil, err
		}
		account, err := deps.apiService.CreateBankAccount(in, userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, account, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_bank_account",
		Description: "Update a bank account (partial: only provided fields change). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID int64 `json:"id" jsonschema:"bank account ID"`
		models.UpdateBankAccount
	}) (*sdk.CallToolResult, *models.BankAccount, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.UpdateBankAccount); err != nil {
			return nil, nil, err
		}
		account, err := deps.apiService.UpdateBankAccount(in.UpdateBankAccount, userID, in.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, account, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_bank_account",
		Description: "Delete a bank account permanently. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.DeleteBankAccount(userID, in.ID); err != nil {
			return nil, nil, err
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})
}

// --- Transactions ---

type listTransactionsInput struct {
	listInput
	HideDisabled bool `json:"hideDisabled,omitempty" jsonschema:"exclude disabled transactions"`
	HideExpired  bool `json:"hideExpired,omitempty" jsonschema:"exclude transactions whose end date is in the past"`
}

func registerTransactionTools(server *sdk.Server, deps *toolDeps) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_transactions",
		Description: "List revenue and expense transactions. Amounts in Rappen/cents; type is 'single' or 'repeating' with a cycle (monthly, quarterly, biannually, yearly). Negative amounts are expenses, positive are revenue.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in listTransactionsInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		in.normalize()
		transactions, total, err := deps.apiService.ListTransactions(userID, in.Page, in.Limit, "name", "ASC", in.Search, in.HideDisabled, in.HideExpired)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": transactions, "total": total}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_transaction",
		Description: "Get a single transaction by ID.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		transaction, err := deps.apiService.GetTransaction(userID, in.ID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(transaction)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_transaction",
		Description: "Create a transaction. Amount in Rappen/cents (negative = expense, positive = revenue). Type 'single' or 'repeating' (cycle required if repeating: monthly, quarterly, biannually, yearly). Dates as YYYY-MM-DD. Category and currency are IDs from list_categories / list_currencies. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.CreateTransaction) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in); err != nil {
			return nil, nil, err
		}
		transaction, err := deps.apiService.CreateTransaction(in, userID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(transaction)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_transaction",
		Description: "Update a transaction (partial: only provided fields change). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID int64 `json:"id" jsonschema:"transaction ID"`
		models.UpdateTransaction
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.UpdateTransaction); err != nil {
			return nil, nil, err
		}
		transaction, err := deps.apiService.UpdateTransaction(in.UpdateTransaction, userID, in.ID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(transaction)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_transaction",
		Description: "Delete a transaction permanently. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.DeleteTransaction(userID, in.ID); err != nil {
			return nil, nil, err
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})
}

// --- Employees & salaries ---

type listEmployeesInput struct {
	listInput
	HideTerminated bool `json:"hideTerminated,omitempty" jsonschema:"exclude employees whose employment ended"`
}

func registerEmployeeTools(server *sdk.Server, deps *toolDeps) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_employees",
		Description: "List employees with their current salary summary (salary amounts in Rappen/cents).",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in listEmployeesInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		in.normalize()
		employees, total, err := deps.apiService.ListEmployees(userID, in.Page, in.Limit, "name", "ASC", in.Search, in.HideTerminated)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": employees, "total": total}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_employee",
		Description: "Get a single employee by ID.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		employee, err := deps.apiService.GetEmployee(userID, in.ID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(employee)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_employee",
		Description: "Create an employee (name only; salaries are managed separately). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.CreateEmployee) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in); err != nil {
			return nil, nil, err
		}
		employee, err := deps.apiService.CreateEmployee(in, userID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(employee)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_employee",
		Description: "Rename an employee. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID int64 `json:"id" jsonschema:"employee ID"`
		models.UpdateEmployee
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.UpdateEmployee); err != nil {
			return nil, nil, err
		}
		employee, err := deps.apiService.UpdateEmployee(in.UpdateEmployee, userID, in.ID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(employee)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_employee",
		Description: "Delete an employee and their salaries permanently. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.DeleteEmployee(userID, in.ID); err != nil {
			return nil, nil, err
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_salaries",
		Description: "List the salary history of one employee (amounts in Rappen/cents, with employer cost details).",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		EmployeeID int64 `json:"employeeId" jsonschema:"employee ID"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		salaries, total, err := deps.apiService.ListSalaries(userID, in.EmployeeID, 1, 1000)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": salaries, "total": total}, nil
	})
}

// --- Forecast ---

type forecastInput struct {
	Months         int64 `json:"months,omitempty" jsonschema:"how many months ahead to forecast, 1-36 (default 12). The forecast engine works in whole months"`
	IncludeDetails bool  `json:"includeDetails,omitempty" jsonschema:"include the per-month revenue/expense breakdown tree (which transactions, employees etc. contribute)"`
}

func registerForecastTools(server *sdk.Server, deps *toolDeps) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_forecast",
		Description: "Recalculate and return the liquidity forecast: per month revenue, expense and cashflow (in Rappen/cents) for the current organisation. Use includeDetails to see exactly which transactions and salaries drive each month, ideal for spotting outdated entries or saving potential.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in forecastInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		months := in.Months
		if months < 1 {
			months = 12
		}
		if months > int64(utils.MaxForecastYears)*12 {
			months = int64(utils.MaxForecastYears) * 12
		}

		if _, err := deps.apiService.CalculateForecast(userID); err != nil {
			return nil, nil, fmt.Errorf("forecast calculation failed: %w", err)
		}
		forecasts, err := deps.apiService.ListForecasts(userID, months)
		if err != nil {
			return nil, nil, err
		}

		result := map[string]any{"months": forecasts}
		if in.IncludeDetails {
			details, err := deps.apiService.ListForecastDetails(userID, months)
			if err != nil {
				return nil, nil, err
			}
			result["details"] = details
		}
		return nil, result, nil
	})
}
