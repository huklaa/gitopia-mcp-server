# TODOs

## v0.2

### Remove git_clone and git_push tools
**What:** Remove `git_clone` and `git_push` from the MCP server. They wrap `git clone` and `git push` with temporary wallet-based credential setup (`AuthMgr.SetupAuth`), but if `git-remote-gitopia` is installed and the user's wallet is configured (keyring, env var, or git config), plain `git clone gitopia://owner/repo` works natively from any MCP client's terminal.

**Why:** These tools only add value in the Docker/auto-wallet scenario where no persistent git credential config exists. For Claude Code, Cursor, Windsurf, and Codex users with `git-remote-gitopia` in PATH, they're redundant. Removing them drops the tool count from 57 to 55 and further narrows the server to pure chain operations.

**Depends on:** Document the prerequisite (`git-remote-gitopia` must be installed) clearly in README. Consider whether the Docker image setup flow needs changes.

### Add issue/PR search and filtering
**What:** Add `state`, `labels`, `assignee`, and `search` filter parameters to `list_issues` and `list_pull_requests`.

**Why:** Currently these tools return all items with no filtering. For repos with 100+ issues, the agent gets a massive JSON response and burns context tokens. GitHub MCP has `search_issues` with full query support. GitLab has `search` with scope filtering.

**Depends on:** Check if the Gitopia gRPC API supports filtered queries or if filtering must be done client-side.

### Tool consolidation (issue_read/issue_write pattern)
**What:** Evaluate GitHub's pattern of consolidating CRUD tools using a `method` parameter (e.g. `issue_read` with `method="get"|"list"|"search"`). Could reduce tool count from ~55 to ~30-35.

**Why:** Even with TOOLSETS filtering, fewer tools means better agent selection accuracy. GitHub proved this with their Copilot benchmarks.

**Depends on:** Real-world usage data from v0.1.0. May not be needed if TOOLSETS=core works well enough.

### Add approval_mode to get_user_context response
**What:** Include `approval_mode: true/false` and `trust_level` in the `get_user_context` response so the agent knows the server's operating mode upfront.

**Why:** When `APPROVAL_MODE=true`, chain-write tools silently return `pending_id` instead of executing. The agent has no way to know this is happening until it sees the unexpected response. Surfacing the mode lets the agent adjust its workflow.

**Depends on:** Nothing. Small change to `context_handlers.go`.
