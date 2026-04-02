package signing

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/grpc"
)

// testWallet is a minimal Wallet implementation for unit tests.
type testWallet struct {
	address string
	signErr error
}

// NewTestWallet returns a Wallet suitable for handler-level testing.
// If signErr is non-nil, SignAndBroadcast will return it.
func NewTestWallet(address string, signErr error) Wallet {
	return &testWallet{address: address, signErr: signErr}
}

func (w *testWallet) SignData(_ []byte) (string, error) {
	return "", w.signErr
}

func (w *testWallet) SignAndBroadcast(_ context.Context, _ *grpc.ClientConn, _ []sdk.Msg) (*tx.GetTxResponse, error) {
	if w.signErr != nil {
		return nil, w.signErr
	}
	return &tx.GetTxResponse{}, nil
}

func (w *testWallet) Type() secretType { return UNKNOWN }
func (w *testWallet) Address() string  { return w.address }
func (w *testWallet) Export() ([]byte, error) {
	return []byte(`{}`), nil
}
