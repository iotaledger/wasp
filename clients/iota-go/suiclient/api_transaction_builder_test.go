package suiclient_test

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/suiclient"
	"github.com/iotaledger/wasp/clients/iota-go/suiconn"
	"github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/suitest"
)

func TestBatchTransaction(t *testing.T) {
	t.Log("TestBatchTransaction TODO")
	// api := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)

	// txnBytes, err := api.BatchTransaction(context.Background(), signer, *coin1, *coin2, nil, 10000)
	// require.NoError(t, err)
	// dryRunTxn(t, api, txnBytes, M1Account(t))
}

func TestMergeCoins(t *testing.T) {
	t.Skip("FIXME create an account has at least two coin objects on chain")
	// api := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	// signer := testAddress
	// coins, err := api.GetCoins(context.Background(), suiclient.GetCoinsRequest{
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
	// 	suiclient.MergeCoinsRequest{
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
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)

	sdkVerifyBytecode := contracts.SDKVerify()

	txnBytes, err := client.Publish(
		context.Background(),
		suiclient.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: sdkVerifyBytecode.Modules,
			Dependencies:    sdkVerifyBytecode.Dependencies,
			GasBudget:       suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		signer,
		txnBytes.TxBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	// test MoveCall with byte array input
	input := []string{"haha", "gogo"}
	txnBytes, err = client.MoveCall(
		context.Background(),
		suiclient.MoveCallRequest{
			Signer:    signer.Address(),
			PackageID: packageID,
			Module:    "sdk_verify",
			Function:  "read_input_bytes_array",
			TypeArgs:  []string{},
			Arguments: []any{input},
			GasBudget: suijsonrpc.NewBigInt((suiclient.DefaultGasBudget)),
		},
	)
	require.NoError(t, err)
	txnResponse, err = client.SignAndExecuteTransaction(
		context.Background(),
		signer,
		txnBytes.TxBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	queryEventsRes, err := client.QueryEvents(
		context.Background(),
		suiclient.QueryEventsRequest{
			Query: &suijsonrpc.EventFilter{Transaction: &txnResponse.Digest},
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
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)

	coins, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)
	limit := len(coins.Data) - 1 // need reserve a coin for gas
	totalBal := suijsonrpc.Coins(coins.Data).TotalBalance().Uint64()

	amount := uint64(123)
	pickedCoins, err := suijsonrpc.PickupCoins(
		coins,
		new(big.Int).SetUint64(amount),
		suiclient.DefaultGasBudget,
		limit,
		0,
	)
	require.NoError(t, err)

	txn, err := client.Pay(
		context.Background(),
		suiclient.PayRequest{
			Signer:     signer.Address(),
			InputCoins: pickedCoins.CoinIds(),
			Recipients: []*sui.Address{recipient.Address()},
			Amount:     []*suijsonrpc.BigInt{suijsonrpc.NewBigInt(amount)},
			GasBudget:  suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
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

func TestPayAllSui(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)

	limit := uint(3)
	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := suijsonrpc.Coins(coinPages.Data)
	// assume the account holds more than 'limit' amount Sui token objects
	require.Len(t, coinPages.Data, 3)
	totalBal := coins.TotalBalance()

	txn, err := client.PayAllSui(
		context.Background(),
		suiclient.PayAllSuiRequest{
			Signer:     signer.Address(),
			Recipient:  recipient.Address(),
			InputCoins: coins.ObjectIDs(),
			GasBudget:  suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
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

func TestPaySui(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient1 := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)
	recipient2 := suitest.MakeSignerWithFunds(2, suiconn.AlphanetFaucetURL)

	limit := uint(4)
	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := suijsonrpc.Coins(coinPages.Data)

	sentAmounts := []uint64{123, 456, 789}
	txn, err := client.PaySui(
		context.Background(),
		suiclient.PaySuiRequest{
			Signer:     signer.Address(),
			InputCoins: coins.ObjectIDs(),
			Recipients: []*sui.Address{
				recipient1.Address(),
				recipient2.Address(),
				recipient2.Address(),
			},
			Amount: []*suijsonrpc.BigInt{
				suijsonrpc.NewBigInt(sentAmounts[0]), // to recipient1
				suijsonrpc.NewBigInt(sentAmounts[1]), // to recipient2
				suijsonrpc.NewBigInt(sentAmounts[2]), // to recipient2
			},
			GasBudget: suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	// 3 stands for the three amounts (3 crated SUI objects) in unsafe_paySui API
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
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)

	testcoinBytecode := contracts.Testcoin()

	txnBytes, err := client.Publish(
		context.Background(),
		suiclient.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: testcoinBytecode.Modules,
			Dependencies:    testcoinBytecode.Dependencies,
			GasBudget:       suijsonrpc.NewBigInt(suiclient.DefaultGasBudget * 5),
		},
	)
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		signer,
		txnBytes.TxBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())
}

func TestSplitCoin(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)

	limit := uint(4)
	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := suijsonrpc.Coins(coinPages.Data)

	txn, err := client.SplitCoin(
		context.Background(),
		suiclient.SplitCoinRequest{
			Signer: signer.Address(),
			Coin:   coins[1].CoinObjectID,
			SplitAmounts: []*suijsonrpc.BigInt{
				// assume coins[0] has more than the sum of the following splitAmounts
				suijsonrpc.NewBigInt(2222),
				suijsonrpc.NewBigInt(1111),
			},
			GasBudget: suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	// 2 mutated and 2 created (split coins)
	require.Len(t, simulate.ObjectChanges, 4)
	// TODO check each element ObjectChanges
	require.Len(t, simulate.BalanceChanges, 1)
	amt, _ := strconv.ParseInt(simulate.BalanceChanges[0].Amount, 10, 64)
	require.Equal(t, amt, -simulate.Effects.Data.GasFee())
}

func TestSplitCoinEqual(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)

	limit := uint(4)
	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := suijsonrpc.Coins(coinPages.Data)

	splitShares := uint64(3)
	txn, err := client.SplitCoinEqual(
		context.Background(),
		suiclient.SplitCoinEqualRequest{
			Signer:     signer.Address(),
			Coin:       coins[0].CoinObjectID,
			SplitCount: suijsonrpc.NewBigInt(splitShares),
			GasBudget:  suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txn.TxBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())

	// 1 mutated and 3 created (split coins)
	require.Len(t, simulate.ObjectChanges, 1+int(splitShares))
	// TODO check each element ObjectChanges
	require.Len(t, simulate.BalanceChanges, 1)
	amt, _ := strconv.ParseInt(simulate.BalanceChanges[0].Amount, 10, 64)
	require.Equal(t, amt, -simulate.Effects.Data.GasFee())
}

func TestTransferObject(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)

	limit := uint(3)
	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	transferCoin := coinPages.Data[0]

	txn, err := client.TransferObject(
		context.Background(),
		suiclient.TransferObjectRequest{
			Signer:    signer.Address(),
			Recipient: recipient.Address(),
			ObjectID:  transferCoin.CoinObjectID,
			GasBudget: suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
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

func TestTransferSui(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)

	limit := uint(3)
	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	transferCoin := coinPages.Data[0]

	txn, err := client.TransferSui(
		context.Background(),
		suiclient.TransferSuiRequest{
			Signer:    signer.Address(),
			Recipient: recipient.Address(),
			ObjectID:  transferCoin.CoinObjectID,
			Amount:    suijsonrpc.NewBigInt(3),
			GasBudget: suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
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
