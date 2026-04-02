# Gitopia MCP Server

MCP server for [Gitopia](https://gitopia.com) — decentralized Git with on-chain governance, bounties, and DAO management. Works with Cursor, VS Code, Claude Code, Claude Desktop, Windsurf, and any MCP-compatible tool.

**57 tools** | **3 prompts** | **4 resource templates**

Zero-config start: a wallet is auto-generated on first use. No setup beyond pasting a config snippet.

## Quick Start

### Docker (Recommended)

```bash
docker pull ghcr.io/gitopia/gitopia-mcp-server:latest
```

Then add the config for your editor (see [Editor Setup](#editor-setup) below).

### Native Binary

Download from [GitHub Releases](https://github.com/gitopia/gitopia-mcp-server/releases), or:

```bash
go install github.com/gitopia/gitopia-mcp-server/cmd/server@latest
```

Requires Go 1.25+ and [git-remote-gitopia](https://docs.gitopia.com/git-remote-gitopia) (`curl https://get.gitopia.com | bash`).

## Editor Setup

<details>
<summary><strong>OpenAI Codex</strong></summary>

```bash
codex mcp add gitopia -- docker run --rm -i --platform linux/amd64 \
  -v "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia" \
  -e MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace \
  ghcr.io/gitopia/gitopia-mcp-server:latest stdio
```

Or add to `~/.codex/config.toml`:

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

[Full Codex guide](docs/installation-guides/codex.md)

</details>

<details>
<summary><strong>Claude Code</strong></summary>

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "gitopia": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i", "--platform", "linux/amd64",
        "-v", "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia",
        "-v", "${PWD}:/workspace", "-w", "/workspace",
        "-e", "MCP_WORKSPACE_PATH=/workspace",
        "ghcr.io/gitopia/gitopia-mcp-server:latest", "stdio"
      ],
      "env": {}
    }
  }
}
```

[Full Claude Code guide](docs/installation-guides/claude-code.md)

</details>

<details>
<summary><strong>Cursor</strong></summary>

Add to `~/.cursor/mcp.json` (macOS) or `%APPDATA%\Cursor\mcp.json` (Windows):

```json
{
  "mcpServers": {
    "gitopia": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i", "--platform", "linux/amd64",
        "-v", "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia",
        "-v", "${workspaceFolder}:/workspace", "-w", "/workspace",
        "-e", "MCP_WORKSPACE_PATH=/workspace",
        "ghcr.io/gitopia/gitopia-mcp-server:latest", "stdio"
      ],
      "env": {}
    }
  }
}
```

[Full Cursor guide](docs/installation-guides/cursor.md)

</details>

<details>
<summary><strong>VS Code</strong></summary>

[Full VS Code guide](docs/installation-guides/vscode.md)

</details>

<details>
<summary><strong>Windsurf</strong></summary>

[Full Windsurf guide](docs/installation-guides/windsurf.md)

</details>

<details>
<summary><strong>Claude Desktop</strong></summary>

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
    "gitopia": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i", "--platform", "linux/amd64",
        "-v", "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia",
        "-e", "MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace",
        "ghcr.io/gitopia/gitopia-mcp-server:latest", "stdio"
      ],
      "env": {}
    }
  }
}
```

</details>

**Using an existing wallet?** Set `GITOPIA_MNEMONIC` in your shell, then add `"-e", "GITOPIA_MNEMONIC"` to the Docker args. Never hardcode mnemonics in config files.

## Tools (57)

### Context Management (4)
| Tool | Description |
|------|-------------|
| `get_user_context` | Get current identity, address, and available DAOs |
| `set_active_dao` | Switch operations to a DAO context |
| `refresh_user_context` | Reload user info from chain |
| `claim_fee_grant` | Claim fee grant from faucet |

### User Management (1)
| Tool | Description |
|------|-------------|
| `create_user` | Create on-chain Gitopia user for wallet |

### Repository Management (7)
| Tool | Description |
|------|-------------|
| `list_repos` | List repositories for owner |
| `create_repo` | Create remote repository |
| `get_repo` | Get repository details |
| `list_branches` | List branches |
| `get_file_contents` | Read file without cloning |
| `fork_repository` | Fork a repository |
| `toggle_repository_forking` | Enable/disable forking |

### Tags, Commits, Releases (4)
| Tool | Description |
|------|-------------|
| `list_tags` | List version tags |
| `list_commits` | Browse commit history for a branch |
| `list_releases` | List published releases |
| `create_release` | Publish a new release |

### Issue Management (5)
| Tool | Description |
|------|-------------|
| `list_issues` | List issues |
| `get_issue` | Get full issue details |
| `create_issue` | Create issue (on-chain) |
| `comment_on_issue` | Comment on issue (on-chain) |
| `update_issue` | Update state, labels, assignees |

