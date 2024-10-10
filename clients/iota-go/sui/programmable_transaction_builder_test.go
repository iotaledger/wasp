package sui_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/suiclient"
	"github.com/iotaledger/wasp/clients/iota-go/suiconn"
	"github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/suitest"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestPTBMoveCall(t *testing.T) {
	t.Run(
		"access_multiple_return_values_from_move_func", func(t *testing.T) {
			client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
			sender := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)

			_, packageID, err := client.PublishContract(
				context.Background(),
				sender,
				contracts.SDKVerify().Modules,
				contracts.SDKVerify().Dependencies,
				suiclient.DefaultGasBudget,
				&suijsonrpc.SuiTransactionBlockResponseOptions{ShowObjectChanges: true, ShowEffects: true},
			)
			require.NoError(t, err)

			coinPages, err := client.GetCoins(
				context.Background(), suiclient.GetCoinsRequest{
					Owner: sender.Address(),
					Limit: 3,
				},
			)
			require.NoError(t, err)
			coins := suijsonrpc.Coins(coinPages.Data)

			ptb := sui.NewProgrammableTransactionBuilder()
			require.NoError(t, err)

			ptb.Command(
				sui.Command{
					MoveCall: &sui.ProgrammableMoveCall{
						Package:       packageID,
						Module:        "sdk_verify",
						Function:      "ret_two_1",
						TypeArguments: []sui.TypeTag{},
						Arguments:     []sui.Argument{},
					},
				},
			)
			ptb.Command(
				sui.Command{
					MoveCall: &sui.ProgrammableMoveCall{
						Package:       packageID,
						Module:        "sdk_verify",
						Function:      "ret_two_2",
						TypeArguments: []sui.TypeTag{},
						Arguments: []sui.Argument{
							{NestedResult: &sui.NestedResult{Cmd: 0, Result: 1}},
							{NestedResult: &sui.NestedResult{Cmd: 0, Result: 0}},
						},
					},
				},
			)
			pt := ptb.Finish()
			txData := sui.NewProgrammable(
				sender.Address(),
				pt,
				[]*sui.ObjectRef{coins[0].Ref()},
				suiclient.DefaultGasBudget,
				suiclient.DefaultGasPrice,
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
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	sender := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)

	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 2,
		},
	)
	require.NoError(t, err)
	coins := suijsonrpc.Coins(coinPages.Data)
	gasCoin := coins[0]
	transferCoin := coins[1]

	ptb := sui.NewProgrammableTransactionBuilder()
	err = ptb.TransferObject(recipient.Address(), transferCoin.Ref())
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := sui.NewProgrammable(
		sender.Address(),
		pt,
		[]*sui.ObjectRef{gasCoin.Ref()},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	txn, err := client.TransferObject(
		context.Background(),
		suiclient.TransferObjectRequest{
			Signer:    sender.Address(),
			Recipient: recipient.Address(),
			ObjectID:  transferCoin.CoinObjectID,
			Gas:       gasCoin.CoinObjectID,
			GasBudget: suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}

func TestPTBTransferSui(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	sender := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)

	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 1,
		},
	)
	require.NoError(t, err)
	coin := suijsonrpc.Coins(coinPages.Data)[0]
	amount := uint64(123)

	// build with BCS
	ptb := sui.NewProgrammableTransactionBuilder()
	err = ptb.TransferSui(recipient.Address(), &amount)
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := sui.NewProgrammable(
		sender.Address(),
		pt,
		[]*sui.ObjectRef{coin.Ref()},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txBytesBCS, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	txn, err := client.TransferSui(
		context.Background(),
		suiclient.TransferSuiRequest{
			Signer:    sender.Address(),
			Recipient: recipient.Address(),
			ObjectID:  coin.CoinObjectID,
			Amount:    suijsonrpc.NewBigInt(amount),
			GasBudget: suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytesBCS, txBytesRemote)
}

func TestPTBPayAllSui(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	sender := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)

	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 3,
		},
	)
	require.NoError(t, err)
	coins := suijsonrpc.Coins(coinPages.Data)

	// build with BCS
	ptb := sui.NewProgrammableTransactionBuilder()
	err = ptb.PayAllSui(recipient.Address())
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := sui.NewProgrammable(
		sender.Address(),
		pt,
		coins.CoinRefs(),
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// build with remote rpc
	txn, err := client.PayAllSui(
		context.Background(),
		suiclient.PayAllSuiRequest{
			Signer:     sender.Address(),
			Recipient:  recipient.Address(),
			InputCoins: coins.ObjectIDs(),
			GasBudget:  suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}

func TestPTBPaySui(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	sender := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient1 := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)
	recipient2 := suitest.MakeSignerWithFunds(2, suiconn.AlphanetFaucetURL)

	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 1,
		},
	)
	require.NoError(t, err)
	coin := coinPages.Data[0]

	ptb := sui.NewProgrammableTransactionBuilder()
	err = ptb.PaySui(
		[]*sui.Address{recipient1.Address(), recipient2.Address()},
		[]uint64{123, 456},
	)
	require.NoError(t, err)
	pt := ptb.Finish()

	tx := sui.NewProgrammable(
		sender.Address(),
		pt,
		[]*sui.ObjectRef{
			coin.Ref(),
		},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())
	require.Equal(t, coin.CoinObjectID.String(), simulate.Effects.Data.V1.GasObject.Reference.ObjectID.String())

	// 1 for Mutated, 2 created (the 2 transfer in pay_sui pt),
	require.Len(t, simulate.ObjectChanges, 3)
	for _, change := range simulate.ObjectChanges {
		if change.Data.Mutated != nil {
			require.Equal(t, coin.CoinObjectID, &change.Data.Mutated.ObjectID)
		} else if change.Data.Created != nil {
			require.Contains(
				t,
				[]*sui.Address{recipient1.Address(), recipient2.Address()},
				change.Data.Created.Owner.AddressOwner,
			)
		}
	}

	// build with remote rpc
	txn, err := client.PaySui(
		context.Background(),
		suiclient.PaySuiRequest{
			Signer:     sender.Address(),
			InputCoins: []*sui.ObjectID{coin.CoinObjectID},
			Recipients: []*sui.Address{recipient1.Address(), recipient2.Address()},
			Amount:     []*suijsonrpc.BigInt{suijsonrpc.NewBigInt(123), suijsonrpc.NewBigInt(456)},
			GasBudget:  suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}

func TestPTBPay(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	sender := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)
	recipient1 := suitest.MakeSignerWithFunds(1, suiconn.AlphanetFaucetURL)
	recipient2 := suitest.MakeSignerWithFunds(2, suiconn.AlphanetFaucetURL)

	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: 3,
		},
	)
	require.NoError(t, err)
	coins := suijsonrpc.Coins(coinPages.Data)
	gasCoin := coins[0] // save the 1st element for gas fee
	transferCoins := coins[1:]
	amounts := []uint64{123, 567}
	totalBal := coins.TotalBalance().Uint64()

	ptb := sui.NewProgrammableTransactionBuilder()
	err = ptb.Pay(
		transferCoins.CoinRefs(),
		[]*sui.Address{recipient1.Address(), recipient2.Address()},
		[]uint64{amounts[0], amounts[1]},
	)
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := sui.NewProgrammable(
		sender.Address(),
		pt,
		[]*sui.ObjectRef{
			gasCoin.Ref(),
		},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
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
				[]*sui.ObjectID{gasCoin.CoinObjectID, transferCoins[0].CoinObjectID},
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
		suiclient.PayRequest{
			Signer:     sender.Address(),
			InputCoins: transferCoins.ObjectIDs(),
			Recipients: []*sui.Address{recipient1.Address(), recipient2.Address()},
			Amount:     []*suijsonrpc.BigInt{suijsonrpc.NewBigInt(amounts[0]), suijsonrpc.NewBigInt(amounts[1])},
			Gas:        gasCoin.CoinObjectID,
			GasBudget:  suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)
	txBytesRemote := txn.TxBytes.Data()
	require.Equal(t, txBytes, txBytesRemote)
}
