package handler

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrorCode represents a machine-readable error classification.
type ErrorCode string

const (
	ErrClientUnavailable ErrorCode = "CLIENT_UNAVAILABLE"
	ErrAuthFailed        ErrorCode = "AUTH_FAILED"
	ErrRateLimited       ErrorCode = "RATE_LIMITED"
	ErrValidation        ErrorCode = "VALIDATION_ERROR"
	ErrNotFound          ErrorCode = "NOT_FOUND"
	ErrConflict          ErrorCode = "CONFLICT"
	ErrNotRepository     ErrorCode = "NOT_REPOSITORY"
	ErrPathTraversal     ErrorCode = "PATH_TRAVERSAL"
	ErrChainTx           ErrorCode = "CHAIN_TX_FAILED"
	ErrGitOperation      ErrorCode = "GIT_OPERATION_FAILED"
	ErrFileSystem        ErrorCode = "FS_ERROR"
	ErrApproval          ErrorCode = "APPROVAL_ERROR"
	ErrInternal          ErrorCode = "INTERNAL_ERROR"
)

// toolSuccess returns a structured JSON success response for mutation operations.
// Extra fields (tx_hash, url, bounty_id, etc.) are merged at the top level.
func toolSuccess(message string, extra map[string]any) (*mcp.CallToolResult, any, error) {
	resp := map[string]any{
		"status":  "success",
		"message": message,
	}
	for k, v := range extra {
		resp[k] = v
	}
	return jsonResult(resp)
}

// toolSuccessData returns a structured JSON success response for query operations.
// The payload is wrapped under a "data" key.
func toolSuccessData(message string, data any) (*mcp.CallToolResult, any, error) {
	resp := map[string]any{
		"status":  "success",
		"message": message,
		"data":    data,
	}
	return jsonResult(resp)
}

// toolErrorf returns a structured JSON error response with an error code.
// It sets IsError: true so the LLM sees the business-logic error.
func toolErrorf(code ErrorCode, format string, args ...any) (*mcp.CallToolResult, any, error) {
	msg := fmt.Sprintf(format, args...)
	resp := map[string]any{
		"status":  "error",
		"message": msg,
		"code":    string(code),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, nil, nil
	}
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

// jsonResult marshals a map into a JSON TextContent CallToolResult.
func jsonResult(resp map[string]any) (*mcp.CallToolResult, any, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}