### Pull Request Management (6)
| Tool | Description |
|------|-------------|
| `list_pull_requests` | List PRs |
| `get_pull_request` | Get full PR details |
| `get_pull_request_diff` | Get unified diff for a PR |
| `create_pull_request` | Create PR (on-chain) |
| `comment_on_pull_request` | Comment on PR (general or inline) |
| `merge_pull_request` | Merge PR (on-chain) |

### Git Operations (5)
| Tool | Description |
|------|-------------|
| `git_clone` | Clone repository to workspace |
| `git_push` | Push commits to remote |
| `create_feature_branch` | Create and checkout new branch |
| `sync_with_remote` | Fetch and merge/rebase |
| `commit_and_push_changes` | Stage + commit + push in one step |

### Workflow Orchestration (3)
| Tool | Description |
|------|-------------|
| `bootstrap_repo` | Create remote + local init + files + push |
| `create_feature_branch_pr` | Branch + changes + commit + push + PR |
| `update_feature_branch` | Add commits to existing branch/PR |

### Label Management (3)
| Tool | Description |
|------|-------------|
| `list_labels` | List repository labels |
| `create_label` | Create label (on-chain) |
| `delete_label` | Delete label (on-chain) |

### DAO & Governance (9)
| Tool | Description |
|------|-------------|
| `get_dao` | Get DAO details including group_id and group_policy_address |
| `create_dao` | Create DAO with members and voting |
| `dao_list_members` | List DAO members and weights |
| `dao_list_proposals` | List governance proposals |
| `dao_get_proposal` | Get proposal details with tally |
| `dao_update_members` | Add/remove/change member weights |
| `dao_submit_proposal` | Create governance proposal |
| `dao_vote` | Vote on proposal |
| `dao_exec` | Execute passed proposal |

### Bounty Management (6)
| Tool | Description |
|------|-------------|
| `list_bounties` | Discover bounties |
| `get_bounty` | Get bounty details |
| `create_bounty` | Attach crypto reward to issue |
| `update_bounty` | Modify bounty expiry |
| `close_bounty` | Deactivate bounty |
| `delete_bounty` | Permanently remove bounty |

### Batch & Approval (4)
| Tool | Description |
|------|-------------|
| `batch_execute` | Execute up to 10 ops in one transaction |
| `confirm_transaction` | Broadcast pending transaction |
| `reject_transaction` | Cancel pending transaction |
| `list_pending_transactions` | List pending transactions |

## Prompts

| Prompt | Description |
|--------|-------------|
| `fix-issue` | Complete workflow: clone, branch, fix, test, PR |
| `review-pr` | Review a pull request: inspect, comment, approve |
| `hunt-bounty` | Discover bounties, evaluate, claim, fix, submit |

## Security

- **Trust tiers** gate tool access: `readonly` < `localwrite` < `chainwrite`
- **Rate limiting** on chain-write operations (default 10/min, 100/hr)
- **Message allowlist** only signs `/gitopia.gitopia.gitopia.*` and `/cosmos.group.v1.*`
- **Workspace isolation** blocks path traversal attacks
- **Approval mode** holds transactions for explicit confirmation before broadcast
- **Non-root Docker** containers run as unprivileged user

See [SECURITY.md](SECURITY.md) for vulnerability reporting.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GITOPIA_MNEMONIC` | auto-generated | BIP-39 wallet mnemonic |
| `GITOPIA_GRPC_ENDPOINTS` | `gitopia-grpc.polkachu.com:11390` | gRPC endpoints |
| `TRUST_LEVEL` | `chainwrite` | `readonly`, `localwrite`, `chainwrite` |
| `TOOLSETS` | `all` | Comma-separated toolsets to enable (see below) |
| `DRY_RUN` | `false` | Preview transactions without broadcasting |
| `APPROVAL_MODE` | `false` | Require confirmation before chain writes |
| `MCP_WORKSPACE_PATH` | `~/.mcp/gitopia/workspace` | Workspace root |
| `MCP_LOG_LEVEL` | `info` | Log level |
| `TRANSPORT` | `stdio` | `stdio` or `http` |
| `PORT` | `8080` | HTTP listen port (when TRANSPORT=http) |

See [env.example](env.example) for the full list.

### Toolsets

Control which tools are exposed to the MCP client. Useful for reducing context window usage in AI assistants.

| Toolset | Tools | Description |
|---------|-------|-------------|
| `core` | 52 | All tools except workflow orchestrators (always included) |
| `workflow` | 5 | `bootstrap_repo`, `create_feature_branch`, `create_feature_branch_pr`, `update_feature_branch`, `commit_and_push_changes` |

- `TOOLSETS=all` (default): All 57 tools
- `TOOLSETS=core`: 52 tools (hides workflow orchestrators)
- `TOOLSETS=core,workflow`: Same as `all`

For Claude Code users, `TOOLSETS=core` is recommended since Claude can orchestrate git operations natively.

## Development

```bash
go build ./cmd/server          # build
go test ./...                  # test
go vet ./...                   # lint
docker build -t gitopia-mcp-server:latest .  # docker
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

[MIT](LICENSE)
