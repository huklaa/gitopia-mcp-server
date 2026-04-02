package handler

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DryRunResult returns a JSON preview of the operation that would be performed
// without actually signing or broadcasting it.
func DryRunResult(toolName string, params any) (*mcp.CallToolResult, any, error) {
	preview := map[string]any{
		"status":  "success",
		"dry_run": true,
		"tool":    toolName,
		"params":  params,
	}
	data, err := json.Marshal(preview)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal dry-run preview: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{
			Text: string(data),
		}},
	}, nil, nil
}
