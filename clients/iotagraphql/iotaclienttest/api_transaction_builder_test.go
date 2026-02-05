package iotaclienttest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestMergeCoins(t *testing.T) {
	t.Skip("FIXME create an account has at least two coin objects on chain")
	// api := l1starter.Instance().L1Client()
	// signer := testAddress
	// coins, err := api.GetCoins(context.Background(), iotagraphql.GetCoinsRequest{
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
	// 	iotagraphql.MergeCoinsRequest{
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
	t.Skip("TODO")
	// client := l1starter.Instance().L1Client()
	// signer := iotatest.MakeSignerWithFunds(0, iotaconn.TestnetFaucetURL)

	// sdkVerifyBytecode := contracts.SDKVerify()

	// txnBytes, err := client.Publish(
	// 	context.Background(),
	// 	iotagraphql.PublishRequest{
	// 		Sender:          signer.Address(),
	// 		CompiledModules: sdkVerifyBytecode.Modules,
	// 		Dependencies:    sdkVerifyBytecode.Dependencies,
	// 		GasBudget:       graphqltypes.NewBigInt(iotaclient.DefaultGasBudget),
	// 	},
	// )
	// require.NoError(t, err)
	// txnResponse, err := client.SignAndExecuteTransaction(
	// 	context.Background(),
	// 	&iotagraphql.SignAndExecuteTransactionRequest{
	// 		TxDataBytes: txnBytes.TxBytes,
	// 		Signer:      signer,
	// 		Options: &iotagraphql.IotaTransactionBlockResponseOptions{
	// 			ShowEffects:       true,
	// 			ShowObjectChanges: true,
	// 		},
	// 	},
	// )
	// require.NoError(t, err)
	// require.True(t, txnResponse.Effects.Data.IsSuccess())

	// packageID, err := txnResponse.GetPublishedPackageID()
	// require.NoError(t, err)

	// // test MoveCall with byte array input
	// input := []string{"haha", "gogo"}
	// txnBytes, err = client.MoveCall(
	// 	context.Background(),
	// 	iotaclient.MoveCallRequest{
	// 		Signer:    signer.Address(),
	// 		PackageID: packageID,
	// 		Module:    "sdk_verify",
	// 		Function:  "read_input_bytes_array",
	// 		TypeArgs:  []string{},
	// 		Arguments: []any{input},
	// 		GasBudget: graphqltypes.NewBigInt((iotaclient.DefaultGasBudget)),
	// 	},
	// )
	// require.NoError(t, err)
	// txnResponse, err = client.SignAndExecuteTransaction(
	// 	context.Background(),
	// 	&iotagraphql.SignAndExecuteTransactionRequest{
	// 		TxDataBytes: txnBytes.TxBytes,
	// 		Signer:      signer,
	// 		Options: &iotagraphql.IotaTransactionBlockResponseOptions{
	// 			ShowEffects: true,
	// 		},
	// 	},
	// )
	// require.NoError(t, err)
	// require.True(t, txnResponse.Effects.Data.IsSuccess())

	// queryEventsRes, err := client.QueryEvents(
	// 	context.Background(),
	// 	iotaclient.QueryEventsRequest{
	// 		Query: &graphqltypes.EventFilter{Transaction: &txnResponse.Digest},
	// 	},
	// )
	// require.NoError(t, err)
	// var queryEventsResMap map[string]any
	// err = json.Unmarshal(queryEventsRes.Data[0].ParsedJson, &queryEventsResMap)
	// require.NoError(t, err)
	// b, err := json.Marshal(queryEventsResMap["data"])
	// require.NoError(t, err)
	// var res [][]byte
	// err = json.Unmarshal(b, &res)
	// require.NoError(t, err)

	// require.Equal(t, []byte("haha"), res[0])
	// require.Equal(t, []byte("gogo"), res[1])
}

