# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-04-02

First public release. 57 tools, 3 prompts, 4 resource templates.

### Highlights

- **57 MCP tools** for Gitopia's decentralized git platform: repositories, issues, pull requests, DAO governance, bounties, labels, releases, tags, commits
- **Trust tiers** (readonly, localwrite, chainwrite) enforced at runtime via `withTrust` wrapper
- **Toolset filtering** via `TOOLSETS` env var — reduce context window usage for AI assistants
- **Auto-wallet** generation on first use with fee grant claiming — zero-config start
- **Approval mode** for human review of chain-write transactions before broadcast
- **Batch execution** of up to 10 operations in a single atomic transaction
- **3 workflow prompts**: fix-issue, review-pr, hunt-bounty
- **4 resource templates**: repos, issues, PRs, bounties via `gitopia://` URIs

### Tools

| Category | Count | Tools |
|----------|-------|-------|
| Context | 4 | `get_user_context`, `set_active_dao`, `refresh_user_context`, `claim_fee_grant` |
| User | 1 | `create_user` |
| Repos | 5 | `list_repos`, `get_repo`, `create_repo`, `fork_repository`, `toggle_repository_forking` |
| Issues | 5 | `list_issues`, `get_issue`, `create_issue`, `update_issue`, `comment_on_issue` |
| PRs | 6 | `list_pull_requests`, `get_pull_request`, `get_pull_request_diff`, `create_pull_request`, `comment_on_pull_request`, `merge_pull_request` |
| Git | 5 | `git_clone`, `git_push`, `create_feature_branch`, `sync_with_remote`, `commit_and_push_changes` |
| Metadata | 5 | `list_branches`, `list_tags`, `list_commits`, `list_releases`, `list_labels` |
| Releases | 1 | `create_release` |
| Labels | 2 | `create_label`, `delete_label` |
| DAO | 9 | `create_dao`, `get_dao`, `dao_list_members`, `dao_update_members`, `dao_list_proposals`, `dao_get_proposal`, `dao_submit_proposal`, `dao_vote`, `dao_exec` |
| Bounties | 6 | `list_bounties`, `get_bounty`, `create_bounty`, `update_bounty`, `close_bounty`, `delete_bounty` |
| Workflow | 3 | `bootstrap_repo`, `create_feature_branch_pr`, `update_feature_branch` |
| Tx | 4 | `batch_execute`, `confirm_transaction`, `reject_transaction`, `list_pending_transactions` |

### Security

- Message allowlist: only `/gitopia.gitopia.gitopia.*` and `/cosmos.group.v1.*` types signed
- Trust tier enforcement on every tool call
- Workspace path traversal protection via `filepath.Rel` + `EvalSymlinks`
- Rate limiting on chain-write operations (default 10/min, 100/hr)
- Session-scoped pending transactions with TTL expiry
- Non-root Docker container
- Wallet key file stored with 0600 permissions

### Supported Clients

- Claude Code (`.mcp.json`)
- Cursor
- VS Code (GitHub Copilot)
- Windsurf
- Claude Desktop
- OpenAI Codex

### Distribution

- Docker: `ghcr.io/gitopia/gitopia-mcp-server:latest`
- GitHub Releases: binaries for linux/darwin x amd64/arm64
- Go install: `go install github.com/gitopia/gitopia-mcp-server/cmd/server@latest`
