# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x     | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in the Gitopia MCP Server, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please email: **security@gitopia.com**

### What to include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Acknowledgement**: Within 48 hours of report
- **Initial assessment**: Within 5 business days
- **Fix timeline**: Depends on severity, typically within 30 days for critical issues

## Wallet and Mnemonic Handling

This server handles sensitive wallet mnemonics for signing Gitopia blockchain transactions. Please note:

- User-provided mnemonics via `GITOPIA_MNEMONIC` env var are held only in process memory
- When using the `file` wallet backend (default for auto-generated wallets), a derived signing key is persisted to `~/.mcp/gitopia/config/wallet.key` with restricted permissions (chmod 600). Never commit wallet.key files to version control
- Use `env.example` as a reference; never commit `.env` files with real credentials
- In production, use a secrets manager or secure environment variable injection
- In Docker, the wallet file lives inside the mounted volume at `/home/mcp/.mcp/gitopia/config/`
- The server validates message types before signing to prevent unauthorized transactions (only `/gitopia.gitopia.gitopia.*` and `/cosmos.group.v1.*` are allowed)
