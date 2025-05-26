package iotaclienttest

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestBatchTransaction(t *testing.T) {
	t.Log("TestBatchTransaction TODO")
	// api := l1starter.Instance().L1Client()

	// txnBytes, err := api.BatchTransaction(context.Background(), signer, *coin1, *coin2, nil, 10000)
	// require.NoError(t, err)
	// dryRunTxn(t, api, txnBytes, M1Account(t))
}

func TestMergeCoins(t *testing.T) {
	t.Skip("FIXME create an account has at least two coin objects on chain")
	// api := l1starter.Instance().L1Client()
	// signer := testAddress
	// coins, err := api.GetCoins(context.Background(), iotaclient.GetCoinsRequest{
	// 	Owner: signer,
	// 	Limit: 10,
	// })
	// require.NoError(t, err)
	// require.True(t, len(coins.Data) >= 3)

	// coin1 := coins.Data[0]
	// coin2 := coins.Data[1]
	// coin3 := coins.Data[2] // gas coin

	// txn, err := api.MergeCoins(
	// 	context.Background(),
	// 	iotaclient.MergeCoinsRequest{
	// 		Signer:      signer,
	// 		PrimaryCoin: coin1.CoinObjectID,
	// 		CoinToMerge: coin2.CoinObjectID,
	// 		Gas:         coin3.CoinObjectID,
	// 		GasBudget:   coin3.Balance,
	// 	},
	// )
	// require.NoError(t, err)

	// dryRunTxn(t, api, txn.TxBytes, true)
}

func TestMoveCall(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

	sdkVerifyBytecode := contracts.SDKVerify()

	txnBytes, err := client.Publish(
		context.Background(),
		iotaclient.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: sdkVerifyBytecode.Modules,
			Dependencies:    sdkVerifyBytecode.Dependencies,
			GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)
	execMode := iotaclient.IotaTransactionBlockBuilderModeCommit

	// test MoveCall with byte array input
	input := []string{"haha", "gogo"}
	txnBytes, err = client.MoveCall(
		context.Background(),
		iotaclient.MoveCallRequest{
			Signer:        signer.Address(),
			PackageID:     packageID,
			Module:        "sdk_verify",
			Function:      "read_input_bytes_array",
			TypeArgs:      []string{},
			Arguments:     []any{input},
			GasBudget:     iotajsonrpc.NewBigInt((iotaclient.DefaultGasBudget)),
			ExecutionMode: &execMode,
		},
	)
	require.NoError(t, err)
	txnResponse, err = client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects: true,
			},
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	queryEventsRes, err := client.QueryEvents(
		context.Background(),
		iotaclient.QueryEventsRequest{
			Query: &iotajsonrpc.EventFilter{Transaction: &txnResponse.Digest},
		},
	)
	require.NoError(t, err)
	queryEventsResMap := queryEventsRes.Data[0].ParsedJson.(map[string]interface{})
	b, err := json.Marshal(queryEventsResMap["data"])
	require.NoError(t, err)
	var res [][]byte
	err = json.Unmarshal(b, &res)
	require.NoError(t, err)

	require.Equal(t, []byte("haha"), res[0])
	require.Equal(t, []byte("gogo"), res[1])
}

