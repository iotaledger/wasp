package iotatest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
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
		parameterstest.L1Mock.Protocol.ReferenceGasPrice.Uint64(),
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
