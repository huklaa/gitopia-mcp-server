package signing

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	cosmosgroup "github.com/cosmos/cosmos-sdk/x/group"
	"github.com/gitopia/gitopia-mcp-server/internal/signing/config"
	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
	"google.golang.org/grpc"
)

func NewMarshaler() codec.ProtoCodecMarshaler {
	interfaceRegistry := types.NewInterfaceRegistry()
	gitopiatypes.RegisterInterfaces(interfaceRegistry)
	cosmosgroup.RegisterInterfaces(interfaceRegistry)
	return codec.NewProtoCodec(interfaceRegistry)
}

// ParseTxResponse parses the data from a transaction response and unmarshals the first message response into the provided type.
func ParseTxResponse(marshaler codec.ProtoCodecMarshaler, txResponse *tx.GetTxResponse, resp codec.ProtoMarshaler) error {
	if txResponse == nil || txResponse.TxResponse == nil || txResponse.TxResponse.Data == "" {
		return fmt.Errorf("invalid transaction response")
	}

	// Parse the transaction response data
	hexString, err := hex.DecodeString(txResponse.TxResponse.Data)
	if err != nil {
		return fmt.Errorf("failed to decode hex string: %w", err)
	}

	// First unmarshal into TxMsgData to get the array of message responses
	var txMsgData sdk.TxMsgData
	err = marshaler.Unmarshal(hexString, &txMsgData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal TxMsgData: %w", err)
	}

	// Check if we have at least one message response
	if len(txMsgData.MsgResponses) == 0 {
		return fmt.Errorf("no message responses found in transaction")
	}

	// Get the first message response (assuming it's the one we want)
	msgResponse := txMsgData.MsgResponses[0]

	// Unmarshal the specific message response
	err = marshaler.Unmarshal(msgResponse.Value, resp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message response: %w", err)
	}

	return nil
}

const (
	GAS_ADJUSTMENT = 1.8
)

func calculateGas(ctx context.Context, cc *grpc.ClientConn, txClient tx.ServiceClient, txCfg client.TxConfig, txBuilder client.TxBuilder) (uint64, error) {
	txBytes, err := txCfg.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return 0, err
	}

	simRes, err := txClient.Simulate(ctx, &tx.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return 0, err
	}

	gas := uint64(GAS_ADJUSTMENT * float64(simRes.GasInfo.GasUsed))

	return gas, nil
}

func calculateFee(gas uint64) (sdk.Coins, error) {
	gasPrice, err := sdk.ParseDecCoin(config.GasPrices)
	if err != nil {
		return nil, err
	}
	fee := float64(gas) * float64(gasPrice.Amount.MustFloat64())
	fee = math.Ceil(fee)

	return sdk.NewCoins(sdk.NewCoin(gasPrice.Denom, sdk.NewInt(int64(fee)))), nil
}
