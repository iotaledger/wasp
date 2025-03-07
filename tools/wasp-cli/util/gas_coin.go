package util

import (
	"context"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
)

func CreateAndSendGasCoin(ctx context.Context, client clients.L1Client, wallet wallets.Wallet, committeeAddress *iotago.Address) (iotago.ObjectID, error) {
	coins, err := client.GetCoinObjsForTargetAmount(ctx, wallet.Address().AsIotaAddress(), isc.GasCoinTargetValue, isc.GasCoinTargetValue)
	if err != nil {
		return iotago.ObjectID{}, err
	}

	txb := iotago.NewProgrammableTransactionBuilder()
	splitCoinCmd := txb.Command(
		iotago.Command{
			SplitCoins: &iotago.ProgrammableSplitCoins{
				Coin:    iotago.GetArgumentGasCoin(),
				Amounts: []iotago.Argument{txb.MustPure(isc.GasCoinTargetValue)},
			},
		},
	)

	txb.TransferArg(committeeAddress, splitCoinCmd)

	txData := iotago.NewProgrammable(
		wallet.Address().AsIotaAddress(),
		txb.Finish(),
		[]*iotago.ObjectRef{coins[0].Ref()},
		uint64(isc.GasCoinTargetValue),
		parameters.L1().Protocol.ReferenceGasPrice.Uint64(),
	)

	txnBytes, err := bcs.Marshal(&txData)
	if err != nil {
		return iotago.ObjectID{}, err
	}

	result, err := client.SignAndExecuteTransaction(
		ctx,
		&iotaclient.SignAndExecuteTransactionRequest{
			Signer:      cryptolib.SignerToIotaSigner(wallet),
			TxDataBytes: txnBytes,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	if err != nil {
		return iotago.ObjectID{}, err
	}

	gasCoin, err := result.GetCreatedCoin("iota", "IOTA")
	if err != nil {
		return iotago.ObjectID{}, err
	}

	return *gasCoin.ObjectID, nil
}
