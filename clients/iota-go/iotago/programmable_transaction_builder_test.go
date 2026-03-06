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
	t.Skip()
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

func TestPTBTransferObject(t *testing.T) {
	t.Skip("Migrate to graphql")
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 2,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Address.Coins.Nodes)
	gasCoin := coins[0]
	transferCoin := coins[1]

	ptb := iotago.NewProgrammableTransactionBuilder()
	transferCoinRef, err := transferCoin.ObjectRef()
	require.NoError(t, err)
	err = ptb.TransferObject(recipient.Address(), transferCoinRef)
	require.NoError(t, err)
	pt := ptb.Finish()
	gasCoinRef, err := gasCoin.ObjectRef()
	require.NoError(t, err)
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		[]*iotago.ObjectRef{gasCoinRef},
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	transferCoinID := transferCoin.ObjectID()
	gasCoinID := gasCoin.ObjectID()
	txn, err := client.TransferObject(
		context.Background(),
		iotagraphql.TransferObjectRequest{
			Signer:    sender.Address(),
			Recipient: recipient.Address(),
			ObjectID:  &transferCoinID,
			Gas:       &gasCoinID,
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
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 1,
		},
	)
	require.NoError(t, err)
	coin := iotagraphql.Coins(coinPages.Address.Coins.Nodes)[0]
	amount := uint64(123)

	// build with BCS
	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.TransferIota(recipient.Address(), &amount)
	require.NoError(t, err)
	pt := ptb.Finish()
	coinRef, err := coin.ObjectRef()
	require.NoError(t, err)
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		[]*iotago.ObjectRef{coinRef},
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytesBCS, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	coinID := coin.ObjectID()
	txn, err := client.TransferIota(
		context.Background(),
		iotagraphql.TransferIotaRequest{
			Signer:    sender.Address(),
			Recipient: recipient.Address(),
			ObjectID:  &coinID,
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
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 3,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Address.Coins.Nodes)

	// build with BCS
	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.PayAllIota(recipient.Address())
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		lo.Must(coins.CoinRefs()),
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

	// build with remote rpc
	coinID := coin.ObjectID()
	txn, err := client.PayIota(
		context.Background(),
		iotagraphql.PayIotaRequest{
			Signer:     sender.Address(),
			InputCoins: []iotago.ObjectID{coinID},
			Recipients: []*iotago.Address{recipient1.Address(), recipient2.Address()},
			Amount:     []*iotagraphql.BigInt{iotagraphql.NewBigInt(123), iotagraphql.NewBigInt(456)},
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}
