package handlers

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"liquiswiss/internal/events"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/logger"
)

const (
	heartbeatInterval = 25 * time.Second
	orgCacheTTL       = 5 * time.Second
	// maxStreamLifetime caps a stream even if the token claims something longer
	maxStreamLifetime = 30 * time.Minute
)

// StreamEvents streams change notifications for the user's current organisation
// as Server-Sent Events. Clients reconnect automatically (EventSource), which
// re-runs the full auth middleware including the refresh token DB check.
func StreamEvents(hub *events.Hub, apiService api_service.IAPIService, c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if hub == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	// At the connection cap the hub evicts the user's oldest stream
	sub := hub.Subscribe(userID)
	defer sub.Close()

	// Stream ends at access token expiry; the client reconnects and re-authenticates
	deadline := time.Now().Add(maxStreamLifetime)
	if expiry, ok := c.Get("tokenExpiry"); ok {
		if expiryTime, ok := expiry.(time.Time); ok && expiryTime.Before(deadline) {
			deadline = expiryTime
		}
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	logger.Logger.Infof("events: stream opened for user %d", userID)
	defer logger.Logger.Infof("events: stream closed for user %d", userID)

	// Organisation is resolved at delivery time (with a short cache) so an org
	// switch or membership removal never leaks events across organisations
	var cachedOrgID int64
	var cachedOrgAt time.Time
	currentOrgID := func() (int64, bool) {
		if time.Since(cachedOrgAt) < orgCacheTTL && cachedOrgID != 0 {
			return cachedOrgID, true
		}
		organisation, err := apiService.GetCurrentOrganisation(userID)
		if err != nil {
			return 0, false
		}
		cachedOrgID = organisation.ID
		cachedOrgAt = time.Now()
		return cachedOrgID, true
	}

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()
	expiryTimer := time.NewTimer(time.Until(deadline))
	defer expiryTimer.Stop()

	clientGone := c.Request.Context().Done()

	for {
		select {
		case <-clientGone:
			return
		case <-sub.Done:
			return
		case <-expiryTimer.C:
			return
		case <-heartbeat.C:
			_, _ = io.WriteString(c.Writer, ": ping\n\n")
			c.Writer.Flush()
		case event := <-sub.Events:
			orgID, ok := currentOrgID()
			if !ok || orgID != event.OrganisationID {
				continue
			}
			c.SSEvent("change", event)
			c.Writer.Flush()
		}
	}
}
