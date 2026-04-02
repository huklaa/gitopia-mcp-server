# 🤖 Claude Desktop Integration Guide

Supercharge Claude Desktop with **decentralized Git operations** and seamless Gitopia integration!

## ✨ What You'll Get

🏗️ **Repository Management** - Create, clone, and manage Gitopia repositories directly in Claude Desktop  
🔄 **Git Operations** - Full git workflow support with AI assistance  
🎯 **Issue & Bounty Management** - Handle crypto-rewarded tasks programmatically  
🌟 **Advanced Workflows** - Feature branches, PRs, and DAO operations  
🔍 **Code Analysis** - AI-powered code review and analysis with Gitopia context

## 📋 Prerequisites

- 🐳 [Docker](https://www.docker.com/) installed and running
- No wallet needed — a wallet is auto-generated on first use

## 🚀 Quick Setup

### Step 1: Pull Docker Image

```bash
docker pull ghcr.io/gitopia/gitopia-mcp-server:latest
```

### Step 2: Configure Claude Desktop

**Open Configuration File:**
- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

**Add Gitopia Configuration:**

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

### Step 3: Restart Claude Desktop

Restart Claude Desktop to load the new MCP server configuration.

## 🎯 How It Works

### Standalone workspace
- **Dedicated Workspace:** Uses `~/.mcp/gitopia/workspace` for all operations
- **Consistent Environment:** Same workspace regardless of where Claude Desktop is launched
- **Persistent Storage:** All repositories and files persist between sessions

### Path Resolution
- All MCP operations use workspace-relative paths (e.g., `myrepo/src/main.go`)
- Paths are resolved relative to the workspace root
- Security: Path traversal attacks are prevented

## ⚙️ Configuration Details

### Wallet
A new wallet is auto-generated on first use and saved inside the mounted volume. To use an existing wallet, set `GITOPIA_MNEMONIC` in the env block:

```json
"env": {
  "GITOPIA_MNEMONIC": "your 24 word mnemonic here"
}
```

Never hardcode mnemonics in shared or committed config files.

### Environment Variables
- `GITOPIA_MNEMONIC` - BIP-39 wallet mnemonic *(optional, auto-generated if not set)*
- `GITOPIA_GRPC_ENDPOINTS` - Gitopia gRPC endpoints *(optional)*
- `MCP_WORKSPACE_PATH` - Workspace directory path

### Volume Mounts
- `${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia` - Persistent wallet, config, and workspace

## ⚠️ Troubleshooting

### Authentication Issues
- If using auto-generated wallet: check Docker volume mount is persisting `~/.mcp/gitopia`
- If using existing wallet: verify `GITOPIA_MNEMONIC` is set correctly (24 words)

### Connection Issues
- Verify `GITOPIA_GRPC_ENDPOINT` is set to `gitopia-grpc.polkachu.com:11390`
- Check your internet connection
- Ensure Docker can access external networks

### Docker Issues
- Ensure Docker is running and up to date
- Check Docker logs for error messages

### Configuration Issues
- Verify JSON syntax in your Claude Desktop configuration
- Restart Claude Desktop after configuration changes
- Check that environment variables are properly set

**Need help?** Check the [main troubleshooting guide](../README.md#troubleshooting) or [open an issue](https://github.com/gitopia/gitopia-mcp-server/issues).
