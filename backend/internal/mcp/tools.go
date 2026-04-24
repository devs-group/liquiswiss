package mcp

import (
	"encoding/json"
	"fmt"

	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/pkg/logger"
)

type toolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

type toolResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var tools = []toolDefinition{
	{
		Name:        "list_bank_accounts",
		Description: "List all bank accounts with their current balances and currencies for this organisation",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	},
	{
		Name:        "list_employees",
		Description: "List all employees in this organisation",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	},
	{
		Name:        "list_transactions",
		Description: "List all transactions (recurring and one-time) for this organisation with amounts and currencies",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	},
	{
		Name:        "get_organisation_info",
		Description: "Get organisation details including name, default currency, and member count",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	},
}

func handleToolsList() any {
	return map[string]any{
		"tools": tools,
	}
}

func handleToolsCall(params json.RawMessage, orgID int64, dbService db_adapter.IDatabaseAdapter) (any, *rpcError) {
	var call toolCallParams
	if err := json.Unmarshal(params, &call); err != nil {
		return nil, &rpcError{Code: -32602, Message: "invalid params"}
	}

	var result toolResult
	var err error

	switch call.Name {
	case "list_bank_accounts":
		result, err = toolListBankAccounts(orgID, dbService)
	case "list_employees":
		result, err = toolListEmployees(orgID, dbService)
	case "list_transactions":
		result, err = toolListTransactions(orgID, dbService)
	case "get_organisation_info":
		result, err = toolGetOrganisationInfo(orgID, dbService)
	default:
		return nil, &rpcError{Code: -32602, Message: fmt.Sprintf("unknown tool: %s", call.Name)}
	}

	if err != nil {
		logger.Logger.Errorf("MCP tool %s error for org %d: %v", call.Name, orgID, err)
		return toolResult{
			Content: []toolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil
	}

	return result, nil
}

func toolListBankAccounts(orgID int64, dbService db_adapter.IDatabaseAdapter) (toolResult, error) {
	accounts, err := dbService.ListBankAccountsByOrganisation(orgID)
	if err != nil {
		return toolResult{}, err
	}

	if len(accounts) == 0 {
		return toolResult{Content: []toolContent{{Type: "text", Text: "Keine Bankkonten vorhanden."}}}, nil
	}

	text := fmt.Sprintf("Bankkonten (%d):\n", len(accounts))
	for _, a := range accounts {
		code := ""
		if a.Currency.Code != nil {
			code = *a.Currency.Code
		}
		// Amount is in cents
		text += fmt.Sprintf("- %s: %s %.2f\n", a.Name, code, float64(a.Amount)/100)
	}

	return toolResult{Content: []toolContent{{Type: "text", Text: text}}}, nil
}

func toolListEmployees(orgID int64, dbService db_adapter.IDatabaseAdapter) (toolResult, error) {
	employees, err := dbService.ListEmployeesByOrganisation(orgID)
	if err != nil {
		return toolResult{}, err
	}

	if len(employees) == 0 {
		return toolResult{Content: []toolContent{{Type: "text", Text: "Keine Mitarbeiter vorhanden."}}}, nil
	}

	text := fmt.Sprintf("Mitarbeiter (%d):\n", len(employees))
	for _, e := range employees {
		text += fmt.Sprintf("- %s\n", e.Name)
	}

	return toolResult{Content: []toolContent{{Type: "text", Text: text}}}, nil
}

func toolListTransactions(orgID int64, dbService db_adapter.IDatabaseAdapter) (toolResult, error) {
	transactions, err := dbService.ListTransactionsByOrganisation(orgID)
	if err != nil {
		return toolResult{}, err
	}

	if len(transactions) == 0 {
		return toolResult{Content: []toolContent{{Type: "text", Text: "Keine Transaktionen vorhanden."}}}, nil
	}

	text := fmt.Sprintf("Transaktionen (%d):\n", len(transactions))
	for _, t := range transactions {
		code := ""
		if t.Currency.Code != nil {
			code = *t.Currency.Code
		}
		cycle := "einmalig"
		if t.Cycle != nil {
			cycle = *t.Cycle
		}
		status := ""
		if t.IsDisabled {
			status = " (deaktiviert)"
		}
		text += fmt.Sprintf("- %s: %s %.2f (%s, %s)%s\n", t.Name, code, float64(t.Amount)/100, t.Type, cycle, status)
	}

	return toolResult{Content: []toolContent{{Type: "text", Text: text}}}, nil
}

func toolGetOrganisationInfo(orgID int64, dbService db_adapter.IDatabaseAdapter) (toolResult, error) {
	org, err := dbService.GetOrganisationByID(orgID)
	if err != nil {
		return toolResult{}, err
	}

	text := fmt.Sprintf("Organisation: %s\n", org.Name)
	if org.Currency.Code != nil {
		text += fmt.Sprintf("Standardwährung: %s\n", *org.Currency.Code)
	}
	text += fmt.Sprintf("Mitglieder: %d\n", org.MemberCount)

	return toolResult{Content: []toolContent{{Type: "text", Text: text}}}, nil
}
