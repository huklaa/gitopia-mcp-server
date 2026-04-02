# 💻 VS Code Integration Guide

Supercharge VS Code with **decentralized Git operations** and seamless Gitopia integration!

## ✨ What You'll Get

🏗️ **Repository Management** - Create, clone, and manage Gitopia repositories directly in VS Code  
🔄 **Git Operations** - Full git workflow support with AI assistance  
🎯 **Issue & Bounty Management** - Handle crypto-rewarded tasks programmatically  
🌟 **Advanced Workflows** - Feature branches, PRs, and DAO operations  
📁 **Workspace Integration** - Seamless file sharing between editor and MCP

## 📋 Prerequisites

- 🐳 [Docker](https://www.docker.com/) installed and running
- GitHub Copilot subscription (for VS Code MCP support)
- No wallet needed — a wallet is auto-generated on first use

## 🚀 Quick Setup

### Option 1: One-Click Installation (Recommended)

For the fastest setup, use our one-click installation buttons:

[![Install with Docker in VS Code](https://img.shields.io/badge/VS_Code-Install_Server-0098FF?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=gitopia&inputs=%5B%7B%22id%22%3A%22gitopia_mnemonic%22%2C%22type%22%3A%22promptString%22%2C%22description%22%3A%22Gitopia%20Wallet%20Mnemonic%20(24%20words)%22%2C%22password%22%3Atrue%7D%5D&config=%7B%22servers%22%3A%7B%22gitopia%22%3A%7B%22command%22%3A%22docker%22%2C%22args%22%3A%5B%22run%22%2C%22--rm%22%2C%22-i%22%2C%22--platform%22%2C%22linux%2Famd64%22%2C%22-v%22%2C%22gitopia-mcp-meta%3A%2Fhome%2Fmcp%2F.mcp%2Fgitopia%22%2C%22-e%22%2C%22GITOPIA_MNEMONIC%22%2C%22-e%22%2C%22GITOPIA_GRPC_ENDPOINT%22%2C%22-e%22%2C%22MCP_WORKSPACE_PATH%22%2C%22gitopia-mcp-server%3Alatest%22%2C%22stdio%22%5D%2C%22env%22%3A%7B%22GITOPIA_MNEMONIC%22%3A%22%24%7Binput%3Agitopia_mnemonic%7D%22%2C%22GITOPIA_GRPC_ENDPOINT%22%3A%22gitopia-grpc.polkachu.com%3A11390%22%2C%22MCP_WORKSPACE_PATH%22%3A%22%2Fhome%2Fmcp%2F.mcp%2Fgitopia%2Fworkspace%22%7D%7D%7D%7D) [![Install with Docker in VS Code Insiders](https://img.shields.io/badge/VS_Code_Insiders-Install_Server-24bfa5?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=gitopia&inputs=%5B%7B%22id%22%3A%22gitopia_mnemonic%22%2C%22type%22%3A%22promptString%22%2C%22description%22%3A%22Gitopia%20Wallet%20Mnemonic%20(24%20words)%22%2C%22password%22%3Atrue%7D%5D&config=%7B%22servers%22%3A%7B%22gitopia%22%3A%7B%22command%22%3A%22docker%22%2C%22args%22%3A%5B%22run%22%2C%22--rm%22%2C%22-i%22%2C%22--platform%22%2C%22linux%2Famd64%22%2C%22-v%22%2C%22gitopia-mcp-meta%3A%2Fhome%2Fmcp%2F.mcp%2Fgitopia%22%2C%22-e%22%2C%22GITOPIA_MNEMONIC%22%2C%22-e%22%2C%22GITOPIA_GRPC_ENDPOINT%22%2C%22-e%22%2C%22MCP_WORKSPACE_PATH%22%2C%22gitopia-mcp-server%3Alatest%22%2C%22stdio%22%5D%2C%22env%22%3A%7B%22GITOPIA_MNEMONIC%22%3A%22%24%7Binput%3Agitopia_mnemonic%7D%22%2C%22GITOPIA_GRPC_ENDPOINT%22%3A%22gitopia-grpc.polkachu.com%3A11390%22%2C%22MCP_WORKSPACE_PATH%22%3A%22%2Fhome%2Fmcp%2F.mcp%2Fgitopia%2Fworkspace%22%7D%7D%7D%7D)

### Option 2: Manual Configuration

**Step 1: Pull Docker Image**

```bash
docker pull ghcr.io/gitopia/gitopia-mcp-server:latest
```

**Step 2: Open MCP Configuration**

You can configure VS Code in several ways:
- Run **MCP: Open User Configuration** command from Command Palette
- Use **MCP: Add Server** command and select **Global**
- Or manually edit the configuration file:
  - **macOS/Linux:** `~/.config/Code/User/mcp.json`
  - **Windows:** `%APPDATA%\Code\User\mcp.json`

**Step 3: Choose Configuration Mode**

**🎯 Attached to your project (recommended)**  
✅ Mounts your current VS Code workspace as the MCP workspace  
✅ Real-time file synchronization between editor and MCP  
✅ All operations work directly on your project files  

**🏠 Standalone workspace**  
✅ Uses dedicated workspace directory  
✅ Consistent workspace regardless of launch context  
✅ Ideal for general Gitopia operations  

**Step 4: Add Configuration (Attached workspace)**

```json
{
  "servers": {
    "gitopia": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "--platform",
        "linux/amd64",
        "-v",
        "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia",
        "-v",
        "${workspaceFolder}:/workspace",
        "-w",
        "/workspace",
        "-e",
        "MCP_WORKSPACE_PATH=/workspace",
        "ghcr.io/gitopia/gitopia-mcp-server:latest",
        "stdio"
      ],
      "env": {}
    }
  }
}
```

> **Wallet:** A new wallet is auto-generated on first use. To use an existing wallet, add `"-e", "GITOPIA_MNEMONIC"` to Docker args and set it in your shell: `export GITOPIA_MNEMONIC="your 24 words"`.

**Alternative: Standalone workspace**

```json
{
  "servers": {
    "gitopia": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "--platform",
        "linux/amd64",
        "-v",
        "${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia",
        "-e",
        "MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace",
        "ghcr.io/gitopia/gitopia-mcp-server:latest",
        "stdio"
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
  "servers": {
    "gitopia": {
      "type": "stdio",
      "command": "/path/to/gitopia-mcp-server",
      "env": {
        "TRUST_LEVEL": "chainwrite"
      }
    }
  }
}
```

**Step 5: Restart VS Code**

Restart VS Code to load the new MCP server configuration.

## 🎯 How It Works

### Attached workspace
- **Workspace Mount:** `${workspaceFolder}:/workspace` mounts your VS Code project directory
- **Working Directory:** `-w /workspace` sets the container's working directory
- **MCP_WORKSPACE_PATH=/workspace:** Points MCP at the mounted project directory
- **File Sharing:** Changes in VS Code immediately appear in MCP and vice versa

### Path Resolution
- All MCP operations use workspace-relative paths (e.g., `myrepo/src/main.go`)
- Paths are resolved relative to your project root
- Security: Path traversal attacks are prevented

## ⚙️ Configuration Details

### Wallet
A new wallet is auto-generated on first use and saved to `~/.mcp/gitopia/config/wallet.key` (inside the mounted volume). To use an existing wallet, set `GITOPIA_MNEMONIC` in your shell environment — never hardcode it in config files.

### Environment Variables
- `GITOPIA_MNEMONIC`: Your 24-word wallet mnemonic (optional — auto-generated if not set)
- `GITOPIA_GRPC_ENDPOINT`: Gitopia gRPC endpoint (optional, defaults to `gitopia-grpc.polkachu.com:11390`)
- `GIT_USER_NAME`: Git commit author name (optional)
- `GIT_USER_EMAIL`: Git commit author email (optional)
- `MCP_WORKSPACE_PATH=/workspace`: Container workspace path

### Volume Mounts
- `${HOME}/.mcp/gitopia:/home/mcp/.mcp/gitopia`: Persistent meta directories (cache, config, logs)
- `${workspaceFolder}:/workspace`: Your project directory (Project Integration mode only)

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

### Configuration Issues
- Verify JSON syntax in your MCP configuration
- Restart VS Code after configuration changes
- Check that the MCP extension is properly installed

**Need help?** Check the [main troubleshooting guide](../README.md#troubleshooting) or [open an issue](https://github.com/gitopia/gitopia-mcp-server/issues).
