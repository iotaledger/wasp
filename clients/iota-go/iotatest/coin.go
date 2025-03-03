package iotatest

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func EnsureCoinSplitWithBalance(
	t *testing.T,
	cryptolibSigner iotasigner.Signer,
	client clients.L1Client,
	splitBalance uint64,
) {
	getCoinsRes, err := client.GetCoins(
		context.Background(),
		iotaclient.GetCoinsRequest{Owner: cryptolibSigner.Address()},
	)
	require.NoError(t, err)

	if len(getCoinsRes.Data) > 1 {
		return
	}

	coins, err := client.GetCoinObjsForTargetAmount(
		context.Background(),
		cryptolibSigner.Address(),
		splitBalance,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)

	txb := iotago.NewProgrammableTransactionBuilder()

	splitCmd := txb.Command(
		iotago.Command{
			SplitCoins: &iotago.ProgrammableSplitCoins{
				Coin:    iotago.GetArgumentGasCoin(),
				Amounts: []iotago.Argument{txb.MustPure(splitBalance)},
			},
		},
	)
	txb.TransferArg(cryptolibSigner.Address(), splitCmd)

	txData := iotago.NewProgrammable(
		cryptolibSigner.Address(),
		txb.Finish(),
		[]*iotago.ObjectRef{coins[0].Ref()},
		iotaclient.DefaultGasBudget,
		parameters.L1Default.Protocol.ReferenceGasPrice.Uint64(),
	)

	txnBytes, err := bcs.Marshal(&txData)
	require.NoError(t, err)

	result, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			Signer:      cryptolibSigner,
			TxDataBytes: txnBytes,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, result)
}
