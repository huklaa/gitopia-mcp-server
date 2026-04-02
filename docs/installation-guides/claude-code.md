# Claude Code Setup

Configure the Gitopia MCP Server for use with [Claude Code](https://claude.ai/code).

## Option 1: Docker (Recommended)

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "gitopia": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i", "--platform", "linux/amd64",
        "-v", "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia",
        "-v", "${PWD}:/workspace",
        "-w", "/workspace",
        "-e", "MCP_WORKSPACE_PATH=/workspace",
        "ghcr.io/gitopia/gitopia-mcp-server:latest", "stdio"
      ],
      "env": {}
    }
  }
}
```

A new wallet is auto-generated on first use. To use an existing wallet, set `GITOPIA_MNEMONIC` in your shell environment and add `"-e", "GITOPIA_MNEMONIC"` to the Docker args.

## Option 2: Native Binary

If you have the `gitopia-mcp-server` binary installed locally, use the full path. Requires [git-remote-gitopia](https://docs.gitopia.com/git-remote-gitopia) for git clone/push (`curl https://get.gitopia.com | bash`):

```json
{
  "mcpServers": {
    "gitopia": {
      "command": "/path/to/gitopia-mcp-server",
      "env": {
        "MCP_WORKSPACE_PATH": "${PWD}",
        "MCP_LOG_LEVEL": "info",
        "TRUST_LEVEL": "chainwrite"
      }
    }
  }
}
```

If installed via `go install`, the binary is at `$(go env GOPATH)/bin/gitopia-mcp-server`.

## Trust Levels

Control tool access via `TRUST_LEVEL`:

| Level | Access |
|-------|--------|
| `readonly` | Query operations only (list repos, issues, PRs, tags, commits, releases, labels) |
| `localwrite` | Readonly + local file/git operations (write, add, commit, clone) |
| `chainwrite` | Full access including on-chain transactions (default) |

## Approval Mode

For extra safety, enable approval mode to review chain-write transactions before they broadcast:

```json
"env": {
  "APPROVAL_MODE": "true",
  "TRUST_LEVEL": "chainwrite"
}
```

Transactions are held until you confirm them via the `confirm_transaction` tool.
