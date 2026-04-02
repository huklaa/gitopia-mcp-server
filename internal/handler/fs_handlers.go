package handler

import (
	"context"
	"fmt"

	"github.com/gitopia/gitopia-mcp-server/internal/fs"
	"github.com/gitopia/gitopia-mcp-server/internal/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- File System Handlers ----

type ReadFileParams struct {
	Path string `json:"path" jsonschema:"File path relative to workspace root (e.g. 'myrepo/src/main.go')"`
}

func (h *ToolHandler) ReadFile(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p ReadFileParams,
) (*mcp.CallToolResult, any, error) {
	content, err := fs.Read(h.WorkspaceMgr.GetBasePath(), p.Path)
	if err != nil {
		return toolErrorf(ErrFileSystem, "Failed to read file '%s': %s", p.Path, err)
	}
	// Exception: raw content delivery, no envelope
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: content}},
	}, nil, nil
}

type WriteFileParams struct {
	Path    string `json:"path" jsonschema:"File path relative to workspace root (e.g. 'myrepo/src/main.go')"`
	Content string `json:"content" jsonschema:"The content to write to the file."`
}

func (h *ToolHandler) WriteFile(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p WriteFileParams,
) (*mcp.CallToolResult, any, error) {
	if err := fs.Write(h.WorkspaceMgr.GetBasePath(), p.Path, p.Content); err != nil {
		return toolErrorf(ErrFileSystem, "Failed to write file '%s': %s", p.Path, err)
	}
	logging.Infof("Successfully wrote %d bytes to %s", len(p.Content), p.Path)
	return toolSuccess(fmt.Sprintf("Wrote file '%s'", p.Path), map[string]any{"bytes": len(p.Content)})
}
