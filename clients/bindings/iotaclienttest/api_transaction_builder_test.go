package iotaclienttest

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/bindings"
	"github.com/iotaledger/wasp/v2/clients/bindings/iota_sdk_ffi"
	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
)

func TestBatchTransaction(t *testing.T) {
	t.Log("TestBatchTransaction TODO")
	// api := bindingClient

	// txnBytes, err := api.BatchTransaction(context.Background(), signer, *coin1, *coin2, nil, 10000)
	// require.NoError(t, err)
	// dryRunTxn(t, api, txnBytes, M1Account(t))
}

func TestMergeCoins(t *testing.T) {
	t.Skip("FIXME create an account has at least two coin objects on chain")
	// api := bindingClient
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
	client := bindings.NewBindingClient(iotaconn.DevnetEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.DevnetFaucetURL)
	// err := iotaclient.RequestFundsFromFaucet(context.TODO(), signer.Address(), iotaconn.DevnetFaucetURL)
	// require.NoError(t, err)
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
	if !txnResponse.Effects.Data.IsSuccess() {
		t.Fatalf("Publish transaction failed with error: %s", txnResponse.Effects.Data.V1.Status.Error)
	}

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)
	time.Sleep(5 * time.Second) // wait a little for gas object

	// test MoveCall with byte array input
	input := []string{"haha", "gogo"}
	txnBytes, err = client.MoveCall(
		context.Background(),
		iotaclient.MoveCallRequest{
			Signer:    signer.Address(),
			PackageID: packageID,
			Module:    "sdk_verify",
			Function:  "read_input_bytes_array",
			TypeArgs:  []string{},
			Arguments: []any{input},
			GasBudget: iotajsonrpc.NewBigInt((iotaclient.DefaultGasBudget)),
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

	// Wait for events to be indexed in GraphQL (GraphQL has indexing delay vs JSON-RPC)
	time.Sleep(2 * time.Second)

	queryEventsRes, err := client.QueryEvents(
		context.Background(),
		iotaclient.QueryEventsRequest{
			Query: &iotajsonrpc.EventFilter{Transaction: &txnResponse.Digest},
		},
	)
	require.NoError(t, err)
	var queryEventsResMap map[string]any
	err = json.Unmarshal(queryEventsRes.Data[0].ParsedJson, &queryEventsResMap)
	require.NoError(t, err)
	b, err := json.Marshal(queryEventsResMap["data"])
	require.NoError(t, err)
	var res [][]byte
	err = json.Unmarshal(b, &res)
	require.NoError(t, err)

	require.Equal(t, []byte("haha"), res[0])
	require.Equal(t, []byte("gogo"), res[1])
}

func TestPay(t *testing.T) {
	client := bindings.NewBindingClient(iotaconn.DevnetEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.DevnetFaucetURL)
	recipient := iotatest.MakeSignerWithFunds(1, iotaconn.DevnetFaucetURL)

	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)
	limit := len(coins.Data) - 1 // need reserve a coin for gas

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

	// Use FFI DryRun instead of JSON-RPC
	dryRunResult, err := client.DryRunTransactionRaw(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.NotNil(t, dryRunResult.Effects)

	effects := (*dryRunResult.Effects).AsV1()

	// Check transaction succeeded
	_, isSuccess := effects.Status.(iota_sdk_ffi.ExecutionStatusSuccess)
	require.True(t, isSuccess, "transaction should succeed")

	// Verify objects were changed (coins were transferred)
	require.NotEmpty(t, effects.ChangedObjects)

	// Count objects owned by sender and recipient after transaction
	senderOwned := 0
	recipientOwned := 0

	for _, changedObj := range effects.ChangedObjects {
		switch output := changedObj.OutputState.(type) {
		case iota_sdk_ffi.ObjectOutObjectWrite:
			if output.Owner.IsAddress() {
				ownerAddr, err := iotago.AddressFromHex(output.Owner.AsAddress().ToHex())
				require.NoError(t, err)
				if ownerAddr == signer.Address() {
					senderOwned++
				} else if ownerAddr == recipient.Address() {
					recipientOwned++
				}
			}
		}
	}

	// Recipient should own at least one object (the transferred coin)
	require.Greater(t, recipientOwned, 0, "recipient should own transferred coins")

	// Decode BCS coin data to verify balances and compute balance changes
	type CoinData struct {
		ID      [32]byte `bcs:"id"`      // ObjectID
		Balance uint64   `bcs:"balance"` // Balance value
	}

	// Map to track balance changes per owner address
	balanceChanges := make(map[string]int64) // address -> balance delta

	for _, changedObj := range effects.ChangedObjects {
		// Only process coin objects
		if changedObj.OutputState == nil {
			continue
		}

		objectWrite, ok := changedObj.OutputState.(iota_sdk_ffi.ObjectOutObjectWrite)
		if !ok || !objectWrite.Owner.IsAddress() {
			continue
		}

		ownerAddr, err := iotago.AddressFromHex(objectWrite.Owner.AsAddress().ToHex())
		require.NoError(t, err)
		ownerKey := ownerAddr.String()

		// Find the coin balance from DryRun results by ObjectID
		objectIDBytes := changedObj.ObjectId.ToBytes()

		for _, effect := range dryRunResult.Results {
			// Check MutatedReferences for this object
			for _, mutation := range effect.MutatedReferences {
				if mutation.TypeTag.String() != "0x2::coin::Coin<0x2::iota::IOTA>" {
					continue
				}

				coinData, err := bcs.Unmarshal[CoinData](mutation.Bcs)
				require.NoError(t, err)

				// Match by ObjectID
				if string(coinData.ID[:]) == string(objectIDBytes) {
					// For simplicity, we track the final balance as the "change"
					// (proper change would require querying before-state)
					balanceChanges[ownerKey] += int64(coinData.Balance)
				}
			}

			// Check ReturnValues for new coins
			for _, ret := range effect.ReturnValues {
				if ret.TypeTag.String() != "0x2::coin::Coin<0x2::iota::IOTA>" {
					continue
				}

				coinData, err := bcs.Unmarshal[CoinData](ret.Bcs)
				require.NoError(t, err)

				if string(coinData.ID[:]) == string(objectIDBytes) {
					balanceChanges[ownerKey] += int64(coinData.Balance)
				}
			}
		}
	}

	// Verify recipient received the amount
	recipientKey := recipient.Address().String()
	require.Contains(t, balanceChanges, recipientKey, "recipient should have balance changes")
	require.GreaterOrEqual(t, balanceChanges[recipientKey], int64(amount),
		fmt.Sprintf("recipient should have at least %d in balance", amount))
}

// func TestPayAllIota(t *testing.T) {
// 	client := bindings.NewBindingClient(iotaconn.DevnetEndpointURL)
// 	signer := iotatest.MakeSignerWithFunds(0, iotaconn.DevnetFaucetURL)
// 	recipient := iotatest.MakeSignerWithFunds(1, iotaconn.DevnetFaucetURL)

// 	limit := uint(3)
// 	coinPages, err := client.GetCoins(
// 		context.Background(), iotaclient.GetCoinsRequest{
// 			Owner: signer.Address(),
// 			Limit: limit,
// 		},
// 	)
// 	require.NoError(t, err)
// 	coins := iotajsonrpc.Coins(coinPages.Data)
// 	// assume the account holds more than 'limit' amount Iota token objects
// 	require.Len(t, coinPages.Data, 3)
// 	totalBal := coins.TotalBalance()

// 	txn, err := client.PayAllIota(
// 		context.Background(),
// 		iotaclient.PayAllIotaRequest{
// 			Signer:     signer.Address(),
// 			Recipient:  recipient.Address(),
// 			InputCoins: coins.ObjectIDs(),
// 			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
// 		},
// 	)
// 	require.NoError(t, err)

// 	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
// 	require.NoError(t, err)
// 	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
// 	require.True(t, simulate.Effects.Data.IsSuccess())

// 	require.Len(t, simulate.ObjectChanges, int(limit))
// 	delObjNum := uint(0)
// 	for _, change := range simulate.ObjectChanges {
// 		if change.Data.Mutated != nil {
// 			require.Equal(t, *signer.Address(), change.Data.Mutated.Sender)
// 			require.Contains(t, coins.ObjectIDVals(), change.Data.Mutated.ObjectID)
// 		} else if change.Data.Deleted != nil {
// 			delObjNum += 1
// 		}
// 	}
// 	// all the input objects are merged into the first input object
// 	// except the first input object, all the other input objects are deleted
// 	require.Equal(t, limit-1, delObjNum)

// 	// one output balance and one input balance
// 	require.Len(t, simulate.BalanceChanges, 2)
// 	for _, balChange := range simulate.BalanceChanges {
// 		if balChange.Owner.AddressOwner == signer.Address() {
// 			require.Equal(t, totalBal.Neg(totalBal), balChange.Amount)
// 		} else if balChange.Owner.AddressOwner == recipient.Address() {
// 			require.Equal(t, totalBal, balChange.Amount)
// 		}
// 	}
// }

// verified result at https://explorer.iota.org/txblock/FJARZfgxJqQL427a4dmHT15MfApGakZHsFUCRT2z2GAS?network=devnet
func TestPayIota(t *testing.T) {
	client := bindings.NewBindingClient(iotaconn.DevnetEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.DevnetFaucetURL)
	recipient1 := iotatest.MakeSignerWithFunds(1, iotaconn.DevnetFaucetURL)
	recipient2 := iotatest.MakeSignerWithFunds(2, iotaconn.DevnetFaucetURL)
	fmt.Println("signer: ", signer.Address().ShortString())
	limit := uint(1)
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

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txn.TxBytes,
			Signer:      signer,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:        true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
			},
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess(), "transaction should succeed")

	// 3 stands for the three amounts (3 created IOTA objects) in payIota API
	amountNum := uint(3)
	delObjNum := uint(0)
	createdObjNum := uint(0)
	for _, change := range txnResponse.ObjectChanges {
		if change.Data.Mutated != nil {
			require.Equal(t, *signer.Address(), change.Data.Mutated.Sender)
			require.Contains(t, coins.ObjectIDVals(), change.Data.Mutated.ObjectID)
		} else if change.Data.Created != nil {
			createdObjNum++
			require.Equal(t, *signer.Address(), change.Data.Created.Sender)
		} else if change.Data.Deleted != nil {
			delObjNum++
		}
	}

	// Input coin is deleted after being merged and split
	require.Equal(t, limit, delObjNum)
	// 1 for recipient1, and 2 for recipient2
	require.Equal(t, amountNum, createdObjNum)

	// Check balance changes
	require.NotEmpty(t, txnResponse.BalanceChanges, "should have balance changes")

	// Track balance changes per address
	balanceChanges := make(map[string]int64)
	for _, balChange := range txnResponse.BalanceChanges {
		if balChange.Owner.AddressOwner != nil {
			addr := balChange.Owner.AddressOwner.String()
			amount, _ := new(big.Int).SetString(balChange.Amount, 10)
			balanceChanges[addr] += amount.Int64()
		}
	}

	// Verify sender's balance decreased (paid amounts + gas)
	senderKey := signer.Address().String()
	require.Contains(t, balanceChanges, senderKey, "sender should have balance changes")
	totalSent := int64(sentAmounts[0] + sentAmounts[1] + sentAmounts[2])
	require.Less(t, balanceChanges[senderKey], -totalSent, "sender should pay at least the sent amounts plus gas")

	// Verify recipient1 received correct amount
	recipient1Key := recipient1.Address().String()
	require.Contains(t, balanceChanges, recipient1Key, "recipient1 should have balance changes")
	require.Equal(t, int64(sentAmounts[0]), balanceChanges[recipient1Key], "recipient1 should receive correct amount")

	// Verify recipient2 received correct total amount
	recipient2Key := recipient2.Address().String()
	require.Contains(t, balanceChanges, recipient2Key, "recipient2 should have balance changes")
	require.Equal(t, int64(sentAmounts[1]+sentAmounts[2]), balanceChanges[recipient2Key], "recipient2 should receive correct total amount")
}

