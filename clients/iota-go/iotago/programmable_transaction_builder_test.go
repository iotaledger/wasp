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

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestPTBMoveCall(t *testing.T) {
	if l1starter.IsSimulatorConfigured() {
		t.Skip("test does not work with simulator")
	}
	t.Run(
		"access_multiple_return_values_from_move_func", func(t *testing.T) {
			client := l1starter.Instance().L1Client()
			sender := iotatest.MakeSigner(0)
			require.NoError(t, client.RequestFundsFromFaucet(t.Context(), sender.Address()))

			senderAddr := sender.Address()
			txnBytes, err := client.Publish(
				context.Background(),
				iotagraphql.PublishRequest{
					Sender:          senderAddr,
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
					Owner: senderAddr,
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
				&senderAddr,
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
	if l1starter.IsSimulatorConfigured() {
		t.Skip("test does not work with simulator")
	}
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSigner(0)
	require.NoError(t, client.RequestFundsFromFaucet(t.Context(), sender.Address()))
	recipient1 := iotatest.MakeSigner(1)
	require.NoError(t, client.RequestFundsFromFaucet(t.Context(), recipient1.Address()))
	recipient2 := iotatest.MakeSigner(2)
	require.NoError(t, client.RequestFundsFromFaucet(t.Context(), recipient2.Address()))

	senderAddr3 := sender.Address()
	recipient1Addr := recipient1.Address()
	recipient2Addr := recipient2.Address()
	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: senderAddr3,
			Limit: 1,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Address.Coins.Nodes)
	coin := coins[0]

	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.PayIota(
		[]*iotago.Address{&recipient1Addr, &recipient2Addr},
		[]uint64{123, 456},
	)
	require.NoError(t, err)
	pt := ptb.Finish()

	tx := iotago.NewProgrammable(
		&senderAddr3,
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

	// build with remote rpc
	coinID := coin.ObjectID()
	txn, err := client.PayIota(
		context.Background(),
		iotagraphql.PayIotaRequest{
			Signer:     senderAddr3,
			InputCoins: []iotago.ObjectID{coinID},
			Recipients: []*iotago.Address{&recipient1Addr, &recipient2Addr},
			Amount:     []*iotagraphql.BigInt{iotagraphql.NewBigInt(123), iotagraphql.NewBigInt(456)},
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}
