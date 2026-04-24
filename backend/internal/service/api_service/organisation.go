package api_service

import (
	"errors"
	"fmt"
	"liquiswiss/config"
	"liquiswiss/internal/adapter/aigent_adapter"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"slices"
)

func (a *APIService) ListOrganisations(userID int64, page int64, limit int64) ([]models.Organisation, int64, error) {
	organisations, totalCount, err := a.dbService.ListOrganisations(userID, page, limit)
	if err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	validator := utils.GetValidator()
	if err := validator.Var(organisations, "dive"); err != nil {
		logger.Logger.Error(err)
		return nil, 0, err
	}
	return organisations, totalCount, nil
}

func (a *APIService) GetOrganisation(userID int64, organisationID int64) (*models.Organisation, error) {
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(organisation); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return organisation, nil
}

func (a *APIService) CreateOrganisation(payload models.CreateOrganisation, userID int64) (*models.Organisation, error) {
	organisationID, err := a.dbService.CreateOrganisation(payload.Name)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.AssignUserToOrganisation(userID, organisationID, "owner", false)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Provision chatbot for the new organisation
	a.provisionChatbot(organisationID, payload.Name)

	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(organisation); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return organisation, nil
}

func (a *APIService) provisionChatbot(organisationID int64, organisationName string) {
	if a.aigentAdapter == nil {
		return
	}

	// 1. Create chatbot
	chatbotReq := models.AigentCreateChatbotRequest{
		Name:               fmt.Sprintf("%d - %s", organisationID, organisationName),
		Description:        fmt.Sprintf("Chatbot for organisation %s", organisationName),
		SystemInstructions: aigent_adapter.SystemPrompt,
		ModelName:          "gemini-flash",
		MaxTokens:          4096,
		TemperatureParam:   1.0,
		AgentEnabled:       true,
		MaxAgentTurns:      15,
		MaxToolCalls:       20,
	}

	chatbotResp, err := a.aigentAdapter.CreateChatbot(chatbotReq)
	if err != nil {
		logger.Logger.Errorf("Failed to create chatbot for organisation %d: %v", organisationID, err)
		return
	}

	// 2. Create skill + MCP server and assign to chatbot
	skillID := a.provisionMCPSkill(organisationID, chatbotResp.ID)

	// 3. Store in DB
	err = a.dbService.CreateOrganisationChatbot(organisationID, chatbotResp.ID, skillID)
	if err != nil {
		logger.Logger.Errorf("Failed to store chatbot ID for organisation %d: %v", organisationID, err)
	}
}

func (a *APIService) provisionMCPSkill(organisationID int64, chatbotID string) *string {
	cfg := config.GetConfig()
	if cfg.BackendPublicURL == "" {
		logger.Logger.Warn("BACKEND_PUBLIC_URL not set, skipping MCP skill provisioning")
		return nil
	}

	logger.Logger.Infof("Provisioning MCP skill for org %d, chatbot %s", organisationID, chatbotID)

	// Create a skill
	skillResp, err := a.aigentAdapter.CreateSkill(models.AigentCreateSkillRequest{
		Title:       "LiquiSwiss Data Access",
		Description: "Access organisation data like bank accounts, employees, and transactions",
		PromptMD:    "Use the available MCP tools to look up organisation data when users ask about their finances, employees, or transactions.",
	})
	if err != nil {
		logger.Logger.Errorf("Failed to create skill for org %d: %v", organisationID, err)
		return nil
	}
	logger.Logger.Infof("Created skill %s for org %d", skillResp.ID, organisationID)

	// Generate MCP JWT token
	mcpToken, err := auth.GenerateMCPToken(organisationID, chatbotID)
	if err != nil {
		logger.Logger.Errorf("Failed to generate MCP token for org %d: %v", organisationID, err)
		return &skillResp.ID
	}

	// Attach MCP server to skill
	mcpURL := cfg.BackendPublicURL + "/api/mcp"
	logger.Logger.Infof("Creating MCP server for skill %s, URL: %s", skillResp.ID, mcpURL)
	_, err = a.aigentAdapter.CreateMCPServer(skillResp.ID, models.AigentCreateMCPServerRequest{
		Alias: "liquiswiss",
		URL:   mcpURL,
		Headers: map[string]string{
			"Authorization": "Bearer " + mcpToken,
		},
	})
	if err != nil {
		logger.Logger.Errorf("Failed to create MCP server for org %d: %v", organisationID, err)
		return &skillResp.ID
	}
	logger.Logger.Infof("Created MCP server for skill %s", skillResp.ID)

	// Assign skill to chatbot
	err = a.aigentAdapter.UpdateChatbot(chatbotID, models.AigentUpdateChatbotRequest{
		SkillIDs: []string{skillResp.ID},
	})
	if err != nil {
		logger.Logger.Errorf("Failed to assign skill to chatbot for org %d: %v", organisationID, err)
	} else {
		logger.Logger.Infof("Assigned skill %s to chatbot %s", skillResp.ID, chatbotID)
	}

	return &skillResp.ID
}

