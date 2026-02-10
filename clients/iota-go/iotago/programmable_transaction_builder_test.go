package iotago_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"

	bcs "github.com/iotaledger/bcs-go"
)

func TestPTBMoveCall(t *testing.T) {
	t.Skip()
	t.Run(
		"access_multiple_return_values_from_move_func", func(t *testing.T) {
			client := l1starter.Instance().L1Client()
			sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

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
				&iotagraphql.SignAndExecuteTransactionRequest{
					TxDataBytes: txnBytes.TxBytes,
					Signer:      sender,
					Options: &iotagraphql.IotaTransactionBlockResponseOptions{
						ShowEffects:       true,
						ShowObjectChanges: true,
					},
				},
			)
			require.NoError(t, err)
			require.True(t, txnResponse.ExecuteTransactionBlock.IsSuccess())

			packageID, err := txnResponse.GetPublishedPackageID()
			require.NoError(t, err)

			coinPages, err := client.GetCoins(
				context.Background(), iotagraphql.GetCoinsRequest{
					Owner: sender.Address(),
					Limit: 3,
				},
			)
			require.NoError(t, err)
			coins := iotagraphql.Coins(coinPages.Data)

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
			txData := iotago.NewProgrammable(
				sender.Address(),
				pt,
				[]*iotago.ObjectRef{coins[0].Ref()},
				iotagraphql.DefaultGasBudget,
				iotagraphql.DefaultGasPrice,
			)
			txBytes, err := bcs.Marshal(&txData)
			require.NoError(t, err)
			simulate, err := client.DryRunTransaction(
				context.Background(), iotagraphql.DryRunTransactionRequest{
					TxDataBytes: txBytes,
				},
			)
			require.NoError(t, err)

			require.Empty(t, simulate.DryRunTransactionBlock.Transaction.Effects.Errors)
			require.Equal(
				t,
				iotagraphql.ExecutionStatusSuccess,
				simulate.DryRunTransactionBlock.Transaction.Effects.Status,
			)
		},
	)
}

func TestPTBTransferObject(t *testing.T) {
	t.Skip("Migrate to graphql")
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 2,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Data)
	gasCoin := coins[0]
	transferCoin := coins[1]

	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.TransferObject(recipient.Address(), transferCoin.Ref())
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		[]*iotago.ObjectRef{gasCoin.Ref()},
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	txn, err := client.TransferObject(
		context.Background(),
		iotagraphql.TransferObjectRequest{
			Signer:    sender.Address(),
			Recipient: recipient.Address(),
			ObjectID:  transferCoin.CoinObjectID,
			Gas:       gasCoin.CoinObjectID,
			GasBudget: iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}

func TestPTBTransferIota(t *testing.T) {
	t.Skip("Migrate to graphql")
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 1,
		},
	)
	require.NoError(t, err)
	coin := iotagraphql.Coins(coinPages.Data)[0]
	amount := uint64(123)

	// build with BCS
	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.TransferIota(recipient.Address(), &amount)
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		[]*iotago.ObjectRef{coin.Ref()},
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytesBCS, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	txn, err := client.TransferIota(
		context.Background(),
		iotagraphql.TransferIotaRequest{
			Signer:    sender.Address(),
			Recipient: recipient.Address(),
			ObjectID:  coin.CoinObjectID,
			Amount:    iotagraphql.NewBigInt(amount),
			GasBudget: iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytesBCS, txBytesRemote)
}

func TestPTBPayAllIota(t *testing.T) {
	t.Skip("Migrate to graphql")
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 3,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Data)

	// build with BCS
	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.PayAllIota(recipient.Address())
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		coins.CoinRefs(),
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	txn, err := client.PayAllIota(
		context.Background(),
		iotagraphql.PayAllIotaRequest{
			Signer:     sender.Address(),
			Recipient:  recipient.Address(),
			InputCoins: coins.ObjectIDs(),
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}

func TestPTBPayIota(t *testing.T) {
	l1starter.TestLocal()
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient1 := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())
	recipient2 := iotatest.MakeSignerWithFunds(2, l1starter.Instance().FaucetURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 1,
		},
	)
	require.NoError(t, err)
	coin := coinPages.Data[0]

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
			coin.Ref(),
		},
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(
		context.Background(), iotagraphql.DryRunTransactionRequest{
			TxDataBytes: txBytes,
		},
	)
	require.NoError(t, err)
	effects := simulate.DryRunTransactionBlock.Transaction.Effects
	require.Empty(t, effects.Errors)
	require.Equal(t, iotagraphql.ExecutionStatusSuccess, effects.Status)

	// 1 for Mutated, 2 created (the 2 transfer in pay_iota pt).
	// Note: GasObject and OutputState Object references are not resolved by the
	// GraphQL server for dry run transactions (only scalar fields are available).
	changes := effects.ObjectChanges.Nodes
	require.Len(t, changes, 3)
	for _, change := range changes {
		if !change.IdCreated && !change.IdDeleted {
			// Mutated - this is the gas coin
			require.Equal(t, *coin.CoinObjectID, change.Address)
		}
	}

	// build with remote rpc
	txn, err := client.PayIota(
		context.Background(),
		iotagraphql.PayIotaRequest{
			Signer:     sender.Address(),
			InputCoins: []*iotago.ObjectID{coin.CoinObjectID},
			Recipients: []*iotago.Address{recipient1.Address(), recipient2.Address()},
			Amount:     []*iotagraphql.BigInt{iotagraphql.NewBigInt(123), iotagraphql.NewBigInt(456)},
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}
