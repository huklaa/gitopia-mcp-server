# Quickstart Guide

Get up and running with the Gitopia MCP Server in 5 minutes.

## Prerequisites

- One of: Docker (recommended), Go 1.25+, or a pre-built binary
- No wallet needed — a wallet is auto-generated on first use

## Installation

### Option 1: Docker (Recommended)

```bash
docker pull ghcr.io/gitopia/gitopia-mcp-server:latest
```

### Option 2: Go Install

Requires [git-remote-gitopia](https://docs.gitopia.com/git-remote-gitopia) for git clone/push: `curl https://get.gitopia.com | bash`

```bash
go install github.com/gitopia/gitopia-mcp-server/cmd/server@latest
```

### Option 3: Build from Source

Requires [git-remote-gitopia](https://docs.gitopia.com/git-remote-gitopia) for git clone/push: `curl https://get.gitopia.com | bash`

```bash
git clone https://github.com/gitopia/gitopia-mcp-server.git
cd gitopia-mcp-server
make build
```

## Configuration

No configuration is required for basic usage. A wallet is auto-generated on first use and saved to `~/.mcp/gitopia/config/wallet.key`.

To use an existing wallet, set your mnemonic as an environment variable:

```bash
export GITOPIA_MNEMONIC="your twelve or twenty-four word mnemonic phrase here"
```

Optional configuration (see `env.example` for all options):

```bash
export GITOPIA_GRPC_ENDPOINTS="gitopia-grpc.polkachu.com:11390"  # default
export MCP_WORKSPACE_PATH="$HOME/.mcp/gitopia/workspace"         # default
export TRUST_LEVEL="chainwrite"                                   # readonly | localwrite | chainwrite
```

## Running the Server

### Standalone (stdio transport)

```bash
./server
```

### With Docker

```bash
docker run -i ghcr.io/gitopia/gitopia-mcp-server:latest
```

### Editor Setup

See the [README](../../README.md#editor-setup) for Claude Code, Cursor, VS Code, Windsurf, and Claude Desktop configurations.

## First Session Walkthrough

Once connected, try these tools in order:

1. **Discover your identity**
   ```
   Tool: get_user_context
   ```
   Returns your username, wallet address, and available DAOs.
   On first use, the server auto-generates a wallet and creates an on-chain user.

2. **Browse repositories**
   ```
   Tool: list_repos
   Params: {"owner": "gitopia"}
   ```

3. **Create a repository**
   ```
   Tool: create_repo
   Params: {"name": "my-first-repo"}
   ```
   If `owner_id` is omitted, uses your current identity.

4. **Create an issue**
   ```
   Tool: create_issue
   Params: {"owner": "your-username", "name": "my-first-repo", "title": "Hello world", "description": "First issue"}
   ```

5. **Browse commit history**
   ```
   Tool: list_commits
   Params: {"owner": "gitopia", "name": "gitopia"}
   ```
   If `branch` is omitted, uses the repo's default branch.

6. **Discover bounties**
   ```
   Tool: list_bounties
   ```

7. **Look up a DAO**
   ```
   Tool: get_dao
   Params: {"id": "gitopia"}
   ```
   Returns group_id and group_policy_address needed for governance tools.

## Bounty Discovery Example

A typical agent workflow for finding and working on bounties:

```
1. list_bounties          -> Find open bounties
2. get_bounty             -> Get bounty details (amount, expiry)
3. get_issue              -> Read the linked issue
4. git_clone                      -> Clone the repository
5. (use your editor's built-in tools to understand the codebase)
6. create_feature_branch_pr       -> Make changes and open a PR
7. comment_on_issue       -> Report progress on the issue
```

## Trust Levels

Control what the server can do by setting `TRUST_LEVEL`:

| Level | Capabilities |
|-------|-------------|
| `readonly` | Query tools only (list, get, search, read) |
| `localwrite` | + Local git/file operations (clone, commit, write) |
| `chainwrite` | + On-chain transactions (create repo, issue, PR, bounty) |

## Next Steps

- See [Configuration Guide](../configuration/) for advanced settings
- See [Installation Guides](../installation-guides/) for platform-specific setup
