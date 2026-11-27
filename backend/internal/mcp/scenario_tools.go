package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"liquiswiss/pkg/models"
)

func registerScenarioTools(server *sdk.Server, deps *toolDeps) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_scenarios",
		Description: "List the scenarios of the current organisation. Every organisation has one base scenario (isBase); other scenarios are independent copies with their own employees, salaries, transactions, bank accounts, exclusions and forecast. Categories, VAT rates and salary cost labels are shared. isCurrent marks the scenario all other tools currently read and write.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in emptyInput) (*sdk.CallToolResult, map[string]any, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		scenarios, err := deps.apiService.ListScenarios(ctx, userID)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": scenarios, "total": len(scenarios)}, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "create_scenario",
		Description: "Create a scenario as a full copy of sourceScenarioId (default: the current scenario) and switch to it. Later changes to the source do not reach the copy. Affects ALL subsequent tool calls and the user's web UI. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in models.CreateScenario) (*sdk.CallToolResult, *models.Scenario, error) {
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
		scenario, err := deps.apiService.CreateScenario(ctx, in, userID)
		if err != nil {
			return nil, nil, notFound(err, "scenario")
		}
		return nil, scenario, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "switch_scenario",
		Description: "Switch the user's active scenario within the current organisation. Affects ALL subsequent tool calls and the user's web UI, other members keep their own choice. Use list_scenarios for valid IDs.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ScenarioID int64 `json:"scenarioId" jsonschema:"scenario ID to switch to"`
	}) (*sdk.CallToolResult, *models.Scenario, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		scenario, err := deps.apiService.SwitchScenario(ctx, userID, in.ScenarioID)
		if err != nil {
			return nil, nil, notFound(err, "scenario")
		}
		return nil, scenario, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "update_scenario",
		Description: "Rename a scenario. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in struct {
		ID int64 `json:"id" jsonschema:"scenario ID"`
		models.UpdateScenario
	}) (*sdk.CallToolResult, *models.Scenario, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := validate(in.UpdateScenario); err != nil {
			return nil, nil, err
		}
		scenario, err := deps.apiService.UpdateScenario(ctx, in.UpdateScenario, userID, in.ID)
		if err != nil {
			return nil, nil, notFound(err, "scenario")
		}
		return nil, scenario, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "delete_scenario",
		Description: "Delete a scenario and all of its data permanently. The base scenario cannot be deleted. Users who had it active fall back to the base scenario. Requires editor role or higher.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, *deleteOutput, error) {
		userID, err := userIDFrom(ctx)
		if err != nil {
			return nil, nil, err
		}
		if err := deps.requireEditor(userID); err != nil {
			return nil, nil, err
		}
		if err := deps.apiService.DeleteScenario(ctx, userID, in.ID); err != nil {
			return nil, nil, notFound(err, "scenario")
		}
		return nil, &deleteOutput{Deleted: true, ID: in.ID}, nil
	})
}
