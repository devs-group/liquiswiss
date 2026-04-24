package handlers

import (
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/mcp"
)

func HandleMCP(dbService db_adapter.IDatabaseAdapter, c *gin.Context) {
	mcp.HandleMCP(dbService, c)
}
