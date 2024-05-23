package sui_types_test

import (
	"context"
	"testing"

	"github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
	"github.com/stretchr/testify/require"
)

func TestPTB_Pay(t *testing.T) {
	client, sender := sui.NewTestSuiClientWithSignerAndFund(conn.DevnetEndpointUrl, sui_signer.TEST_MNEMONIC)
	recipients := sui_signer.NewRandomSigners(sui_signer.KeySchemeFlagDefault, 2)
	coinType := models.SuiCoinType
	limit := uint(3)
	coinPages, err := client.GetCoins(context.Background(), sender.Address, &coinType, nil, limit)
	require.NoError(t, err)
	coins := models.Coins(coinPages.Data)
	gasCoin := coins[0] // save the 1st element for gas fee
	transferCoins := coins[1:]
	amounts := []uint64{123, 567}
	totalBal := coins.TotalBalance().Uint64()

	ptb := sui_types.NewProgrammableTransactionBuilder()
	err = ptb.Pay(
		transferCoins.CoinRefs(),
		[]*sui_types.SuiAddress{recipients[0].Address, recipients[1].Address},
		[]uint64{amounts[0], amounts[1]},
	)
	require.NoError(t, err)
	pt := ptb.Finish()
	tx := sui_types.NewProgrammable(
		sender.Address,
		pt,
		[]*sui_types.ObjectRef{
			gasCoin.Ref(),
		},
		sui.DefaultGasBudget,
		sui.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(tx)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())
	require.Equal(t, gasCoin.CoinObjectID.String(), simulate.Effects.Data.V1.GasObject.Reference.ObjectID)

	// 2 for Mutated (1 gas coin and 1 merged coin in pay pt), 2 created (the 2 transfer in pay pt),
	require.Len(t, simulate.ObjectChanges, 5)
	for _, change := range simulate.ObjectChanges {
		if change.Data.Mutated != nil {
			require.Contains(
				t,
				[]*sui_types.ObjectID{gasCoin.CoinObjectID, transferCoins[0].CoinObjectID},
				&change.Data.Mutated.ObjectID,
			)
		} else if change.Data.Deleted != nil {
			require.Equal(t, transferCoins[1].CoinObjectID, &change.Data.Deleted.ObjectID)
		}
	}
	require.Len(t, simulate.BalanceChanges, 3)
	for _, balChange := range simulate.BalanceChanges {
		if balChange.Owner.AddressOwner == sender.Address {
			require.Equal(t, totalBal-(amounts[0]+amounts[1]), balChange.Amount)
		} else if balChange.Owner.AddressOwner == recipients[0].Address {
			require.Equal(t, amounts[0], balChange.Amount)
		} else if balChange.Owner.AddressOwner == recipients[1].Address {
			require.Equal(t, amounts[1], balChange.Amount)
		}
	}
}

