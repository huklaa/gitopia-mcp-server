package signing

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMarshaler(t *testing.T) {
	m := NewMarshaler()
	require.NotNil(t, m)
}

func TestCalculateFee(t *testing.T) {
	coins, err := calculateFee(200000)
	require.NoError(t, err)
	require.Len(t, coins, 1)
	assert.Equal(t, "ulore", coins[0].Denom)
	// 200000 * 0.001 = 200
	assert.Equal(t, int64(200), coins[0].Amount.Int64())
}

func TestCalculateFee_Zero(t *testing.T) {
	coins, err := calculateFee(0)
	require.NoError(t, err)
	// Zero gas produces zero-amount coin which is pruned from the coin set
	assert.True(t, coins.IsZero())
}

func TestParseTxResponse_NilResponse(t *testing.T) {
	m := NewMarshaler()
	err := ParseTxResponse(m, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction response")
}

func TestParseTxResponse_EmptyData(t *testing.T) {
	m := NewMarshaler()
	resp := &tx.GetTxResponse{
		TxResponse: &sdk.TxResponse{Data: ""},
	}
	err := ParseTxResponse(m, resp, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction response")
}

func TestParseTxResponse_NilTxResponse(t *testing.T) {
	m := NewMarshaler()
	resp := &tx.GetTxResponse{TxResponse: nil}
	err := ParseTxResponse(m, resp, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction response")
}
