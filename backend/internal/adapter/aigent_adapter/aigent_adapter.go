//go:generate mockgen -package=mocks -destination ../../mocks/aigent_adapter.go liquiswiss/internal/adapter/aigent_adapter IAigentAdapter
package aigent_adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"liquiswiss/config"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"net/http"
	"sync"
	"time"
)

type IAigentAdapter interface {
	SendMessage(chatID, query, sessionID string) (*models.AigentChatResponse, error)
	CreateChatbot(req models.AigentCreateChatbotRequest) (*models.AigentCreateChatbotResponse, error)
	DeleteChatbot(chatID string) error
	CreateSkill(req models.AigentCreateSkillRequest) (*models.AigentCreateSkillResponse, error)
	DeleteSkill(skillID string) error
	CreateMCPServer(skillID string, req models.AigentCreateMCPServerRequest) (*models.AigentMCPServerResponse, error)
	UpdateChatbot(chatbotID string, req models.AigentUpdateChatbotRequest) error
}

type AigentAdapter struct {
	apiURL       string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func NewAigentAdapter(cfg config.Config) IAigentAdapter {
	return &AigentAdapter{
		apiURL:       cfg.AigentAPIURL,
		clientID:     cfg.AigentClientID,
		clientSecret: cfg.AigentClientSecret,
		httpClient:   &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *AigentAdapter) getToken() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return a.accessToken, nil
	}

	return a.fetchToken()
}

func (a *AigentAdapter) fetchToken() (string, error) {
	tokenReq := models.AigentTokenRequest{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
	}

	body, err := json.Marshal(tokenReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	resp, err := a.httpClient.Post(
		a.apiURL+"/public/oauth/token",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tokenResp models.AigentTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	a.accessToken = tokenResp.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return a.accessToken, nil
}

func (a *AigentAdapter) refreshToken() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.accessToken = ""
	a.tokenExpiry = time.Time{}

	return a.fetchToken()
}

func (a *AigentAdapter) SendMessage(chatID, query, sessionID string) (*models.AigentChatResponse, error) {
	resp, err := a.doSendMessage(chatID, query, sessionID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		logger.Logger.Info("AIgent token expired, refreshing")
		if _, err := a.refreshToken(); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		resp, err = a.doSendMessage(chatID, query, sessionID)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp models.AigentChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode chat response: %w", err)
	}

	return &chatResp, nil
}

func (a *AigentAdapter) doSendMessage(chatID, query, sessionID string) (*http.Response, error) {
	token, err := a.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	chatReq := models.AigentChatRequest{
		Query:     query,
		SessionID: sessionID,
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/%s/message", a.apiURL, chatID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return a.httpClient.Do(req)
}

func (a *AigentAdapter) CreateChatbot(chatbotReq models.AigentCreateChatbotRequest) (*models.AigentCreateChatbotResponse, error) {
	resp, err := a.doCreateChatbot(chatbotReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		logger.Logger.Info("AIgent token expired, refreshing")
		if _, err := a.refreshToken(); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		resp, err = a.doCreateChatbot(chatbotReq)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create chatbot request failed with status %d", resp.StatusCode)
	}

	var chatbotResp models.AigentCreateChatbotResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatbotResp); err != nil {
		return nil, fmt.Errorf("failed to decode create chatbot response: %w", err)
	}

	return &chatbotResp, nil
}

func (a *AigentAdapter) DeleteChatbot(chatID string) error {
	resp, err := a.doDeleteChatbot(chatID)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		logger.Logger.Info("AIgent token expired, refreshing")
		if _, err := a.refreshToken(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		resp, err = a.doDeleteChatbot(chatID)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete chatbot request failed with status %d", resp.StatusCode)
	}

	return nil
}

func (a *AigentAdapter) doDeleteChatbot(chatID string) (*http.Response, error) {
	token, err := a.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	url := fmt.Sprintf("%s/chat/chatbot/%s", a.apiURL, chatID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return a.httpClient.Do(req)
}

// CreateSkill creates a new skill in AIgent
func (a *AigentAdapter) CreateSkill(req models.AigentCreateSkillRequest) (*models.AigentCreateSkillResponse, error) {
	resp, err := a.doAuthenticatedRequest(http.MethodPost, "/skills", req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		if _, err := a.refreshToken(); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		resp, err = a.doAuthenticatedRequest(http.MethodPost, "/api/skills", req)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create skill failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result models.AigentCreateSkillResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode create skill response: %w", err)
	}
	return &result, nil
}

// DeleteSkill deletes a skill from AIgent
func (a *AigentAdapter) DeleteSkill(skillID string) error {
	resp, err := a.doAuthenticatedRequest(http.MethodDelete, fmt.Sprintf("/skills/%s", skillID), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		if _, err := a.refreshToken(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		resp, err = a.doAuthenticatedRequest(http.MethodDelete, fmt.Sprintf("/skills/%s", skillID), nil)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete skill failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// CreateMCPServer attaches an MCP server to a skill
func (a *AigentAdapter) CreateMCPServer(skillID string, req models.AigentCreateMCPServerRequest) (*models.AigentMCPServerResponse, error) {
	path := fmt.Sprintf("/skills/%s/mcp-servers", skillID)
	resp, err := a.doAuthenticatedRequest(http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		if _, err := a.refreshToken(); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		resp, err = a.doAuthenticatedRequest(http.MethodPost, path, req)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create MCP server failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result models.AigentMCPServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode MCP server response: %w", err)
	}
	return &result, nil
}

// UpdateChatbot assigns skills to a chatbot
func (a *AigentAdapter) UpdateChatbot(chatbotID string, req models.AigentUpdateChatbotRequest) error {
	path := fmt.Sprintf("/chat/chatbot/%s", chatbotID)
	resp, err := a.doAuthenticatedRequest(http.MethodPut, path, req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		if _, err := a.refreshToken(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		resp, err = a.doAuthenticatedRequest(http.MethodPut, path, req)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update chatbot failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// doAuthenticatedRequest is a generic helper for authenticated API calls
func (a *AigentAdapter) doAuthenticatedRequest(method, path string, body any) (*http.Response, error) {
	token, err := a.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, a.apiURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return a.httpClient.Do(req)
}

func (a *AigentAdapter) doCreateChatbot(chatbotReq models.AigentCreateChatbotRequest) (*http.Response, error) {
	token, err := a.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	body, err := json.Marshal(chatbotReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create chatbot request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/chatbot", a.apiURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return a.httpClient.Do(req)
}
