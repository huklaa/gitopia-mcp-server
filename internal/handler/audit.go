package handler

import (
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/logging"
)

// AuditLog emits a structured log entry for chain-write operations.
// Non-blocking, uses existing logrus infrastructure.
func AuditLog(toolName, walletAddress string, success bool, duration time.Duration, detail string) {
	entry := logging.WithFields(map[string]any{
		"audit":          true,
		"tool":           toolName,
		"wallet_address": walletAddress,
		"success":        success,
		"duration_ms":    duration.Milliseconds(),
	})
	if detail != "" {
		entry = entry.WithField("detail", detail)
	}
	if success {
		entry.Info("chain transaction completed")
	} else {
		entry.Warn("chain transaction failed")
	}
}
