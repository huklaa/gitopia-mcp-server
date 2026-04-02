# Config Sources & Precedence

This page explains how the Gitopia MCP Server resolves configuration.

## Precedence (highest to lowest)

1. Environment variables (inside the server process)
2. Server config file (`~/.mcp/gitopia/config/config.json`, or path from `MCP_CONFIG_FILE`)
3. Built-in defaults

Notes:
- Docker `-e NAME=...` overrides values baked into the image via `ENV` in the Dockerfile.
- `--env-file` injects variables in bulk; any `-e` flags or host-provided env still override those.

## Environment variables

- `MCP_WORKSPACE_PATH` – workspace root (defaults to `~/.mcp/gitopia/workspace`)
- `MCP_CONFIG_FILE` – optional path to the server config JSON
- `GIT_USER_NAME`, `GIT_USER_EMAIL` – default git identity
- `MCP_LOG_LEVEL` – logging level
- `GITOPIA_MNEMONIC` – wallet mnemonic (optional; auto-generated if not set, saved to `~/.mcp/gitopia/config/wallet.key`). Set via shell env, never in config files
- `GITOPIA_GRPC_ENDPOINT` – gRPC endpoint (env or default constant)

## Server config file

Default path: `~/.mcp/gitopia/config/config.json` (inside containers: `/home/mcp/.mcp/gitopia/config/config.json`). Use it for non-secret defaults.

Example (copy from `config.example.json`):

```json
{
  "workspace": {
    "base_path": "~/.mcp/gitopia/workspace"
  },
  "git": {
    "default_user_name": "Your Name",
    "default_user_email": "your.email@example.com"
  },
  "logging": {
    "level": "info"
  }
}
```

## Code references

- `internal/config/config.go` – `LoadConfigFromEnv()` loads from `GetConfigPath()` (default or `MCP_CONFIG_FILE`), then overlays env variables.
- `cmd/server/main.go` – reads `GITOPIA_GRPC_ENDPOINT` from env or uses `gitopia.DefaultGRPC`.

## Multi-client setups

- Share non-secrets via the server config file (mounted into the container).
- Override per client via env in each client’s MCP config (`-e` flags or host env UI).
- Keep secrets out of files; prompt or use host-level secret storage.
