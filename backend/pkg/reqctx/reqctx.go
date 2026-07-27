// Package reqctx carries request-scoped metadata (e.g. the browser tab's
// client id) through context.Context into the service layer, where events
// are tagged with their origin.
package reqctx

import "context"

type contextKey int

const clientIDKey contextKey = iota

// maxClientIDLength bounds the accepted client id (browser tabs send UUIDs)
const maxClientIDLength = 64

// WithClientID returns a context carrying the given client id. Empty or
// oversized ids are ignored (context returned unchanged).
func WithClientID(ctx context.Context, clientID string) context.Context {
	if clientID == "" || len(clientID) > maxClientIDLength {
		return ctx
	}
	return context.WithValue(ctx, clientIDKey, clientID)
}

// ClientID returns the client id carried by the context, or "" when absent
// (e.g. MCP requests, background jobs).
func ClientID(ctx context.Context) string {
	if id, ok := ctx.Value(clientIDKey).(string); ok {
		return id
	}
	return ""
}
