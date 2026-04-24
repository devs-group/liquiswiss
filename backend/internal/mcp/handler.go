package mcp

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/logger"
)

// JSON-RPC types
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id"`
	Result  any         `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

func HandleMCP(dbService db_adapter.IDatabaseAdapter, c *gin.Context) {
	// Extract and verify MCP JWT token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.VerifyMCPToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Validate chatbot still exists
	has, err := dbService.HasOrganisationChatbots(claims.OrganisationID)
	if err != nil || !has {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "chatbot no longer active"})
		return
	}

	var req jsonRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error:   &rpcError{Code: -32700, Message: "parse error"},
		})
		return
	}

	var result any
	var rpcErr *rpcError

	switch req.Method {
	case "initialize":
		result = handleInitialize()
	case "tools/list":
		result = handleToolsList()
	case "tools/call":
		result, rpcErr = handleToolsCall(req.Params, claims.OrganisationID, dbService)
	default:
		rpcErr = &rpcError{Code: -32601, Message: "method not found"}
	}

	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}
	if rpcErr != nil {
		resp.Error = rpcErr
		logger.Logger.Errorf("MCP RPC error for org %d, method %s: %s", claims.OrganisationID, req.Method, rpcErr.Message)
	} else {
		resp.Result = result
	}

	c.JSON(http.StatusOK, resp)
}

func handleInitialize() any {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "liquiswiss-mcp",
			"version": "1.0.0",
		},
	}
}
