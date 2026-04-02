# Workspace & Meta Dirs

This page clarifies how the workspace root is chosen and where meta directories live.

## Workspace root resolution

1. Explicit path set via `MCP_WORKSPACE_PATH` (env) or `workspace.base_path` (server config file) wins.
2. Otherwise, the server falls back to `~/.mcp/gitopia/workspace`.

## Meta directories location

- Meta directories (`cache/`, `config/`, `logs/`) always reside under `~/.mcp/gitopia/` (inside containers: `$HOME/.mcp/gitopia/` -> `/home/mcp/.mcp/gitopia/`).
- Meta directories never live inside the workspace root.

## Examples

- Local development with explicit workspace:
  - `export MCP_WORKSPACE_PATH=~/my-project`
  - Workspace root: `~/my-project`
  - Meta dirs: `~/.mcp/gitopia/{cache,config,logs}`
- Docker run with an explicit workspace path:
  - Set `-e MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace`
  - Meta dirs unaffected and remain under `/home/mcp/.mcp/gitopia`

## Quick reference: persistent Docker args/env

Use these when configuring MCP hosts to persist workspace and meta dirs:

- Mount meta root: `-v ${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia`
- Set workspace path: `-e MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace`
- Optional envs: `-e GITOPIA_MNEMONIC`, `-e GITOPIA_GRPC_ENDPOINT`, `-e GIT_USER_NAME`, `-e GIT_USER_EMAIL`, `-e MCP_LOG_LEVEL`

Example:

```sh
docker run --rm -i \
  -v ${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia \
  -e MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace \
  -e GITOPIA_MNEMONIC \
  -e GITOPIA_GRPC_ENDPOINT=gitopia-grpc.polkachu.com:11390 \
  -e GIT_USER_NAME \
  -e GIT_USER_EMAIL \
  -e MCP_LOG_LEVEL=info \
  gitopia-mcp-server:latest stdio
```

Note: meta dirs always remain under `/home/mcp/.mcp/gitopia` inside the container; only the workspace root is controlled by `MCP_WORKSPACE_PATH`.

## Related configuration

- Server config file path: `~/.mcp/gitopia/config/config.json` (override with `MCP_CONFIG_FILE`).
- Git identity defaults: `GIT_USER_NAME`, `GIT_USER_EMAIL` (env) or `git.default_user_*` in the server config file.
- Logging level: `MCP_LOG_LEVEL`.
