package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func PTBAssetsBagNew(ptb *sui.ProgrammableTransactionBuilder, packageID sui.PackageID, owner *cryptolib.Address) *sui.ProgrammableTransactionBuilder {
	arg1 := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "new",
				TypeArguments: []sui.TypeTag{},
				Arguments:     []sui.Argument{},
			},
		},
	)
	ptb.Command(
		sui.Command{
			TransferObjects: &sui.ProgrammableTransferObjects{
				Objects: []sui.Argument{arg1},
				Address: ptb.MustPure(owner.AsSuiAddress()),
			},
		},
	)
	return ptb
}

func PTBAssetsBagPlaceCoin(ptb *sui.ProgrammableTransactionBuilder, packageID sui.PackageID, argAssetsBag sui.Argument, coin *sui.ObjectRef, coinType string) *sui.ProgrammableTransactionBuilder {
	typeTag, err := sui.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin",
				TypeArguments: []sui.TypeTag{*typeTag},
				Arguments: []sui.Argument{
					argAssetsBag,
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: coin}),
				},
			},
		},
	)
	return ptb
}

func PTBAssetsBagPlaceCoinWithAmount(ptb *sui.ProgrammableTransactionBuilder, packageID sui.PackageID, assetsBagRef *sui.ObjectRef, coin *sui.ObjectRef, amount uint64, coinType string) *sui.ProgrammableTransactionBuilder {
	typeTag, err := sui.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	splitCoinArg := ptb.Command(
		sui.Command{
			SplitCoins: &sui.ProgrammableSplitCoins{
				Coin:    ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: coin}),
				Amounts: []sui.Argument{ptb.MustForceSeparatePure(amount)},
			},
		},
	)
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin",
				TypeArguments: []sui.TypeTag{*typeTag},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
					splitCoinArg,
				},
			},
		},
	)
	return ptb
}

func PTBAssetsDestroyEmpty(ptb *sui.ProgrammableTransactionBuilder, packageID sui.PackageID, argAssetsBag sui.Argument) *sui.ProgrammableTransactionBuilder {
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "destroy_empty",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					argAssetsBag,
				},
			},
		},
	)
	return ptb
}
