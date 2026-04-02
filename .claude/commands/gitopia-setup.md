# Gitopia Setup Verification

Verify your Gitopia MCP server environment is correctly configured.

## Steps

1. **Check wallet** — verify `GITOPIA_MNEMONIC` is set by calling `get_user_context`
2. **Test connectivity** — `get_user_context` also validates gRPC connection to Gitopia
3. **Review identity** — confirm your username, wallet address, and DAO memberships
4. **Refresh if needed** — use `refresh_user_context` after any account changes

## Example

```
1. get_user_context  -> username, address, DAOs
2. refresh_user_context  -> reload from chain (if stale)
```

## Troubleshooting

- **"missing mnemonic"** — set `GITOPIA_MNEMONIC` env var with your BIP-39 mnemonic
- **gRPC connection failed** — check `GITOPIA_GRPC_ENDPOINTS` or network connectivity
- **"trust level insufficient"** — set `TRUST_LEVEL=chainwrite` for full access
- **Rate limited** — wait a moment, or adjust rate limits in config
- **Dry-run mode** — set `DRY_RUN=false` to enable real transactions

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `GITOPIA_MNEMONIC` | Yes (for writes) | BIP-39 wallet mnemonic |
| `GITOPIA_GRPC_ENDPOINTS` | No | Custom gRPC endpoints |
| `TRUST_LEVEL` | No | `readonly`, `localwrite`, `chainwrite` |
| `DRY_RUN` | No | `true` to preview without broadcasting |