func TestPublish(t *testing.T) {
	client := bindings.NewBindingClient(iotaconn.DevnetEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.DevnetFaucetURL)

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

// func TestSplitCoin(t *testing.T) {
// 	client := bindings.NewBindingClient(iotaconn.DevnetEndpointURL)
// 	signer := iotatest.MakeSignerWithFunds(0, iotaconn.DevnetFaucetURL)

// 	limit := uint(4)
// 	coinPages, err := client.GetCoins(
// 		context.Background(), iotaclient.GetCoinsRequest{
// 			Owner: signer.Address(),
// 			Limit: limit,
// 		},
// 	)
// 	require.NoError(t, err)
// 	coins := iotajsonrpc.Coins(coinPages.Data)

// 	txn, err := client.SplitCoin(
// 		context.Background(),
// 		iotaclient.SplitCoinRequest{
// 			Signer: signer.Address(),
// 			Coin:   coins[1].CoinObjectID,
// 			SplitAmounts: []*iotajsonrpc.BigInt{
// 				// assume coins[0] has more than the sum of the following splitAmounts
// 				iotajsonrpc.NewBigInt(2222),
// 				iotajsonrpc.NewBigInt(1111),
// 			},
// 			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
// 		},
// 	)
// 	require.NoError(t, err)

// 	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
// 	require.NoError(t, err)
// 	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
// 	require.True(t, simulate.Effects.Data.IsSuccess())

// 	// 2 mutated and 2 created (split coins)
// 	require.Len(t, simulate.ObjectChanges, 4)
// 	require.Len(t, simulate.BalanceChanges, 1)
// 	amt, _ := strconv.ParseInt(simulate.BalanceChanges[0].Amount, 10, 64)
// 	require.Equal(t, amt, -simulate.Effects.Data.GasFee())
// }

// func TestSplitCoinEqual(t *testing.T) {
// 	client := bindings.NewBindingClient(iotaconn.DevnetEndpointURL)
// 	signer := iotatest.MakeSignerWithFunds(0, iotaconn.DevnetFaucetURL)

// 	limit := uint(4)
// 	coinPages, err := client.GetCoins(
// 		context.Background(), iotaclient.GetCoinsRequest{
// 			Owner: signer.Address(),
// 			Limit: limit,
// 		},
// 	)
// 	require.NoError(t, err)
// 	coins := iotajsonrpc.Coins(coinPages.Data)

// 	splitShares := uint64(3)
// 	txn, err := client.SplitCoinEqual(
// 		context.Background(),
// 		iotaclient.SplitCoinEqualRequest{
// 			Signer:     signer.Address(),
// 			Coin:       coins[0].CoinObjectID,
// 			SplitCount: iotajsonrpc.NewBigInt(splitShares),
// 			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
// 		},
// 	)
// 	require.NoError(t, err)

// 	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
// 	require.NoError(t, err)
// 	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
// 	require.True(t, simulate.Effects.Data.IsSuccess())

// 	// 1 mutated and 3 created (split coins)
// 	require.Len(t, simulate.ObjectChanges, 1+int(splitShares))
// 	require.Len(t, simulate.BalanceChanges, 1)
// 	amt, _ := strconv.ParseInt(simulate.BalanceChanges[0].Amount, 10, 64)
// 	require.Equal(t, amt, -simulate.Effects.Data.GasFee())
// }

// func TestTransferObject(t *testing.T) {
// 	client := bindings.NewBindingClient(iotaconn.DevnetEndpointURL)
// 	signer := iotatest.MakeSignerWithFunds(0, iotaconn.DevnetFaucetURL)
// 	recipient := iotatest.MakeSignerWithFunds(1, iotaconn.DevnetFaucetURL)

// 	limit := uint(3)
// 	coinPages, err := client.GetCoins(
// 		context.Background(), iotaclient.GetCoinsRequest{
// 			Owner: signer.Address(),
// 			Limit: limit,
// 		},
// 	)
// 	require.NoError(t, err)
// 	transferCoin := coinPages.Data[0]

// 	txn, err := client.TransferObject(
// 		context.Background(),
// 		iotaclient.TransferObjectRequest{
// 			Signer:    signer.Address(),
// 			Recipient: recipient.Address(),
// 			ObjectID:  transferCoin.CoinObjectID,
// 			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
// 		},
// 	)
// 	require.NoError(t, err)

// 	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
// 	require.NoError(t, err)
// 	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
// 	require.True(t, simulate.Effects.Data.IsSuccess())

// 	// one is transferred object, one is the gas object
// 	require.Len(t, simulate.ObjectChanges, 2)

// 	require.Len(t, simulate.BalanceChanges, 2)
// }

// func TestTransferIota(t *testing.T) {
// 	client := bindings.NewBindingClient(iotaconn.DevnetEndpointURL)
// 	signer := iotatest.MakeSignerWithFunds(0, iotaconn.DevnetFaucetURL)
// 	recipient := iotatest.MakeSignerWithFunds(1, iotaconn.DevnetFaucetURL)

// 	limit := uint(3)
// 	coinPages, err := client.GetCoins(
// 		context.Background(), iotaclient.GetCoinsRequest{
// 			Owner: signer.Address(),
// 			Limit: limit,
// 		},
// 	)
// 	require.NoError(t, err)
// 	transferCoin := coinPages.Data[0]

// 	txn, err := client.TransferIota(
// 		context.Background(),
// 		iotaclient.TransferIotaRequest{
// 			Signer:    signer.Address(),
// 			Recipient: recipient.Address(),
// 			ObjectID:  transferCoin.CoinObjectID,
// 			Amount:    iotajsonrpc.NewBigInt(3),
// 			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
// 		},
// 	)
// 	require.NoError(t, err)

// 	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
// 	require.NoError(t, err)
// 	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
// 	require.True(t, simulate.Effects.Data.IsSuccess())

// 	// one is transferred object, one is the gas object
// 	require.Len(t, simulate.ObjectChanges, 2)
// 	for _, change := range simulate.ObjectChanges {
// 		if change.Data.Mutated != nil {
// 			require.Equal(t, *transferCoin.CoinObjectID, change.Data.Mutated.ObjectID)
// 			require.Equal(t, signer.Address(), change.Data.Mutated.Owner.AddressOwner)
// 		} else if change.Data.Created != nil {
// 			require.Equal(t, recipient.Address(), change.Data.Created.Owner.AddressOwner)
// 		}
// 	}

// 	require.Len(t, simulate.BalanceChanges, 2)
// }
