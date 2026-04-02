# OpenAI Codex Setup

Configure the Gitopia MCP Server for use with [OpenAI Codex CLI](https://developers.openai.com/codex/cli).

## Option 1: CLI (Recommended)

### Docker

```bash
codex mcp add gitopia -- docker run --rm -i --platform linux/amd64 \
  -v "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia" \
  -e MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace \
  ghcr.io/gitopia/gitopia-mcp-server:latest stdio
```

### Native Binary

Use the full path to the server binary (relative paths won't work outside the install directory). Requires [git-remote-gitopia](https://docs.gitopia.com/git-remote-gitopia) for git clone/push (`curl https://get.gitopia.com | bash`):

```bash
codex mcp add gitopia -- /path/to/gitopia-mcp-server stdio
```

Or with environment variables:

```bash
codex mcp add gitopia \
  --env TRUST_LEVEL=chainwrite \
  --env MCP_WORKSPACE_PATH=/tmp/gitopia-workspace \
  -- /path/to/gitopia-mcp-server stdio
```

## Option 2: Manual Config

Edit `~/.codex/config.toml` (global) or `.codex/config.toml` (project-scoped):

### Docker

```toml
[mcp_servers.gitopia]
command = "docker"
args = [
  "run", "--rm", "-i", "--platform", "linux/amd64",
  "-v", "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia",
  "-e", "MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace",
  "ghcr.io/gitopia/gitopia-mcp-server:latest", "stdio"
]
```

### Native Binary

```toml
[mcp_servers.gitopia]
command = "/path/to/gitopia-mcp-server"
args = ["stdio"]

[mcp_servers.gitopia.env]
TRUST_LEVEL = "chainwrite"
MCP_WORKSPACE_PATH = "/tmp/gitopia-workspace"
MCP_LOG_LEVEL = "info"
```

## Verification

After adding the server:

1. Run `/mcp` in the Codex TUI to confirm `gitopia` shows tools
2. Test with: "List all repositories for the gitopia organization"
3. Codex should call `list_repos` and return results

## Using an Existing Wallet

To use an existing Gitopia wallet, add the mnemonic as an environment variable:

```bash
codex mcp add gitopia \
  --env GITOPIA_MNEMONIC="your 24 word mnemonic here" \
  --env TRUST_LEVEL=chainwrite \
  -- ./server stdio
```

Or in config.toml:

```toml
[mcp_servers.gitopia.env]
GITOPIA_MNEMONIC = "your 24 word mnemonic here"
```

If no mnemonic is provided, a new wallet is auto-generated on first use.

## Trust Levels

| Level | Access |
|-------|--------|
| `readonly` | Query operations only (list repos, issues, PRs, tags, commits, releases, labels) |
| `localwrite` | Readonly + local file/git operations (write, add, commit, clone) |
| `chainwrite` | Full access including on-chain transactions (default) |