func TestPay(t *testing.T) {
	t.Skip("FIXME there is only 1 coin object, because there is only 1 coin object returned from faucet")
	// client := l1starter.Instance().L1Client()
	// signer := iotatest.MakeSignerWithFunds(0, iotaconn.TestnetFaucetURL)
	// recipient := iotatest.MakeSignerWithFunds(1, iotaconn.TestnetFaucetURL)

	// coins, err := client.GetCoins(
	// 	context.Background(), iotagraphql.GetCoinsRequest{
	// 		Owner: signer.Address(),
	// 		Limit: 10,
	// 	},
	// )
	// require.NoError(t, err)
	// limit := len(coins.Data) - 1 // need reserve a coin for gas

	// amount := uint64(123)
	// pickedCoins, err := graphqltypes.PickupCoins(
	// 	coins,
	// 	new(big.Int).SetUint64(amount),
	// 	iotaclient.DefaultGasBudget,
	// 	limit,
	// 	0,
	// )
	// require.NoError(t, err)

	// // Find a coin for gas that's not in the picked coins
	// var gasCoin *iotago.ObjectID
	// pickedCoinSet := make(map[string]bool)
	// for _, coinID := range pickedCoins.CoinIds() {
	// 	pickedCoinSet[coinID.String()] = true
	// }
	// for _, coin := range coins.Data {
	// 	if !pickedCoinSet[coin.CoinObjectID.String()] {
	// 		gasCoin = coin.CoinObjectID
	// 		break
	// 	}
	// }
	// require.NotNil(t, gasCoin, "should have a coin available for gas")

	// txn, err := client.Pay(
	// 	context.Background(),
	// 	iotagraphql.PayRequest{
	// 		Signer:     signer.Address(),
	// 		InputCoins: pickedCoins.CoinIds(),
	// 		Recipients: []*iotago.Address{recipient.Address()},
	// 		Amount:     []*graphqltypes.BigInt{graphqltypes.NewBigInt(amount)},
	// 		Gas:        gasCoin,
	// 		GasBudget:  graphqltypes.NewBigInt(iotaclient.DefaultGasBudget),
	// 	},
	// )
	// require.NoError(t, err)

	// simulate, err := client.DryRunTransaction(context.Background(), iotagraphql.DryRunTransactionRequest{
	// 	TxDataBytes: txn.TxBytes,
	// })
	// require.NoError(t, err)
	// require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	// require.True(t, simulate.Effects.Data.IsSuccess())

	// // TODO: Enable balance changes validation once GraphQL dry run supports it
	// // require.Len(t, simulate.BalanceChanges, 2)
	// // recipientAmount := strconv.FormatUint(amount, 10)
	// // signerAmount := strconv.FormatUint(totalBal-amount, 10)
	// // for _, balChange := range simulate.BalanceChanges {
	// // 	if balChange.Owner.AddressOwner == recipient.Address() {
	// // 		require.Equal(t, recipientAmount, balChange.Amount)
	// // 	} else if balChange.Owner.AddressOwner == signer.Address() {
	// // 		require.Equal(t, signerAmount, balChange.Amount)
	// // 	}
	// // }
}

func TestPayAllIota(t *testing.T) {
	t.Skip("FIXME there is only 1 coin object, because there is only 1 coin object returned from faucet")
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.TestnetFaucetURL)
	recipient := iotatest.MakeSignerWithFunds(1, iotaconn.TestnetFaucetURL)

	limit := int(3)
	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Data)
	// assume the account holds more than 'limit' amount Iota token objects
	require.Len(t, coinPages.Data, 3)

	txn, err := client.PayAllIota(
		context.Background(),
		iotagraphql.PayAllIotaRequest{
			Signer:     signer.Address(),
			Recipient:  recipient.Address(),
			InputCoins: coins.ObjectIDs(),
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), iotagraphql.DryRunTransactionRequest{
		TxDataBytes: txn.TxBytes,
	})
	require.NoError(t, err)
	require.Empty(t, simulate.DryRunTransactionBlock.Transaction.Effects.Errors)
	require.True(t, simulate.DryRunTransactionBlock.Transaction.Effects.IsSuccess())

	// TODO: ObjectChanges assertions need migration to GraphQL response types
	// require.Len(t, simulate.DryRunTransactionBlock.Transaction.Effects.ObjectChanges.Nodes, limit)
}

