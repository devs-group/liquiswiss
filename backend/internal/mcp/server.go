package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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

// validateCostTargetDate enforces that once-cycle costs carry a targetDate
// before the DB check constraint surfaces as a raw SQL error. Recurring cycles
// may also set targetDate; it then acts as the recurrence anchor date.
func validateCostTargetDate(payload models.CreateSalaryCost) error {
	if payload.Cycle == "once" && (payload.TargetDate == nil || *payload.TargetDate == "") {
		return errors.New("cycle 'once' requires a targetDate (YYYY-MM-DD)")
	}
	return nil
}

// notFound maps raw sql/no-rows errors to a clear entity message, everything else passes through
func notFound(err error, entity string) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "no rows in result set") ||
		strings.Contains(err.Error(), "converting NULL to int64") {
		return fmt.Errorf("%s not found", entity)
	}
	return err
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
	Search string `json:"search,omitempty" jsonschema:"optional fuzzy search; results ranked by matchScore (1 = exact, typos tolerated)"`
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
		Name:        "update_organisation",
		Description: "Update the current organisation's name and/or base currency (partial: only provided fields change). Changing the currency affects how forecasts are calculated. Requires admin role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.UpdateOrganisation) (*sdk.CallToolResult, *models.Organisation, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		current, err := deps.apiService.GetCurrentOrganisation(userID)
		if err != nil {
			return nil, nil, err
		}
		if err := validate(in); err != nil {
			return nil, nil, err
		}
		org, err := deps.apiService.UpdateOrganisation(in, userID, current.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, org, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_organisations",
		Description: "List all organisations the user belongs to, with their role in each. Use switch_organisation to change the active one; all other tools operate on the currently active organisation.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		organisations, total, err := deps.apiService.ListOrganisations(userID, 1, 100)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": organisations, "total": total}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "switch_organisation",
		Description: "Switch the user's active organisation. Affects ALL subsequent tool calls (and the web UI session of the user). Use list_organisations for valid IDs.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		OrganisationID int64 `json:"organisationId" jsonschema:"organisation ID to switch to"`
	}) (*sdk.CallToolResult, *models.Organisation, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		err = deps.apiService.SetUserCurrentOrganisation(models.UpdateUserCurrentOrganisation{OrganisationID: in.OrganisationID}, userID)
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
		Name:        "list_organisation_members",
		Description: "List the members of the current organisation with user ID, name, email and role (owner, admin, editor, read-only).",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		org, err := deps.apiService.GetCurrentOrganisation(userID)
		if err != nil {
			return nil, nil, err
		}
		members, err := deps.apiService.ListOrganisationMembers(userID, org.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": members}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "remove_organisation_member",
		Description: "Remove a member from the current organisation. Their access ends immediately (open sessions are terminated); re-joining requires a new invitation. The last owner and yourself cannot be removed. Requires owner role.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		UserID int64 `json:"userId" jsonschema:"user ID of the member to remove (from list_organisation_members)"`
	}) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		org, err := deps.apiService.GetCurrentOrganisation(userID)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.RemoveOrganisationMember(userID, org.ID, in.UserID); err != nil {
			return nil, nil, err
		}
		return nil, &deleteOutput{Deleted: true, ID: in.UserID}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_invitations",
		Description: "List pending invitations of the current organisation (email, role, invited by, expiry). Requires admin role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		org, err := deps.apiService.GetCurrentOrganisation(userID)
		if err != nil {
			return nil, nil, err
		}
		invitations, err := deps.apiService.ListOrganisationInvitations(userID, org.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": invitations, "total": len(invitations)}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_invitation",
		Description: "Invite someone to the current organisation by email with a role (admin, editor, read-only). Sends an invitation mail; the invite expires after 7 days and can be resent. Requires admin role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.CreateInvitation) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		org, err := deps.apiService.GetCurrentOrganisation(userID)
		if err != nil {
			return nil, nil, err
		}
		if err := validate(in); err != nil {
			return nil, nil, err
		}
		invitation, err := deps.apiService.CreateOrganisationInvitation(in, userID, org.ID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(invitation)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_invitation",
		Description: "Revoke a pending invitation of the current organisation. Requires admin role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		org, err := deps.apiService.GetCurrentOrganisation(userID)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.DeleteOrganisationInvitation(userID, org.ID, in.ID); err != nil {
			return nil, nil, notFound(err, "invitation")
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "resend_invitation",
		Description: "Resend the invitation mail for a pending invitation of the current organisation. Rate-limited (anti-spam). Requires admin role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		org, err := deps.apiService.GetCurrentOrganisation(userID)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.ResendOrganisationInvitation(userID, org.ID, in.ID); err != nil {
			return nil, nil, notFound(err, "invitation")
		}
		return nil, map[string]any{"resent": true, "id": in.ID}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_vat_setting",
		Description: "Get the organisation's automatic VAT billing settings: enabled flag, billingDate (first billing), transactionMonthOffset (months between billing date and money movement) and interval (monthly, quarterly, biannually, yearly). Returns an error if not configured.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		setting, err := deps.apiService.GetVatSetting(userID)
		if err != nil {
			return nil, nil, errors.New("no VAT settings configured for this organisation")
		}
		return toMapResult(setting)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_vat_setting",
		Description: "Create or update the organisation's automatic VAT billing settings (partial: only provided fields change). Fields: enabled, billingDate (YYYY-MM-DD, first billing), transactionMonthOffset (0-12 months between billing and money movement), interval (monthly, quarterly, biannually, yearly). When no settings exist yet, enabled, billingDate and interval are required. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.UpdateVatSetting) (*sdk.CallToolResult, map[string]any, error) {
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
		existing, err := deps.apiService.GetVatSetting(userID)
		if err == nil && existing != nil {
			setting, err := deps.apiService.UpdateVatSetting(in, userID)
			if err != nil {
				return nil, nil, err
			}
			return toMapResult(setting)
		}
		// First-time setup requires the full payload
		if in.Enabled == nil || in.BillingDate == nil || in.Interval == nil {
			return nil, nil, errors.New("no VAT settings exist yet: enabled, billingDate and interval are required")
		}
		offset := 0
		if in.TransactionMonthOffset != nil {
			offset = *in.TransactionMonthOffset
		}
		setting, err := deps.apiService.CreateVatSetting(models.CreateVatSetting{
			Enabled:                *in.Enabled,
			BillingDate:            *in.BillingDate,
			TransactionMonthOffset: offset,
			Interval:               *in.Interval,
		}, userID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(setting)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_categories",
		Description: "List all transaction categories (needed as category ID when creating transactions). Optional search filters by name.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		Search string `json:"search,omitempty" jsonschema:"optional name filter, case-insensitive substring"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		categories, _, err := deps.apiService.ListCategories(userID, 1, 1000)
		if err != nil {
			return nil, nil, err
		}
		if in.Search != "" {
			ranked, err := fuzzyRank(categories, in.Search, func(category models.Category) string { return category.Name })
			if err != nil {
				return nil, nil, err
			}
			return nil, map[string]any{"items": ranked, "total": len(ranked)}, nil
		}
		return nil, map[string]any{"items": categories, "total": len(categories)}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_fiat_rates",
		Description: "Get the currency exchange rates used for conversions (base = the given currency code, e.g. CHF). The forecast converts foreign-currency amounts into the organisation's base currency with these rates. Rates sync from fixer.io twice a day.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		Base string `json:"base" jsonschema:"base currency code, e.g. CHF"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		if _, err := userIDFrom(ctx); err != nil {
			return nil, nil, err
		}
		base := strings.ToUpper(in.Base)
		rates, err := deps.apiService.ListFiatRates(base)
		if err != nil {
			return nil, nil, err
		}
		if len(rates) == 0 {
			return nil, nil, fmt.Errorf("no exchange rates found for base currency %q", base)
		}
		return nil, map[string]any{"items": rates}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_category",
		Description: "Create an organisation-owned transaction category. Global preset categories (canEdit=false) are shared; own categories (canEdit=true) can be renamed and deleted. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.CreateCategory) (*sdk.CallToolResult, *models.Category, error) {
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
		category, err := deps.apiService.CreateCategory(in, &userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, category, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_category",
		Description: "Rename an organisation-owned category (canEdit=true only; global presets cannot be changed). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID int64 `json:"id" jsonschema:"category ID"`
		models.UpdateCategory
	}) (*sdk.CallToolResult, *models.Category, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.UpdateCategory); err != nil {
			return nil, nil, err
		}
		existing, err := deps.apiService.GetCategory(userID, in.ID)
		if err != nil {
			return nil, nil, errors.New("category not found")
		}
		if !existing.CanEdit {
			return nil, nil, errors.New("this category is a global preset and cannot be modified")
		}
		category, err := deps.apiService.UpdateCategory(in.UpdateCategory, userID, in.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, category, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_category",
		Description: "Delete an organisation-owned category permanently (canEdit=true only; global presets cannot be deleted). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.DeleteCategory(userID, in.ID); err != nil {
			return nil, nil, err
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "reassign_category_transactions",
		Description: "Move ALL transactions of the current organisation from one category to another (e.g. to free up a category before delete_category). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		FromID int64 `json:"fromId" jsonschema:"source category id whose transactions get moved"`
		ToID   int64 `json:"toId" jsonschema:"target category id the transactions get assigned to"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		affected, err := deps.apiService.ReassignCategoryTransactions(userID, in.FromID, in.ToID)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"affected": affected}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_currencies",
		Description: "List all currencies (needed as currency ID when creating bank accounts or transactions). Optional search filters by code or description.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		Search string `json:"search,omitempty" jsonschema:"optional filter on code or description, case-insensitive"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		currencies, err := deps.apiService.ListCurrencies(userID)
		if err != nil {
			return nil, nil, err
		}
		if in.Search != "" {
			ranked, err := fuzzyRank(currencies, in.Search, func(currency models.Currency) string {
				code, description := "", ""
				if currency.Code != nil {
					code = *currency.Code
				}
				if currency.Description != nil {
					description = *currency.Description
				}
				return code + " " + description
			})
			if err != nil {
				return nil, nil, err
			}
			return nil, map[string]any{"items": ranked, "total": len(ranked)}, nil
		}
		return nil, map[string]any{"items": currencies, "total": len(currencies)}, nil
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
		backendSearch, page, limit := in.Search, in.Page, in.Limit
		if in.Search != "" {
			backendSearch, page, limit = "", 1, 10000
		}
		accounts, total, err := deps.apiService.ListBankAccounts(userID, page, limit, "name", "ASC", backendSearch)
		if err != nil {
			return nil, nil, err
		}
		if in.Search != "" {
			ranked, err := fuzzyRank(accounts, in.Search, func(account models.BankAccount) string { return account.Name })
			if err != nil {
				return nil, nil, err
			}
			if int64(len(ranked)) > in.Limit {
				ranked = ranked[:in.Limit]
			}
			return nil, map[string]any{"items": ranked, "total": len(ranked)}, nil
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
			return nil, nil, notFound(err, "bank account")
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})
}

