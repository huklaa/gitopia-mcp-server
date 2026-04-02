# Workspace Configuration

The server has one workspace root. All file operations, git clones, and local repositories live under this single directory.

## How It Works

At startup, the server resolves the workspace root in this order:

1. `MCP_WORKSPACE_PATH` environment variable (if set)
2. `workspace.base_path` from config file (if set)
3. Default: `~/.mcp/gitopia/workspace`

Every tool that touches the filesystem (git_clone, bootstrap_repo, etc.) operates relative to this root. Paths outside it are blocked.

## Common Configurations

### Attached to Your Project

Point the workspace at your editor's project directory. MCP reads and writes your actual project files.

**Docker:**
```
-v "${workspaceFolder}:/workspace"
-w "/workspace"
-e "MCP_WORKSPACE_PATH=/workspace"
```

**Native binary:**
```
MCP_WORKSPACE_PATH=/path/to/your/project
```

Use this when you want the AI to edit your code, create branches in your repo, and push changes.

### Standalone Workspace

Leave `MCP_WORKSPACE_PATH` unset. The server uses `~/.mcp/gitopia/workspace` and manages its own directory.

**Docker:**
```
-e "MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace"
```

**Native binary:**
No configuration needed. The default is used.

Use this for general Gitopia operations: cloning repos, browsing code, fixing bounties. Each `git_clone` creates a subdirectory (e.g., `myrepo/`) under the workspace root.

### Multiple Repos

Clone multiple repos into the same workspace. Each gets its own subdirectory:

```
git_clone(repo_url="gitopia://dao-alpha/frontend", local_path="alpha-frontend")
git_clone(repo_url="gitopia://dao-beta/contracts", local_path="beta-contracts")
```

Both repos live under the workspace root. Switch between them using the `repo_path` parameter on git tools.

## Meta Directories

Wallet, config, and cache files are stored separately from the workspace at `~/.mcp/gitopia/`:

```
~/.mcp/gitopia/
├── config/
│   ├── config.json       # Server configuration
│   └── wallet.key        # Auto-generated wallet (if no GITOPIA_MNEMONIC)
├── cache/                # Cache files
├── logs/                 # Log files
└── workspace/            # Default workspace root (only if MCP_WORKSPACE_PATH is unset)
```

In Docker, mount this directory to persist wallet and config between container restarts:
```
-v "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia"
```
