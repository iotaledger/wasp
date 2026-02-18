package iotatest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
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
		iotagraphql.GetCoinsRequest{Owner: *cryptolibSigner.Address()},
	)
	require.NoError(t, err)

	if len(getCoinsRes.Address.Coins.Nodes) > 1 {
		return
	}

	coins, err := client.GetCoinObjsForTargetAmount(
		context.Background(),
		*cryptolibSigner.Address(),
		splitBalance,
		iotagraphql.DefaultGasBudget,
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

	coinRef, err := coins[0].ObjectRef()
	require.NoError(t, err)

	txData := iotago.NewProgrammable(
		cryptolibSigner.Address(),
		txb.Finish(),
		[]*iotago.ObjectRef{coinRef},
		iotagraphql.DefaultGasBudget,
		parameterstest.L1Mock.Protocol.ReferenceGasPrice.Uint64(),
	)

	txnBytes, err := bcs.Marshal(&txData)
	require.NoError(t, err)

	result, err := client.SignAndExecuteTransaction(
		context.Background(),
		txnBytes,
		cryptolibSigner,
	)
	require.NoError(t, err)
	require.NotNil(t, result)
}