func TestPay(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)
	limit := len(coins.Data) - 1 // need reserve a coin for gas
	totalBal := iotajsonrpc.Coins(coins.Data).TotalBalance().Uint64()

	amount := uint64(123)
	pickedCoins, err := iotajsonrpc.PickupCoins(
		coins,
		new(big.Int).SetUint64(amount),
		iotaclient.DefaultGasBudget,
		limit,
		0,
	)
	require.NoError(t, err)

	txn, err := client.Pay(
		context.Background(),
		iotaclient.PayRequest{
			Signer:     signer.Address(),
			InputCoins: pickedCoins.CoinIds(),
			Recipients: []*iotago.Address{recipient.Address()},
			Amount:     []*iotajsonrpc.BigInt{iotajsonrpc.NewBigInt(amount)},
			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	require.Len(t, simulate.BalanceChanges, 2)
	for _, balChange := range simulate.BalanceChanges {
		if balChange.Owner.AddressOwner == recipient.Address() {
			require.Equal(t, amount, balChange.Amount)
		} else if balChange.Owner.AddressOwner == signer.Address() {
			require.Equal(t, totalBal-amount, balChange.Amount)
		}
	}
}

func TestPayAllIota(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	limit := uint(3)
	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := iotajsonrpc.Coins(coinPages.Data)
	// assume the account holds more than 'limit' amount Iota token objects
	require.Len(t, coinPages.Data, 3)
	totalBal := coins.TotalBalance()

	txn, err := client.PayAllIota(
		context.Background(),
		iotaclient.PayAllIotaRequest{
			Signer:     signer.Address(),
			Recipient:  recipient.Address(),
			InputCoins: coins.ObjectIDs(),
			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	require.Len(t, simulate.ObjectChanges, int(limit))
	delObjNum := uint(0)
	for _, change := range simulate.ObjectChanges {
		if change.Data.Mutated != nil {
			require.Equal(t, *signer.Address(), change.Data.Mutated.Sender)
			require.Contains(t, coins.ObjectIDVals(), change.Data.Mutated.ObjectID)
		} else if change.Data.Deleted != nil {
			delObjNum += 1
		}
	}
	// all the input objects are merged into the first input object
	// except the first input object, all the other input objects are deleted
	require.Equal(t, limit-1, delObjNum)

	// one output balance and one input balance
	require.Len(t, simulate.BalanceChanges, 2)
	for _, balChange := range simulate.BalanceChanges {
		if balChange.Owner.AddressOwner == signer.Address() {
			require.Equal(t, totalBal.Neg(totalBal), balChange.Amount)
		} else if balChange.Owner.AddressOwner == recipient.Address() {
			require.Equal(t, totalBal, balChange.Amount)
		}
	}
}

func TestPayIota(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient1 := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())
	recipient2 := iotatest.MakeSignerWithFunds(2, l1starter.Instance().FaucetURL())

	limit := uint(4)
	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := iotajsonrpc.Coins(coinPages.Data)

	sentAmounts := []uint64{123, 456, 789}
	txn, err := client.PayIota(
		context.Background(),
		iotaclient.PayIotaRequest{
			Signer:     signer.Address(),
			InputCoins: coins.ObjectIDs(),
			Recipients: []*iotago.Address{
				recipient1.Address(),
				recipient2.Address(),
				recipient2.Address(),
			},
			Amount: []*iotajsonrpc.BigInt{
				iotajsonrpc.NewBigInt(sentAmounts[0]), // to recipient1
				iotajsonrpc.NewBigInt(sentAmounts[1]), // to recipient2
				iotajsonrpc.NewBigInt(sentAmounts[2]), // to recipient2
			},
			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	// 3 stands for the three amounts (3 crated IOTA objects) in unsafe_payIota API
	amountNum := uint(3)
	require.Len(t, simulate.ObjectChanges, int(limit)+int(amountNum))
	delObjNum := uint(0)
	createdObjNum := uint(0)
	for _, change := range simulate.ObjectChanges {
		if change.Data.Mutated != nil {
			require.Equal(t, *signer.Address(), change.Data.Mutated.Sender)
			require.Contains(t, coins.ObjectIDVals(), change.Data.Mutated.ObjectID)
		} else if change.Data.Created != nil {
			createdObjNum += 1
			require.Equal(t, *signer.Address(), change.Data.Created.Sender)
		} else if change.Data.Deleted != nil {
			delObjNum += 1
		}
	}

	// all the input objects are merged into the first input object
	// except the first input object, all the other input objects are deleted
	require.Equal(t, limit-1, delObjNum)
	// 1 for recipient1, and 2 for recipient2
	require.Equal(t, amountNum, createdObjNum)

	// one output balance and one input balance for recipient1 and one input balance for recipient2
	require.Len(t, simulate.BalanceChanges, 3)
	for _, balChange := range simulate.BalanceChanges {
		if balChange.Owner.AddressOwner == signer.Address() {
			require.Equal(t, coins.TotalBalance().Neg(coins.TotalBalance()), balChange.Amount)
		} else if balChange.Owner.AddressOwner == recipient1.Address() {
			require.Equal(t, sentAmounts[0], balChange.Amount)
		} else if balChange.Owner.AddressOwner == recipient2.Address() {
			require.Equal(t, sentAmounts[1]+sentAmounts[2], balChange.Amount)
		}
	}
}

func TestPublish(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

	testcoinBytecode := contracts.Testcoin()

	txnBytes, err := client.Publish(
		context.Background(),
		iotaclient.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: testcoinBytecode.Modules,
			Dependencies:    testcoinBytecode.Dependencies,
			GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 5),
		},
	)
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects: true,
			},
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())
}