// --- Transactions ---

type listTransactionsInput struct {
	listInput
	HideDisabled bool   `json:"hideDisabled,omitempty" jsonschema:"exclude disabled transactions"`
	HideExpired  bool   `json:"hideExpired,omitempty" jsonschema:"exclude transactions whose end date is in the past"`
	EmployeeID   int64  `json:"employeeId,omitempty" jsonschema:"only transactions linked to this employee"`
	CategoryID   int64  `json:"categoryId,omitempty" jsonschema:"only transactions in this category"`
	Type         string `json:"type,omitempty" jsonschema:"only 'single' or 'repeating' transactions"`
	Direction    string `json:"direction,omitempty" jsonschema:"'revenue' (positive amounts) or 'expense' (negative amounts)"`
}

func (l *listTransactionsInput) matches(transaction models.Transaction) bool {
	if l.EmployeeID != 0 && (transaction.Employee == nil || transaction.Employee.ID != l.EmployeeID) {
		return false
	}
	if l.CategoryID != 0 && transaction.Category.ID != l.CategoryID {
		return false
	}
	if l.Type != "" && transaction.Type != l.Type {
		return false
	}
	if l.Direction == "revenue" && transaction.Amount < 0 {
		return false
	}
	if l.Direction == "expense" && transaction.Amount >= 0 {
		return false
	}
	return true
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
		hasExtraFilters := in.EmployeeID != 0 || in.CategoryID != 0 || in.Type != "" || in.Direction != ""
		useFuzzy := in.Search != ""
		page, limit := in.Page, in.Limit
		if hasExtraFilters || useFuzzy {
			// Filters and fuzzy ranking are applied in this layer, fetch the full set
			page, limit = 1, 10000
		}
		backendSearch := ""
		transactions, total, err := deps.apiService.ListTransactions(userID, page, limit, "name", "ASC", backendSearch, in.HideDisabled, in.HideExpired)
		if err != nil {
			return nil, nil, err
		}
		if hasExtraFilters {
			filtered := make([]models.Transaction, 0, len(transactions))
			for _, transaction := range transactions {
				if in.matches(transaction) {
					filtered = append(filtered, transaction)
				}
			}
			transactions = filtered
			total = int64(len(filtered))
		}
		if useFuzzy {
			ranked, err := fuzzyRank(transactions, in.Search, func(transaction models.Transaction) string {
				name := transaction.Name
				if transaction.Employee != nil {
					name += " " + transaction.Employee.Name
				}
				return name + " " + transaction.Category.Name
			})
			if err != nil {
				return nil, nil, err
			}
			if int64(len(ranked)) > in.Limit {
				ranked = ranked[:in.Limit]
			}
			return nil, map[string]any{"items": ranked, "total": len(ranked)}, nil
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
			return nil, nil, notFound(err, "transaction")
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
		Description: "Update a transaction. WARNING, special partial-update semantics: when isDisabled is provided (true/false), nullable fields sent as null are PRESERVED; when isDisabled is omitted/null, nullable fields sent as null (link, cycle, endDate, employee, vat) are CLEARED. To safely change single fields without wiping others, always pass isDisabled (e.g. its current value). Requires editor role or higher.",
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
			return nil, nil, notFound(err, "transaction")
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
		backendSearch, page, limit := in.Search, in.Page, in.Limit
		if in.Search != "" {
			backendSearch, page, limit = "", 1, 10000
		}
		employees, total, err := deps.apiService.ListEmployees(userID, page, limit, "name", "ASC", backendSearch, in.HideTerminated)
		if err != nil {
			return nil, nil, err
		}
		if in.Search != "" {
			ranked, err := fuzzyRank(employees, in.Search, func(employee models.Employee) string { return employee.Name })
			if err != nil {
				return nil, nil, err
			}
			if int64(len(ranked)) > in.Limit {
				ranked = ranked[:in.Limit]
			}
			return nil, map[string]any{"items": ranked, "total": len(ranked)}, nil
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
			return nil, nil, notFound(err, "employee")
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
			return nil, nil, notFound(err, "employee")
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_salaries",
		Description: "List the salary history of one employee (amounts in Rappen/cents, with employer cost details). Salaries form a contiguous timeline of employment periods, ordered by fromDate: each entry is valid from its fromDate until its toDate (null = open-ended). Entries with isTermination=true mark an employment end at their fromDate.",
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

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_salary",
		Description: "Get a single salary entry by ID.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		salary, err := deps.apiService.GetSalary(userID, in.ID)
		if err != nil {
			return nil, nil, notFound(err, "salary")
		}
		return toMapResult(salary)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_salary",
		Description: "Create a salary entry for an employee. Amount is the gross salary per cycle in Rappen/cents (e.g. 1000000 = 10'000.00), cycle one of monthly, quarterly, biannually, yearly. Dates as YYYY-MM-DD. IMPORTANT concept: salaries form a contiguous employment timeline and the system auto-adjusts neighbours. Inserting a salary automatically caps the previous salary's toDate at one cycle before the new fromDate, and the new salary itself gets capped by the next existing salary. Leave toDate null; it is managed automatically. The latest salary stays open-ended. To model an employment end (Austritt), create an entry with isTermination=true, amount 0 and fromDate = end boundary; the employee then shows willBeTerminated/isTerminated. Multiple exits and re-entries are supported: a salary created after a termination models a rehire and caps the termination entry, enabling employment gaps. Only ONE entry per employee per fromDate (salary or termination); creating a second one on the same date fails. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		EmployeeID int64 `json:"employeeId" jsonschema:"employee ID"`
		models.CreateSalary
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.CreateSalary); err != nil {
			return nil, nil, err
		}
		salary, err := deps.apiService.CreateSalary(in.CreateSalary, userID, in.EmployeeID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(salary)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_salary",
		Description: "Update a salary entry (partial: only provided fields change). Note: changing fromDate re-triggers the automatic timeline shifts of neighbouring salaries (see create_salary). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID int64 `json:"id" jsonschema:"salary ID"`
		models.UpdateSalary
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.UpdateSalary); err != nil {
			return nil, nil, err
		}
		salary, err := deps.apiService.UpdateSalary(in.UpdateSalary, userID, in.ID)
		if err != nil {
			return nil, nil, err
		}
		return toMapResult(salary)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_salary_costs",
		Description: "List the employer/employee cost entries (Lohnnebenkosten like AHV, BVG, insurances) attached to one salary. Each cost has a label, cycle (once, monthly, quarterly, biannually, yearly), amountType (fixed = Rappen/cents, percentage in thousandths of a percent, 3 decimals supported: 5325 = 5.325% of the salary), distributionType (employee deduction, employer cost, or both) and calculated amounts/execution dates.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		SalaryID int64 `json:"salaryId" jsonschema:"salary ID"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		costs, total, err := deps.apiService.ListSalaryCosts(userID, in.SalaryID, 1, 1000, false)
		if err != nil {
			return nil, nil, err
		}
		items := make([]map[string]any, 0, len(costs))
		for _, cost := range costs {
			item, err := toMap(cost)
			if err != nil {
				return nil, nil, err
			}
			items = append(items, item)
		}
		return nil, map[string]any{"items": items, "total": total}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_salary_cost",
		Description: "Add a cost entry (Lohnnebenkosten) to a salary. amountType 'fixed' (amount in Rappen/cents) or 'percentage' in thousandths of a percent with 3 decimals supported (5325 = 5.325% of the salary). distributionType: 'employee' (deducted from gross), 'employer' (on top of gross) or 'both'. Cycle 'once' needs targetDate (YYYY-MM-DD); recurring cycles use relativeOffset (>=1, e.g. 1 = every cycle). Optionally labelID from list_salary_cost_labels and baseSalaryCostIDs to compute the percentage on top of other cost entries instead of the salary. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		SalaryID int64 `json:"salaryId" jsonschema:"salary ID"`
		models.CreateSalaryCost
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.CreateSalaryCost); err != nil {
			return nil, nil, err
		}
		if err := validateCostTargetDate(in.CreateSalaryCost); err != nil {
			return nil, nil, err
		}
		cost, err := deps.apiService.CreateSalaryCost(in.CreateSalaryCost, userID, in.SalaryID)
		if err != nil {
			return nil, nil, notFound(err, "salary")
		}
		return toMapResult(cost)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_salary_cost",
		Description: "Update a salary cost entry. Full replace: send ALL fields (cycle, amountType, amount, distributionType, relativeOffset, and targetDate/labelID/baseSalaryCostIDs as applicable), not a partial patch. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID int64 `json:"id" jsonschema:"salary cost ID"`
		models.CreateSalaryCost
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.CreateSalaryCost); err != nil {
			return nil, nil, err
		}
		if err := validateCostTargetDate(in.CreateSalaryCost); err != nil {
			return nil, nil, err
		}
		cost, err := deps.apiService.UpdateSalaryCost(in.CreateSalaryCost, userID, in.ID)
		if err != nil {
			return nil, nil, notFound(err, "salary cost")
		}
		return toMapResult(cost)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_salary_cost",
		Description: "Delete a salary cost entry permanently. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.DeleteSalaryCost(userID, in.ID); err != nil {
			return nil, nil, notFound(err, "salary cost")
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "copy_salary_costs",
		Description: "Copy cost entries from another salary onto the target salary. Provide sourceSalaryID (copies all its costs) or specific cost ids. Useful when a new salary period should keep the same Lohnnebenkosten. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		SalaryID int64 `json:"salaryId" jsonschema:"target salary ID"`
		models.CopySalaryCosts
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.CopySalaryCosts); err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.CopySalaryCosts(in.CopySalaryCosts, userID, in.SalaryID); err != nil {
			return nil, nil, err
		}
		costs, total, err := deps.apiService.ListSalaryCosts(userID, in.SalaryID, 1, 1000, false)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"copied": true, "total": total, "count": len(costs)}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "duplicate_salary",
		Description: "Duplicate an existing salary entry as a new period starting at fromDate, including ALL its cost entries (Lohnnebenkosten). Mirrors the frontend's salary copy function: same amount, cycle, currency, hours and vacation days; the timeline auto-adjusts neighbouring salaries (see create_salary). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID       int64  `json:"id" jsonschema:"source salary ID"`
		FromDate string `json:"fromDate" jsonschema:"start date of the new salary period (YYYY-MM-DD)"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		source, err := deps.apiService.GetSalary(userID, in.ID)
		if err != nil {
			return nil, nil, notFound(err, "salary")
		}
		if source.IsTermination {
			return nil, nil, fmt.Errorf("termination entries cannot be duplicated")
		}
		payload := models.CreateSalary{
			HoursPerMonth:       source.HoursPerMonth,
			Amount:              source.Amount,
			Cycle:               source.Cycle,
			CurrencyID:          *source.Currency.ID,
			VacationDaysPerYear: source.VacationDaysPerYear,
			FromDate:            in.FromDate,
		}
		if err := validate(payload); err != nil {
			return nil, nil, err
		}
		salary, err := deps.apiService.CreateSalary(payload, userID, source.EmployeeID)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.CopySalaryCosts(models.CopySalaryCosts{
			SourceSalaryID: &in.ID,
		}, userID, salary.ID); err != nil {
			return nil, nil, fmt.Errorf("salary created (id %d) but copying costs failed: %w", salary.ID, err)
		}
		return toMapResult(salary)
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_salary_cost_labels",
		Description: "List the organisation's salary cost labels (categories for Lohnnebenkosten, e.g. AHV/IV/EO, BVG, KTG). Use their IDs as labelID on salary cost entries. Optional search filters by name.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		Search string `json:"search,omitempty" jsonschema:"optional name filter, case-insensitive substring"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		labels, _, err := deps.apiService.ListSalaryCostLabels(userID, 1, 1000)
		if err != nil {
			return nil, nil, err
		}
		if in.Search != "" {
			ranked, err := fuzzyRank(labels, in.Search, func(label models.SalaryCostLabel) string { return label.Name })
			if err != nil {
				return nil, nil, err
			}
			return nil, map[string]any{"items": ranked, "total": len(ranked)}, nil
		}
		return nil, map[string]any{"items": labels, "total": len(labels)}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_salary_cost_label",
		Description: "Create a salary cost label (category for Lohnnebenkosten). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.CreateSalaryCostLabel) (*sdk.CallToolResult, *models.SalaryCostLabel, error) {
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
		label, err := deps.apiService.CreateSalaryCostLabel(in, userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, label, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_salary_cost_label",
		Description: "Rename a salary cost label. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID   int64  `json:"id" jsonschema:"label ID"`
		Name string `json:"name" jsonschema:"new label name"`
	}) (*sdk.CallToolResult, *models.SalaryCostLabel, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		payload := models.CreateSalaryCostLabel{Name: in.Name}
		if err := validate(payload); err != nil {
			return nil, nil, err
		}
		label, err := deps.apiService.UpdateSalaryCostLabel(payload, userID, in.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, label, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_salary_cost_label",
		Description: "Delete a salary cost label permanently. Cost entries using it keep working but lose the label. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.DeleteSalaryCostLabel(userID, in.ID); err != nil {
			return nil, nil, err
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_salary",
		Description: "Delete a salary entry permanently. The timeline auto-heals: the previous salary re-expands up to the next remaining salary (or open-ended). Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.DeleteSalary(userID, in.ID); err != nil {
			return nil, nil, notFound(err, "salary")
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
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

	sdk.AddTool(server, &sdk.Tool{
		Name:        "set_forecast_exclusions",
		Description: "Exclude or re-include specific entries from the forecast for specific months, without disabling them. Each update needs relatedID + relatedTable (from the get_forecast details tree, e.g. 'transactions' or 'salaries'), month as YYYY-MM and isExcluded. Accepts multiple updates at once. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.UpdateForecastExclusions) (*sdk.CallToolResult, map[string]any, error) {
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
		if err := deps.apiService.UpdateForecastExclusions(in, userID); err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"updated": len(in.Updates)}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_forecast_exclusions",
		Description: "List ALL active forecast exclusions of the current organisation: per row the month (YYYY-MM), the excluded entry's relatedTable/relatedId, its name and amount (Rappen/cents). Use these IDs with set_forecast_exclusions to re-include entries. Optional filters: month (YYYY-MM) and relatedTable.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		Month        string `json:"month,omitempty" jsonschema:"only exclusions for this month (YYYY-MM)"`
		RelatedTable string `json:"relatedTable,omitempty" jsonschema:"only exclusions of this table, e.g. transactions or salaries"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		exclusions, err := deps.apiService.ListAllForecastExclusions(userID)
		if err != nil {
			return nil, nil, err
		}
		if in.Month != "" || in.RelatedTable != "" {
			filtered := make([]models.ForecastExclusionInfo, 0, len(exclusions))
			for _, exclusion := range exclusions {
				if in.Month != "" && exclusion.Month != in.Month {
					continue
				}
				if in.RelatedTable != "" && exclusion.RelatedTable != in.RelatedTable {
					continue
				}
				filtered = append(filtered, exclusion)
			}
			exclusions = filtered
		}
		return nil, map[string]any{"items": exclusions, "total": len(exclusions)}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_forecast_settings",
		Description: "Get the user's forecast view settings for the current organisation: forecastMonths (how many months the web UI shows) and forecastPerformance (a 0-200% scaling the user applies to revenue in their web view; get_forecast returns UNSCALED numbers, so mention this when the user compares figures).",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		setting, err := deps.apiService.GetUserOrganisationSetting(userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{
			"forecastMonths":      setting.ForecastMonths,
			"forecastPerformance": setting.ForecastPerformance,
		}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_forecast_settings",
		Description: "Update the user's forecast view settings for the current organisation: forecastMonths (1-60) and/or forecastPerformance (0-200, percent scaling applied to revenue in the web forecast view). Only provided fields change.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ForecastMonths      *int `json:"forecastMonths,omitempty" jsonschema:"months shown in the web forecast, 1-60"`
		ForecastPerformance *int `json:"forecastPerformance,omitempty" jsonschema:"revenue performance scaling in percent, 0-200"`
	}) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		payload := models.UpdateUserOrganisationSetting{
			ForecastMonths:      in.ForecastMonths,
			ForecastPerformance: in.ForecastPerformance,
		}
		if err := validate(payload); err != nil {
			return nil, nil, err
		}
		setting, err := deps.apiService.UpdateUserOrganisationSetting(payload, userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{
			"forecastMonths":      setting.ForecastMonths,
			"forecastPerformance": setting.ForecastPerformance,
		}, nil
	})
}
