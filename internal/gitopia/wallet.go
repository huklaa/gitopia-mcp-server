package gitopia

import (
	"context"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Wallet = signing.Wallet // type alias for brevity

// WalletFromHeaders constructs a signing wallet using the configured backend.
// If GITOPIA_WALLET_BACKEND is set, it delegates to signing.InitWallet.
// Otherwise, it uses the fast path: GITOPIA_MNEMONIC env var directly.
func WalletFromHeaders(_ *mcp.ServerSession, client *Client) (signing.Wallet, error) {
	if client == nil {
		return nil, fmt.Errorf("gitopia client is not available")
	}

	bankClient := types.NewQueryClient(client.conn)
	feegrantClient := feegrant.NewQueryClient(client.conn)

	backend := os.Getenv("GITOPIA_WALLET_BACKEND")
	if backend == "" {
		// Fast path: if mnemonic is set, use it directly (backward compatible)
		if mnemonic := os.Getenv("GITOPIA_MNEMONIC"); mnemonic != "" {
			return signing.FromMnemonicWithFeegrant(context.Background(), mnemonic, bankClient, feegrantClient)
		}
		// Fall through to InitWallet with "auto" for full fallback chain
		backend = "auto"
	}
	return signing.InitWallet(context.Background(), bankClient, feegrantClient, backend)
}
