package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"liquiswiss/internal/adapter/aigent_adapter"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/logger"
)

var validPages = map[string]bool{
	"forecast":      true,
	"employees":     true,
	"transactions":  true,
	"bank-accounts": true,
	"settings":      true,
}

type chatRequest struct {
	Page      string `json:"page" validate:"required"`
	Message   string `json:"message" validate:"required"`
	SessionID string `json:"session_id"`
}

type chatResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

func SendChatMessage(apiService api_service.IAPIService, aigentService aigent_adapter.IAigentAdapter, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer"})
		return
	}

	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Anfrage"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nachricht darf nicht leer sein"})
		return
	}

	if !validPages[req.Page] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Seite"})
		return
	}

	chatID, err := apiService.GetOrganisationChatbot(userID)
	if err != nil || chatID == "" {
		logger.Logger.Errorf("GetOrganisationChatbot failed for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Chat-Konfiguration fehlt"})
		return
	}

	context := gatherPageContext(apiService, userID, req.Page)
	enrichedMessage := req.Message
	if context != "" {
		enrichedMessage = fmt.Sprintf("Kontext:\n%s\n\nFrage: %s", context, req.Message)
	}

	resp, err := aigentService.SendMessage(chatID, enrichedMessage, req.SessionID)
	if err != nil {
		logger.Logger.Errorf("SendMessage failed for chatID %s: %v", chatID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler bei der Kommunikation mit dem Assistenten"})
		return
	}

	c.JSON(http.StatusOK, chatResponse{
		Message:   resp.Response,
		SessionID: resp.SessionID,
	})
}

func gatherPageContext(apiService api_service.IAPIService, userID int64, page string) string {
	switch page {
	case "forecast":
		return gatherForecastContext(apiService, userID)
	case "employees":
		return gatherEmployeesContext(apiService, userID)
	case "transactions":
		return gatherTransactionsContext(apiService, userID)
	case "bank-accounts":
		return gatherBankAccountsContext(apiService, userID)
	case "settings":
		return gatherSettingsContext(apiService, userID)
	default:
		return ""
	}
}

func gatherForecastContext(apiService api_service.IAPIService, userID int64) string {
	bankAccounts, _, err := apiService.ListBankAccounts(userID, 1, 100, "name", "ASC", "")
	if err != nil {
		return ""
	}
	result := fmt.Sprintf("Bankkonten: %d", len(bankAccounts))
	for _, ba := range bankAccounts {
		result += fmt.Sprintf("\n- %s: %s %d", ba.Name, *ba.Currency.Code, ba.Amount)
	}
	return result
}

func gatherEmployeesContext(apiService api_service.IAPIService, userID int64) string {
	employees, _, err := apiService.ListEmployees(userID, 1, 100, "name", "ASC", "", false)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("Mitarbeiter: %d", len(employees))
}

func gatherTransactionsContext(apiService api_service.IAPIService, userID int64) string {
	transactions, _, err := apiService.ListTransactions(userID, 1, 100, "name", "ASC", "", false)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("Transaktionen: %d", len(transactions))
}

func gatherBankAccountsContext(apiService api_service.IAPIService, userID int64) string {
	bankAccounts, _, err := apiService.ListBankAccounts(userID, 1, 100, "name", "ASC", "")
	if err != nil {
		return ""
	}
	result := fmt.Sprintf("Bankkonten: %d", len(bankAccounts))
	for _, ba := range bankAccounts {
		result += fmt.Sprintf("\n- %s: %s %d", ba.Name, *ba.Currency.Code, ba.Amount)
	}
	return result
}

func gatherSettingsContext(apiService api_service.IAPIService, userID int64) string {
	orgs, _, err := apiService.ListOrganisations(userID, 1, 100)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("Organisationen: %d", len(orgs))
}
