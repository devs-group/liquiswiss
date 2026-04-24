package models

type AigentTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AigentTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type AigentChatRequest struct {
	Query     string `json:"query"`
	SessionID string `json:"session_id"`
}

type AigentChatResponse struct {
	Response  string `json:"response"`
	SessionID string `json:"session_id"`
}

type AigentCreateChatbotRequest struct {
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	SystemInstructions string  `json:"system_instructions"`
	ModelName          string  `json:"model_name"`
	MaxTokens          int     `json:"max_tokens"`
	TemperatureParam   float64 `json:"temperature_param"`
	AgentEnabled       bool    `json:"agent_enabled,omitempty"`
	MaxAgentTurns      int     `json:"max_agent_turns,omitempty"`
	MaxToolCalls       int     `json:"max_tool_calls,omitempty"`
}

type AigentCreateChatbotResponse struct {
	ID string `json:"id"`
}

type AigentCreateSkillRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PromptMD    string `json:"prompt_md"`
}

type AigentCreateSkillResponse struct {
	ID string `json:"id"`
}

type AigentCreateMCPServerRequest struct {
	Alias   string            `json:"alias"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type AigentMCPServerResponse struct {
	ID      string `json:"id"`
	Alias   string `json:"alias"`
	URL     string `json:"url"`
	SkillID string `json:"skill_id"`
}

type AigentUpdateChatbotRequest struct {
	SkillIDs []string `json:"skill_ids"`
}

type OrganisationChatbot struct {
	ID             int64   `db:"id" json:"id"`
	OrganisationID int64   `db:"organisation_id" json:"organisationID"`
	ChatbotID      string  `db:"chatbot_id" json:"chatbotID"`
	SkillID        *string `db:"skill_id" json:"skillID"`
}
