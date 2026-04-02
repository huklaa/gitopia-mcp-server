# Gitopia MCP Server

Decentralized Git + blockchain bounties + DAO governance via the Model Context Protocol.

57 tools, 3 prompts, 4 resource templates.

## Architecture

3-layer design: **MCP client** -> **handler** (internal/handler/) -> **registration** (cmd/server/main.go).

Key packages:
- `cmd/server/` — entry point, server initialization
- `internal/handler/register.go` — tool/prompt/resource registration
- `internal/handler/` — tool handler implementations (params structs + handler methods on `ToolHandler`). Includes tag_handlers.go, commit_handlers.go, release_handlers.go, label_handlers.go for tags/commits/releases/labels
- `internal/gitopia/` — gRPC client for Gitopia chain queries and tx signing
- `internal/signing/` — wallet management, transaction signing, fee grants
- `internal/config/` — configuration loading (file + env var overlay)
- `internal/git/` — local git operations
- `internal/workspace/` — sandboxed workspace path resolution

## Build & Test

```bash
go build ./cmd/server        # compile server binary
go test ./...                 # run all tests
go vet ./...                  # static analysis
go test -cover ./internal/handler/...  # handler coverage
```

## Environment Variables

| Variable | Purpose | Default |
|---|---|---|
| `GITOPIA_MNEMONIC` | BIP-39 mnemonic for signing | auto-generated if not set |
| `GITOPIA_GRPC_ENDPOINTS` | Comma-separated gRPC endpoints | `gitopia-grpc.polkachu.com:11390` |
| `DRY_RUN` | `true`/`1` to preview without broadcasting | `false` |
| `TRUST_LEVEL` | `readonly`, `localwrite`, `chainwrite` | `chainwrite` |
| `TOOLSETS` | Comma-separated toolsets (`core`, `workflow`, or `all`) | `all` |
| `TRANSPORT` | `stdio` or `http` | `stdio` |
| `PORT` | HTTP listen port (when TRANSPORT=http) | `8080` |
| `GITOPIA_WALLET_BACKEND` | `mnemonic`, `keyring`, `file`, or `auto` | `auto` |
| `APPROVAL_MODE` | `true`/`1` to require confirmation before chain-write broadcast | `false` |
| `APPROVAL_TTL` | Duration before pending transactions expire | `5m` |
| `MCP_CONFIG_FILE` | Path to JSON config file | `~/.mcp/gitopia/config.json` |
| `MCP_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `MCP_WORKSPACE_PATH` | Workspace root for cloned repos | `~/.mcp/gitopia/workspace` |
| `GIT_USER_NAME` | Default git user name | `Gitopia MCP Server` |
| `GIT_USER_EMAIL` | Default git user email | `mcp@gitopia.com` |

## Security Notes

- Message allowlist in `gitopia_wallet.go`: only `/gitopia.gitopia.gitopia.*` and `/cosmos.group.v1.*` message types are signed
- Trust tiers gate tool access (readonly < localwrite < chainwrite)
- Rate limiting on chain-write operations (default 10/min, 100/hr)
- Workspace path traversal is blocked by `workspace.Manager.ResolvePath`

## Conventions

Handler pattern (chain-write tools):
0. Trust enforcement (`withTrust` wrapper checks `h.TrustLvl` at registration)
1. Rate-limit check (`h.ChainLimiter.Allow()`)
2. Dry-run check (`h.DryRunMode` -> `DryRunResult()`)
3. Get wallet (`gitopia.WalletFromHeaders()`)
4. Validate params
5. Build SDK message
6. Sign & broadcast via `signOrHold()` (holds for approval if `APPROVAL_MODE` is enabled)
7. Audit log (`AuditLog()`)
8. Return result

Handler signature: `func (h *ToolHandler) Name(ctx, req, params) (*CallToolResult, any, error)`

Tool errors use `toolError(msg)` (sets `IsError: true`), not Go errors.