func TestPTB_PaySui(t *testing.T) {
	var err error
	client, sender := sui.NewTestSuiClientWithSignerAndFund(conn.DevnetEndpointUrl, sui_signer.TEST_MNEMONIC)
	recipients := sui_signer.NewRandomSigners(sui_signer.KeySchemeFlagDefault, 2)
	coinType := models.SuiCoinType
	limit := uint(3)
	coinPages, err := client.GetCoins(context.Background(), sender.Address, &coinType, nil, limit)
	require.NoError(t, err)
	coin := coinPages.Data[0]

	ptb := sui_types.NewProgrammableTransactionBuilder()
	err = ptb.PaySui(
		[]*sui_types.SuiAddress{recipients[0].Address, recipients[1].Address},
		[]uint64{123, 456},
	)
	require.NoError(t, err)
	pt := ptb.Finish()

	tx := sui_types.NewProgrammable(
		sender.Address,
		pt,
		[]*sui_types.ObjectRef{
			coin.Ref(),
		},
		sui.DefaultGasBudget,
		sui.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(tx)
	require.NoError(t, err)

	// txnResponse, err := client.SignAndExecuteTransaction(context.TODO(), sender, txBytes, &models.SuiTransactionBlockResponseOptions{
	// 	ShowInput:          true,
	// 	ShowEffects:        true,
	// 	ShowObjectChanges:  true,
	// 	ShowBalanceChanges: true,
	// 	ShowEvents:         true,
	// })
	// require.NoError(t, err)
	// require.True(t, txnResponse.Effects.Data.IsSuccess())

	simulate, err := client.DryRunTransaction(context.Background(), txBytes)
	require.NoError(t, err)
	require.Empty(t, simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())
	require.Equal(t, coin.CoinObjectID.String(), simulate.Effects.Data.V1.GasObject.Reference.ObjectID)

	// 2 for Mutated (1 gas coin and 1 merged coin in pay pt), 2 created (the 2 transfer in pay pt),
	// 	require.Len(t, simulate.ObjectChanges, 5)
	// for _, change := range simulate.ObjectChanges {
	// 	if change.Data.Mutated != nil {
	// 		fmt.Println("change.Data.Mutated.ObjectID: ", change.Data.Mutated.ObjectID)
	// 	} else if change.Data.Created != nil {
	// 		fmt.Println("change.Data.Created.ObjectID: ", change.Data.Created.ObjectID)
	// 	} else if change.Data.Deleted != nil {
	// 		fmt.Println("change.Data.Deleted.ObjectID: ", change.Data.Deleted.ObjectID)
	// 	}
	// if change.Data.Mutated != nil {
	// 	require.Contains(t, []*sui_types.ObjectID{gasCoin.CoinObjectID, transferCoins[0].CoinObjectID}, &change.Data.Mutated.ObjectID)
	// } else if change.Data.Deleted != nil {
	// 	require.Equal(t, transferCoins[1].CoinObjectID, &change.Data.Deleted.ObjectID)
	// }
	// }
	// 	require.Len(t, simulate.BalanceChanges, 3)
	// 	for _, balChange := range simulate.BalanceChanges {
	// 		if balChange.Owner.AddressOwner == sender.Address {
	// 			require.Equal(t, totalBal-(amounts[0]+amounts[1]), balChange.Amount)
	// 		} else if balChange.Owner.AddressOwner == recipients[0].Address {
	// 			require.Equal(t, amounts[0], balChange.Amount)
	// 		} else if balChange.Owner.AddressOwner == recipients[1].Address {
	// 			require.Equal(t, amounts[1], balChange.Amount)
	// 		}
	// 	}
}

// func TestPTB_TransferObject(t *testing.T) {
// 	sender := sui_signer.TEST_ADDRESS
// 	recipient := sui_signer.TEST_ADDRESS
// 	gasBudget := sui_types.SUI(0.1).Uint64()

// 	api := sui.NewSuiClient(conn.TestnetEndpointUrl)
// 	err := sui.RequestFundFromFaucet(sender, conn.TestnetFaucetUrl)
// 	require.NoError(t, err)
// 	coins := getCoins(t, api, sender, 2)
// 	coin, gas := coins[0], coins[1]

// 	gasPrice := uint64(1000)
// 	// gasPrice, err := api.GetReferenceGasPrice(context.Background())

// 	// build with BCS
// 	ptb := sui_types.NewProgrammableTransactionBuilder()
// 	err = ptb.TransferObject(recipient, []*sui_types.ObjectRef{coin.Ref()})
// 	require.NoError(t, err)
// 	pt := ptb.Finish()
// 	tx := sui_types.NewProgrammable(
// 		sender,
// 		pt,
// 		[]*sui_types.ObjectRef{
// 			gas.Ref(),
// 		},
// 		gasBudget,
// 		gasPrice,
// 	)
// 	txBytesBCS, err := bcs.Marshal(tx)
// 	require.NoError(t, err)

// 	// build with remote rpc
// 	txn, err := api.TransferObject(
// 		context.Background(), sender, recipient,
// 		coin.CoinObjectID,
// 		gas.CoinObjectID,
// 		models.NewSafeSuiBigInt(gasBudget),
// 	)
// 	require.NoError(t, err)
// 	txBytesRemote := txn.TxBytes.Data()

// 	require.Equal(t, txBytesBCS, txBytesRemote)
// }

// func TestPTB_TransferSui(t *testing.T) {
// 	sender := sui_signer.TEST_ADDRESS
// 	recipient := sender
// 	amount := sui_types.SUI(0.001).Uint64()
// 	gasBudget := sui_types.SUI(0.01).Uint64()

// 	api := sui.NewSuiClient(conn.TestnetEndpointUrl)
// 	err := sui.RequestFundFromFaucet(sender, conn.TestnetFaucetUrl)
// 	require.NoError(t, err)
// 	coin := getCoins(t, api, sender, 1)[0]

// 	gasPrice := uint64(1000)
// 	// gasPrice, err := api.GetReferenceGasPrice(context.Background())

// 	// build with BCS
// 	ptb := sui_types.NewProgrammableTransactionBuilder()
// 	err = ptb.TransferSui(recipient, &amount)
// 	require.NoError(t, err)
// 	pt := ptb.Finish()
// 	tx := sui_types.NewProgrammable(
// 		sender,
// 		pt,
// 		[]*sui_types.ObjectRef{
// 			coin.Ref(),
// 		},
// 		gasBudget,
// 		gasPrice,
// 	)
// 	txBytesBCS, err := bcs.Marshal(tx)
// 	require.NoError(t, err)

// 	// build with remote rpc
// 	txn, err := api.TransferSui(
// 		context.Background(), sender, recipient, coin.CoinObjectID,
// 		models.NewSafeSuiBigInt(amount),
// 		models.NewSafeSuiBigInt(gasBudget),
// 	)
// 	require.NoError(t, err)
// 	txBytesRemote := txn.TxBytes.Data()

// 	require.Equal(t, txBytesBCS, txBytesRemote)
// }

// func TestPTB_PayAllSui(t *testing.T) {
// 	sender := sui_signer.TEST_ADDRESS
// 	recipient := sender
// 	// gasBudget := sui_types.SUI(0.01).Uint64()

// 	api := sui.NewSuiClient(conn.TestnetEndpointUrl)
// 	err := sui.RequestFundFromFaucet(sender, conn.TestnetFaucetUrl)
// 	require.NoError(t, err)
// 	coins := getCoins(t, api, sender, 2)
// 	coin, coin2 := coins[0], coins[1]

// 	// build with BCS
// 	ptb := sui_types.NewProgrammableTransactionBuilder()
// 	err = ptb.PayAllSui(recipient)
// 	require.NoError(t, err)
// 	pt := ptb.Finish()
// 	tx := sui_types.NewProgrammable(
// 		sender,
// 		pt,
// 		[]*sui_types.ObjectRef{
// 			coin.Ref(),
// 			coin2.Ref(),
// 		},
// 		sui.DefaultGasBudget,
// 		sui.DefaultGasPrice,
// 	)
// 	txBytesBCS, err := bcs.Marshal(tx)
// 	require.NoError(t, err)

// 	// build with remote rpc
// 	txn, err := api.PayAllSui(
// 		context.Background(), sender, recipient,
// 		[]*sui_types.ObjectID{
// 			coin.CoinObjectID, coin2.CoinObjectID,
// 		},
// 		models.NewSafeSuiBigInt(sui.DefaultGasBudget),
// 	)
// 	require.NoError(t, err)
// 	txBytesRemote := txn.TxBytes.Data()

// 	require.Equal(t, txBytesBCS, txBytesRemote)
// }

// func TestPTB_MoveCall(t *testing.T) {
// 	sender := sui_signer.TEST_ADDRESS
// 	gasBudget := sui_types.SUI(0.1).Uint64()
// 	gasPrice := uint64(1000)

// 	api := sui.NewSuiClient(conn.TestnetEndpointUrl)
// 	err := sui.RequestFundFromFaucet(sender, conn.TestnetFaucetUrl)
// 	require.NoError(t, err)
// 	coins := getCoins(t, api, sender, 2)
// 	coin, coin2 := coins[0], coins[1]

// 	validatorAddress, err := sui_types.SuiAddressFromHex(ComingChatValidatorAddress)
// 	require.NoError(t, err)

// 	// build with BCS
// 	ptb := sui_types.NewProgrammableTransactionBuilder()

// 	// case 1: split target amount
// 	amtArg, err := ptb.Pure(sui_types.SUI(1).Uint64())
// 	require.NoError(t, err)
// 	arg1 := ptb.Command(
// 		sui_types.Command{
// 			SplitCoins: &sui_types.ProgrammableSplitCoins{
// 				Coin:    sui_types.Argument{GasCoin: &serialization.EmptyEnum{}},
// 				Amounts: []sui_types.Argument{amtArg},
// 			},
// 		},
// 	) // the coin is split result argument
// 	arg2, err := ptb.Pure(validatorAddress)
// 	require.NoError(t, err)
// 	arg0, err := ptb.Obj(sui_types.SuiSystemMutObj)
// 	require.NoError(t, err)
// 	ptb.Command(
// 		sui_types.Command{
// 			MoveCall: &sui_types.ProgrammableMoveCall{
// 				Package:  sui_types.SuiSystemAddress,
// 				Module:   sui_types.SuiSystemModuleName,
// 				Function: sui_types.AddStakeFunName,
// 				Arguments: []sui_types.Argument{
// 					arg0, arg1, arg2,
// 				},
// 			},
// 		},
// 	)
// 	pt := ptb.Finish()
// 	tx := sui_types.NewProgrammable(
// 		sender,
// 		pt,
// 		[]*sui_types.ObjectRef{
// 			coin.Ref(),
// 			coin2.Ref(),
// 		},
// 		gasBudget,
// 		gasPrice,
// 	)

// 	// case 2: direct stake the specified coin
// 	// coinArg := sui_types.CallArg{
// 	// 	Object: &sui_types.ObjectArg{
// 	// 		ImmOrOwnedObject: coin.Ref(),
// 	// 	},
// 	// }
// 	// addrBytes := validatorAddress.Data()
// 	// addrArg := sui_types.CallArg{
// 	// 	Pure: &addrBytes,
// 	// }
// 	// err = ptb.MoveCall(
// 	// 	*sui_types.SuiSystemAddress,
// 	// 	sui_system_state.SuiSystemModuleName,
// 	// 	sui_types.AddStakeFunName,
// 	// 	[]sui_types.TypeTag{},
// 	// 	[]sui_types.CallArg{
// 	// 		sui_types.SuiSystemMut,
// 	// 		coinArg,
// 	// 		addrArg,
// 	// 	},
// 	// )
// 	// require.NoError(t, err)
// 	// pt := ptb.Finish()
// 	// tx := sui_types.NewProgrammable(
// 	// 	sender, []*sui_types.ObjectRef{
// 	// 		coin2.Ref(),
// 	// 	},
// 	// 	pt, gasBudget, gasPrice,
// 	// )

// 	// build & simulate
// 	txBytesBCS, err := bcs.Marshal(tx)
// 	require.NoError(t, err)
// 	resp := dryRunTxn(t, api, txBytesBCS, true)
// 	t.Log(resp.Effects.Data.GasFee())
// }

func TestTransferSui(t *testing.T) {
	recipient, err := sui_types.SuiAddressFromHex("0x7e875ea78ee09f08d72e2676cf84e0f1c8ac61d94fa339cc8e37cace85bebc6e")
	require.NoError(t, err)

	ptb := sui_types.NewProgrammableTransactionBuilder()
	amount := uint64(100000)
	err = ptb.TransferSui(recipient, &amount)
	require.NoError(t, err)
	pt := ptb.Finish()
	digest := sui_types.MustNewDigest("HvbE2UZny6cP4KukaXetmj4jjpKTDTjVo23XEcu7VgSn")
	objectId, err := sui_types.SuiAddressFromHex("0x13c1c3d0e15b4039cec4291c75b77c972c10c8e8e70ab4ca174cf336917cb4db")
	require.NoError(t, err)
	tx := sui_types.NewProgrammable(
		recipient,
		pt,
		[]*sui_types.ObjectRef{
			{
				ObjectID: objectId,
				Version:  14924029,
				Digest:   digest,
			},
		},
		10000000,
		1000,
	)
	txByte, err := bcs.Marshal(tx)
	require.NoError(t, err)
	t.Logf("%x", txByte)
}
