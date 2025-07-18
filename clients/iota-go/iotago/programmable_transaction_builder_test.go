package iotago_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"

	bcs "github.com/iotaledger/bcs-go"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestPTBMoveCall(t *testing.T) {
	t.Run(
		"access_multiple_return_values_from_move_func", func(t *testing.T) {
			client := l1starter.Instance().L1Client()
			sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

			_, packageID, err := client.PublishContract(
				context.Background(),
				sender,
				contracts.SDKVerify().Modules,
				contracts.SDKVerify().Dependencies,
				iotaclient.DefaultGasBudget,
				&iotajsonrpc.IotaTransactionBlockResponseOptions{ShowObjectChanges: true, ShowEffects: true},
			)
			require.NoError(t, err)

			coinPages, err := client.GetCoins(
				context.Background(), iotaclient.GetCoinsRequest{
					Owner: sender.Address(),
					Limit: 3,
				},
			)
			require.NoError(t, err)
			coins := iotajsonrpc.Coins(coinPages.Data)

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
				iotaclient.DefaultGasBudget,
				iotaclient.DefaultGasPrice,
			)
			txBytes, err := bcs.Marshal(&txData)
			require.NoError(t, err)
			simulate, err := client.DryRunTransaction(context.Background(), txBytes)
			require.NoError(t, err)

			require.Empty(t, simulate.Effects.Data.V1.Status.Error)
			require.True(t, simulate.Effects.Data.IsSuccess())
			require.Equal(t, coins[0].CoinObjectID, simulate.Effects.Data.V1.GasObject.Reference.ObjectID)
		},
	)
}

func TestPTBTransferObject(t *testing.T) {
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 2,
		},
	)
	require.NoError(t, err)
	coins := iotajsonrpc.Coins(coinPages.Data)
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
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	txn, err := client.TransferObject(
		context.Background(),
		iotaclient.TransferObjectRequest{
			Signer:    sender.Address(),
			Recipient: recipient.Address(),
			ObjectID:  transferCoin.CoinObjectID,
			Gas:       gasCoin.CoinObjectID,
			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}

func TestPTBTransferIota(t *testing.T) {
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 1,
		},
	)
	require.NoError(t, err)
	coin := iotajsonrpc.Coins(coinPages.Data)[0]
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
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytesBCS, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	txn, err := client.TransferIota(
		context.Background(),
		iotaclient.TransferIotaRequest{
			Signer:    sender.Address(),
			Recipient: recipient.Address(),
			ObjectID:  coin.CoinObjectID,
			Amount:    iotajsonrpc.NewBigInt(amount),
			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytesBCS, txBytesRemote)
}

func TestPTBPayAllIota(t *testing.T) {
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 3,
		},
	)
	require.NoError(t, err)
	coins := iotajsonrpc.Coins(coinPages.Data)

	// build with BCS
	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.PayAllIota(recipient.Address())
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		coins.CoinRefs(),
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	txn, err := client.PayAllIota(
		context.Background(),
		iotaclient.PayAllIotaRequest{
			Signer:     sender.Address(),
			Recipient:  recipient.Address(),
			InputCoins: coins.ObjectIDs(),
			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}

func TestPTBPayIota(t *testing.T) {
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient1 := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())
	recipient2 := iotatest.MakeSignerWithFunds(2, l1starter.Instance().FaucetURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
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
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())
	require.Equal(t, coin.CoinObjectID.String(), simulate.Effects.Data.V1.GasObject.Reference.ObjectID.String())

	// 1 for Mutated, 2 created (the 2 transfer in pay_iota pt),
	require.Len(t, simulate.ObjectChanges, 3)
	for _, change := range simulate.ObjectChanges {
		if change.Data.Mutated != nil {
			require.Equal(t, coin.CoinObjectID, &change.Data.Mutated.ObjectID)
		} else if change.Data.Created != nil {
			require.Contains(
				t,
				[]*iotago.Address{recipient1.Address(), recipient2.Address()},
				change.Data.Created.Owner.AddressOwner,
			)
		}
	}

	// build with remote rpc
	txn, err := client.PayIota(
		context.Background(),
		iotaclient.PayIotaRequest{
			Signer:     sender.Address(),
			InputCoins: []*iotago.ObjectID{coin.CoinObjectID},
			Recipients: []*iotago.Address{recipient1.Address(), recipient2.Address()},
			Amount:     []*iotajsonrpc.BigInt{iotajsonrpc.NewBigInt(123), iotajsonrpc.NewBigInt(456)},
			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}

func TestPTBPay(t *testing.T) {
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient1 := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())
	recipient2 := iotatest.MakeSignerWithFunds(2, l1starter.Instance().FaucetURL())

	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 3,
		},
	)
	require.NoError(t, err)
	coins := iotajsonrpc.Coins(coinPages.Data)
	gasCoin := coins[0] // save the 1st element for gas fee
	transferCoins := coins[1:]
	amounts := []uint64{123, 567}
	totalBal := coins.TotalBalance().Uint64()

	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.Pay(
		transferCoins.CoinRefs(),
		[]*iotago.Address{recipient1.Address(), recipient2.Address()},
		[]uint64{amounts[0], amounts[1]},
	)
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		[]*iotago.ObjectRef{
			gasCoin.Ref(),
		},
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())
	require.Equal(t, gasCoin.CoinObjectID.String(), simulate.Effects.Data.V1.GasObject.Reference.ObjectID.String())

	// 2 for Mutated (1 gas coin and 1 merged coin in pay pt), 2 created (the 2 transfer in pay pt),
	require.Len(t, simulate.ObjectChanges, 5)
	for _, change := range simulate.ObjectChanges {
		if change.Data.Mutated != nil {
			require.Contains(
				t,
				[]*iotago.ObjectID{gasCoin.CoinObjectID, transferCoins[0].CoinObjectID},
				&change.Data.Mutated.ObjectID,
			)
		} else if change.Data.Deleted != nil {
			require.Equal(t, transferCoins[1].CoinObjectID, &change.Data.Deleted.ObjectID)
		}
	}
	require.Len(t, simulate.BalanceChanges, 3)
	for _, balChange := range simulate.BalanceChanges {
		if balChange.Owner.AddressOwner == sender.Address() {
			require.Equal(t, totalBal-(amounts[0]+amounts[1]), balChange.Amount)
		} else if balChange.Owner.AddressOwner == recipient1.Address() {
			require.Equal(t, amounts[0], balChange.Amount)
		} else if balChange.Owner.AddressOwner == recipient2.Address() {
			require.Equal(t, amounts[1], balChange.Amount)
		}
	}

	// build with remote rpc
	txn, err := client.Pay(
		context.Background(),
		iotaclient.PayRequest{
			Signer:     sender.Address(),
			InputCoins: transferCoins.ObjectIDs(),
			Recipients: []*iotago.Address{recipient1.Address(), recipient2.Address()},
			Amount:     []*iotajsonrpc.BigInt{iotajsonrpc.NewBigInt(amounts[0]), iotajsonrpc.NewBigInt(amounts[1])},
			Gas:        gasCoin.CoinObjectID,
			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}
