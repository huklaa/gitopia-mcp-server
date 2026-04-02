# 🎯 Cursor Integration Guide

Supercharge Cursor with **decentralized Git operations** and seamless Gitopia integration!

## ✨ What You'll Get

🏗️ **Repository Management** - Create, clone, and manage Gitopia repositories directly in Cursor  
🔄 **Git Operations** - Full git workflow support with AI assistance  
🎯 **Issue & Bounty Management** - Handle crypto-rewarded tasks programmatically  
🌟 **Advanced Workflows** - Feature branches, PRs, and DAO operations  

## 📋 Prerequisites

- 🐳 [Docker](https://www.docker.com/) installed and running
- No wallet needed — a wallet is auto-generated on first use

## 🚀 Quick Setup

### Step 1: Pull Docker Image

```bash
docker pull ghcr.io/gitopia/gitopia-mcp-server:latest
```

### Step 2: Configure Cursor

**Open MCP Configuration File:**
- **macOS:** `~/.cursor/mcp.json`
- **Windows:** `%APPDATA%\Cursor\mcp.json`
- **Linux:** `~/.config/cursor/mcp.json`

**Add Gitopia Configuration:**

```json
{
  "mcpServers": {
    "gitopia": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i", "--platform", "linux/amd64",
        "-v", "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia",
        "-v", "${workspaceFolder}:/workspace",
        "-w", "/workspace",
        "-e", "MCP_WORKSPACE_PATH=/workspace",
        "ghcr.io/gitopia/gitopia-mcp-server:latest", "stdio"
      ],
      "env": {}
    }
  }
}
```

**Alternative: Native Binary**

If you have the binary installed locally (via [GitHub Releases](https://github.com/gitopia/gitopia-mcp-server/releases) or `go install`). Requires [git-remote-gitopia](https://docs.gitopia.com/git-remote-gitopia) for git clone/push (`curl https://get.gitopia.com | bash`):

```json
{
  "mcpServers": {
    "gitopia": {
      "command": "/path/to/gitopia-mcp-server",
      "env": {
        "TRUST_LEVEL": "chainwrite"
      }
    }
  }
}
```

### Step 3: Restart Cursor

Restart Cursor to load the new MCP server configuration.

## 🎯 How It Works

### Attached workspace
- **Workspace Mount:** `${workspaceFolder}:/workspace` mounts your Cursor project directory
- **Working Directory:** `-w /workspace` sets the container's working directory
- **MCP_WORKSPACE_PATH=/workspace:** Points MCP at the mounted project directory
- **File Sharing:** Changes in Cursor immediately appear in MCP and vice versa

### Path Resolution
- All MCP operations use workspace-relative paths (e.g., `myrepo/src/main.go`)
- Paths are resolved relative to your project root
- Security: Path traversal attacks are prevented

## ⚙️ Configuration Details

### Wallet
A new wallet is auto-generated on first use and saved to `~/.mcp/gitopia/config/wallet.key`. To use an existing wallet, set `GITOPIA_MNEMONIC` in your shell environment — never hardcode it in config files.

### Environment Variables
- `GITOPIA_MNEMONIC` - Your 24-word wallet mnemonic *(optional — auto-generated if not set)*
- `GITOPIA_GRPC_ENDPOINT` - Gitopia gRPC endpoint *(optional)*
- `GIT_USER_NAME` - Git commit author name *(optional)*
- `GIT_USER_EMAIL` - Git commit author email *(optional)*
- `MCP_WORKSPACE_PATH=/workspace` - Container workspace path

### Volume Mounts
- `${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia` - Persistent meta directories (cache, config, logs)
- `${workspaceFolder}:/workspace` - Your project directory

## ⚠️ Troubleshooting

### File Access Issues
- Ensure Docker has permission to access your project directory
- On macOS: Check Docker Desktop > Settings > Resources > File Sharing
- On Linux: Ensure proper file permissions

### Git Authentication
- The server automatically configures git authentication using your Gitopia wallet
- No additional git credentials needed for Gitopia repositories

### Workspace Path Issues
- Use relative paths in MCP operations (e.g., `myrepo` not `/absolute/path/myrepo`)
- The server automatically resolves paths relative to your project root

**Need help?** Check the [main troubleshooting guide](../README.md#troubleshooting) or [open an issue](https://github.com/gitopia/gitopia-mcp-server/issues).
