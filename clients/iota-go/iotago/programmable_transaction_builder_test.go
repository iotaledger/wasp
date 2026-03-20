package iotago_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"

	bcs "github.com/iotaledger/bcs-go"
)

func TestPTBMoveCall(t *testing.T) {
	l1starter.TestLocal()
	t.Run(
		"access_multiple_return_values_from_move_func", func(t *testing.T) {
			client := l1starter.Instance().L1Client()
			sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())

			txnBytes, err := client.Publish(
				context.Background(),
				iotagraphql.PublishRequest{
					Sender:          sender.Address(),
					CompiledModules: contracts.SDKVerify().Modules,
					Dependencies:    contracts.SDKVerify().Dependencies,
					GasBudget:       iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
				},
			)
			require.NoError(t, err)
			txnResponse, err := client.SignAndExecuteTransaction(
				context.Background(),
				txnBytes.TxBytes,
				sender,
			)
			require.NoError(t, err)
			require.True(t, txnResponse.IsSuccess())

			packageID, err := txnResponse.GetPublishedPackageID()
			require.NoError(t, err)

			coinPages, err := client.GetCoins(
				context.Background(), iotagraphql.GetCoinsRequest{
					Owner: sender.Address(),
					Limit: 3,
				},
			)
			require.NoError(t, err)
			coins := iotagraphql.Coins(coinPages.Address.Coins.Nodes)

			ptb := iotago.NewProgrammableTransactionBuilder()
			require.NoError(t, err)

			ptb.Command(
				iotago.Command{
					MoveCall: &iotago.ProgrammableMoveCall{
						Package:       packageID,
						Module:        "sdk_verify",
						Function:      "ret_two_1",
						TypeArguments: []iotago.TypeTag{},
						Arguments:     []iotago.Argument{},
					},
				},
			)
			ptb.Command(
				iotago.Command{
					MoveCall: &iotago.ProgrammableMoveCall{
						Package:       packageID,
						Module:        "sdk_verify",
						Function:      "ret_two_2",
						TypeArguments: []iotago.TypeTag{},
						Arguments: []iotago.Argument{
							{NestedResult: &iotago.NestedResult{Cmd: 0, Result: 1}},
							{NestedResult: &iotago.NestedResult{Cmd: 0, Result: 0}},
						},
					},
				},
			)
			pt := ptb.Finish()
			coinRef, err := coins[0].ObjectRef()
			require.NoError(t, err)
			txData := iotago.NewProgrammable(
				sender.Address(),
				pt,
				[]*iotago.ObjectRef{coinRef},
				iotagraphql.DefaultGasBudget,
				iotagraphql.DefaultGasPrice,
			)
			txBytes, err := bcs.Marshal(&txData)
			require.NoError(t, err)
			simulate, err := client.DryRunTransaction(
				context.Background(), txBytes,
			)
			require.NoError(t, err)

			require.True(t, simulate.DryRunTransactionBlock.Transaction.Effects.IsSuccess())
		},
	)
}

func TestPTBPayIota(t *testing.T) {
	l1starter.TestLocal()
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())
	recipient1 := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())
	recipient2 := iotatest.MakeSignerWithFunds(2, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 1,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Address.Coins.Nodes)
	coin := coins[0]

	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.PayIota(
		[]*iotago.Address{recipient1.Address(), recipient2.Address()},
		[]uint64{123, 456},
	)
	require.NoError(t, err)
	pt := ptb.Finish()

	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		[]*iotago.ObjectRef{
			lo.Must(coin.ObjectRef()),
		},
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(
		context.Background(), txBytes,
	)
	require.NoError(t, err)
	require.True(t, simulate.DryRunTransactionBlock.Transaction.Effects.IsSuccess())
}