func TestPayIota(t *testing.T) {
	t.Skip("TODO")
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.TestnetFaucetURL)
	recipient1 := iotatest.MakeSignerWithFunds(1, iotaconn.TestnetFaucetURL)
	recipient2 := iotatest.MakeSignerWithFunds(2, iotaconn.TestnetFaucetURL)

	limit := int(4)
	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Data)

	sentAmounts := []uint64{123, 456, 789}
	txn, err := client.PayIota(
		context.Background(),
		iotagraphql.PayIotaRequest{
			Signer:     signer.Address(),
			InputCoins: coins.ObjectIDs(),
			Recipients: []*iotago.Address{
				recipient1.Address(),
				recipient2.Address(),
				recipient2.Address(),
			},
			Amount: []*iotagraphql.BigInt{
				iotagraphql.NewBigInt(sentAmounts[0]), // to recipient1
				iotagraphql.NewBigInt(sentAmounts[1]), // to recipient2
				iotagraphql.NewBigInt(sentAmounts[2]), // to recipient2
			},
			GasBudget: iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), iotagraphql.DryRunTransactionRequest{
		TxDataBytes: txn.TxBytes,
	})
	require.NoError(t, err)
	require.Empty(t, simulate.DryRunTransactionBlock.Transaction.Effects.Errors)
	require.True(t, simulate.DryRunTransactionBlock.Transaction.Effects.IsSuccess())

	// Verify the sender at the transaction level
	require.Equal(t, *signer.Address(), simulate.DryRunTransactionBlock.Transaction.Sender.Address)

	// 3 stands for the three amounts (3 created IOTA objects) in unsafe_payIota API
	amountNum := uint(3)
	objectChanges := simulate.DryRunTransactionBlock.Transaction.Effects.ObjectChanges.Nodes
	require.Len(t, objectChanges, limit+int(amountNum))
	delObjNum := uint(0)
	createdObjNum := uint(0)
	for _, change := range objectChanges {
		if change.IdDeleted {
			delObjNum++
		} else if change.IdCreated {
			createdObjNum++
		} else {
			// mutated object - verify it's one of the input coins
			require.Contains(t, coins.ObjectIDVals(), change.Address)
		}
	}

	// all the input objects are merged into the first input object
	// except the first input object, all the other input objects are deleted
	require.Equal(t, limit-1, delObjNum)
	// 1 for recipient1, and 2 for recipient2
	require.Equal(t, amountNum, createdObjNum)
}

func TestPublish(t *testing.T) {
	client := l1starter.Instance().L1Client()
	// Use the faucet URL from the running node (LoadConfig() leaves it empty when using the local testnode).
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())
	time.Sleep(1 * time.Second) // FIXME tmp for graphql

	testcoinBytecode := contracts.Testcoin()

	txnBytes, err := client.Publish(
		context.Background(),
		iotagraphql.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: testcoinBytecode.Modules,
			Dependencies:    testcoinBytecode.Dependencies,
			GasBudget:       iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget * 5),
		},
	)
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotagraphql.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options: &iotagraphql.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.IsSuccess())

	// Verify that published package is returned correctly
	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)
	require.NotNil(t, packageID)
	t.Logf("Published package ID: %s", packageID)
}

func TestSplitCoin(t *testing.T) {
	t.Skip("TODO")
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.TestnetFaucetURL)

	limit := int(4)
	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Data)

	txn, err := client.SplitCoin(
		context.Background(),
		iotagraphql.SplitCoinRequest{
			Signer: signer.Address(),
			Coin:   coins[1].CoinObjectID,
			SplitAmounts: []*iotagraphql.BigInt{
				// assume coins[0] has more than the sum of the following splitAmounts
				iotagraphql.NewBigInt(2222),
				iotagraphql.NewBigInt(1111),
			},
			GasBudget: iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), iotagraphql.DryRunTransactionRequest{
		TxDataBytes: txn.TxBytes,
	})
	require.NoError(t, err)
	effects := simulate.DryRunTransactionBlock.Transaction.Effects
	require.Empty(t, effects.Errors)
	require.True(t, effects.IsSuccess())

	// 2 mutated and 2 created (split coins)
	require.Len(t, effects.ObjectChanges.Nodes, 4)
	require.Len(t, effects.BalanceChanges.Nodes, 1)
	require.Equal(t, effects.BalanceChanges.Nodes[0].Amount.Int64(), -effects.GasFee())
}

func TestTransferObject(t *testing.T) {
	t.Skip("TODO")
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.TestnetFaucetURL)
	recipient := iotatest.MakeSignerWithFunds(1, iotaconn.TestnetFaucetURL)

	limit := int(3)
	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	transferCoin := coinPages.Data[0]

	txn, err := client.TransferObject(
		context.Background(),
		iotagraphql.TransferObjectRequest{
			Signer:    signer.Address(),
			Recipient: recipient.Address(),
			ObjectID:  transferCoin.CoinObjectID,
			GasBudget: iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), iotagraphql.DryRunTransactionRequest{
		TxDataBytes: txn.TxBytes,
	})
	require.NoError(t, err)
	effects := simulate.DryRunTransactionBlock.Transaction.Effects
	require.Empty(t, effects.Errors)
	require.True(t, effects.IsSuccess())

	// one is transferred object, one is the gas object
	require.Len(t, effects.ObjectChanges.Nodes, 2)

	require.Len(t, effects.BalanceChanges.Nodes, 2)
}