func TestSplitCoin(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

	limit := uint(4)
	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := iotajsonrpc.Coins(coinPages.Data)

	txn, err := client.SplitCoin(
		context.Background(),
		iotaclient.SplitCoinRequest{
			Signer: signer.Address(),
			Coin:   coins[1].CoinObjectID,
			SplitAmounts: []*iotajsonrpc.BigInt{
				// assume coins[0] has more than the sum of the following splitAmounts
				iotajsonrpc.NewBigInt(2222),
				iotajsonrpc.NewBigInt(1111),
			},
			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	// 2 mutated and 2 created (split coins)
	require.Len(t, simulate.ObjectChanges, 4)
	require.Len(t, simulate.BalanceChanges, 1)
	amt, _ := strconv.ParseInt(simulate.BalanceChanges[0].Amount, 10, 64)
	require.Equal(t, amt, -simulate.Effects.Data.GasFee())
}

func TestSplitCoinEqual(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

	limit := uint(4)
	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := iotajsonrpc.Coins(coinPages.Data)

	splitShares := uint64(3)
	txn, err := client.SplitCoinEqual(
		context.Background(),
		iotaclient.SplitCoinEqualRequest{
			Signer:     signer.Address(),
			Coin:       coins[0].CoinObjectID,
			SplitCount: iotajsonrpc.NewBigInt(splitShares),
			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	// 1 mutated and 3 created (split coins)
	require.Len(t, simulate.ObjectChanges, 1+int(splitShares))
	require.Len(t, simulate.BalanceChanges, 1)
	amt, _ := strconv.ParseInt(simulate.BalanceChanges[0].Amount, 10, 64)
	require.Equal(t, amt, -simulate.Effects.Data.GasFee())
}

func TestTransferObject(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	limit := uint(3)
	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	transferCoin := coinPages.Data[0]

	txn, err := client.TransferObject(
		context.Background(),
		iotaclient.TransferObjectRequest{
			Signer:    signer.Address(),
			Recipient: recipient.Address(),
			ObjectID:  transferCoin.CoinObjectID,
			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	// one is transferred object, one is the gas object
	require.Len(t, simulate.ObjectChanges, 2)

	require.Len(t, simulate.BalanceChanges, 2)
}

func TestTransferIota(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	recipient := iotatest.MakeSignerWithFunds(1, l1starter.Instance().FaucetURL())

	limit := uint(3)
	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	transferCoin := coinPages.Data[0]

	txn, err := client.TransferIota(
		context.Background(),
		iotaclient.TransferIotaRequest{
			Signer:    signer.Address(),
			Recipient: recipient.Address(),
			ObjectID:  transferCoin.CoinObjectID,
			Amount:    iotajsonrpc.NewBigInt(3),
			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	// one is transferred object, one is the gas object
	require.Len(t, simulate.ObjectChanges, 2)
	for _, change := range simulate.ObjectChanges {
		if change.Data.Mutated != nil {
			require.Equal(t, *transferCoin.CoinObjectID, change.Data.Mutated.ObjectID)
			require.Equal(t, signer.Address(), change.Data.Mutated.Owner.AddressOwner)
		} else if change.Data.Created != nil {
			require.Equal(t, recipient.Address(), change.Data.Created.Owner.AddressOwner)
		}
	}

	require.Len(t, simulate.BalanceChanges, 2)
}