func (a *APIService) GetOrganisationChatbot(userID int64) (string, error) {
	user, err := a.dbService.GetProfile(userID)
	if err != nil {
		return "", err
	}
	return a.dbService.GetOrganisationChatbot(user.CurrentOrganisationID)
}

func (a *APIService) HasOrganisationChatbots(userID int64) (bool, error) {
	user, err := a.dbService.GetProfile(userID)
	if err != nil {
		return false, err
	}
	return a.dbService.HasOrganisationChatbots(user.CurrentOrganisationID)
}

func (a *APIService) ProvisionOrganisationChatbots(userID int64, organisationID int64) error {
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		return err
	}
	if !a.hasEditingPermission(organisation.Role) {
		return errors.New("Permission denied")
	}

	// Check if already provisioned
	has, err := a.dbService.HasOrganisationChatbots(organisationID)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	a.provisionChatbot(organisationID, organisation.Name)
	return nil
}

func (a *APIService) DeleteOrganisationChatbots(userID int64, organisationID int64) error {
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		return err
	}
	if !a.hasEditingPermission(organisation.Role) {
		return errors.New("Permission denied")
	}

	// Get all chatbot IDs to delete from AIgent
	chatbots, err := a.dbService.ListOrganisationChatbots(organisationID)
	if err != nil {
		logger.Logger.Errorf("Failed to list chatbots for org %d: %v", organisationID, err)
		return err
	}

	logger.Logger.Infof("Deleting %d chatbots for org %d", len(chatbots), organisationID)

	// Delete each chatbot and skill from AIgent API
	if a.aigentAdapter != nil {
		for _, chatbot := range chatbots {
			if chatbot.SkillID != nil {
				if err := a.aigentAdapter.DeleteSkill(*chatbot.SkillID); err != nil {
					logger.Logger.Errorf("Failed to delete skill %s from AIgent: %v", *chatbot.SkillID, err)
				}
			}
			if err := a.aigentAdapter.DeleteChatbot(chatbot.ChatbotID); err != nil {
				logger.Logger.Errorf("Failed to delete chatbot %s from AIgent: %v", chatbot.ChatbotID, err)
			}
		}
	}

	// Delete from local DB
	return a.dbService.DeleteOrganisationChatbots(organisationID)
}

func (a *APIService) UpdateOrganisation(payload models.UpdateOrganisation, userID int64, organisationID int64) (*models.Organisation, error) {
	existingOrganisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	// Check if user is allowed to edit
	if !a.hasEditingPermission(existingOrganisation.Role) {
		err = errors.New("Permission denied")
		logger.Logger.Error(err)
		return nil, err
	}
	err = a.dbService.UpdateOrganisation(payload, userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	organisation, err := a.dbService.GetOrganisation(userID, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	validator := utils.GetValidator()
	if err := validator.Struct(organisation); err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	return organisation, err
}

func (a *APIService) hasEditingPermission(role string) bool {
	editingRoles := []string{"owner", "admin"}
	return slices.Contains(editingRoles, role)
}
