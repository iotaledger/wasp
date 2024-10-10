package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove"
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func PTBAssetsBagNew(ptb *sui2.ProgrammableTransactionBuilder, packageID sui2.PackageID, owner *cryptolib.Address) *sui2.ProgrammableTransactionBuilder {
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "new",
				TypeArguments: []sui2.TypeTag{},
				Arguments:     []sui2.Argument{},
			},
		},
	)
	return ptb
}

func PTBAssetsBagNewAndTransfer(ptb *sui2.ProgrammableTransactionBuilder, packageID sui2.PackageID, owner *cryptolib.Address) *sui2.ProgrammableTransactionBuilder {
	arg1 := ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "new",
				TypeArguments: []sui2.TypeTag{},
				Arguments:     []sui2.Argument{},
			},
		},
	)
	ptb.Command(
		sui2.Command{
			TransferObjects: &sui2.ProgrammableTransferObjects{
				Objects: []sui2.Argument{arg1},
				Address: ptb.MustForceSeparatePure(owner.AsSuiAddress()),
			},
		},
	)
	return ptb
}

func PTBAssetsBagPlaceCoin(
	ptb *sui2.ProgrammableTransactionBuilder,
	packageID sui2.PackageID,
	argAssetsBag sui2.Argument,
	argCoin sui2.Argument,
	coinType string,
) *sui2.ProgrammableTransactionBuilder {
	typeTag, err := sui2.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin",
				TypeArguments: []sui2.TypeTag{*typeTag},
				Arguments: []sui2.Argument{
					argAssetsBag,
					argCoin,
				},
			},
		},
	)
	return ptb
}

func PTBAssetsBagPlaceCoinBalance(ptb *sui2.ProgrammableTransactionBuilder, packageID sui2.PackageID, argAssetsBag sui2.Argument, argCoinBalance sui2.Argument, coinType string) *sui2.ProgrammableTransactionBuilder {
	typeTag, err := sui2.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin_balance",
				TypeArguments: []sui2.TypeTag{*typeTag},
				Arguments: []sui2.Argument{
					argAssetsBag,
					argCoinBalance,
				},
			},
		},
	)
	return ptb
}

func PTBAssetsBagPlaceCoinWithAmount(ptb *sui2.ProgrammableTransactionBuilder, packageID sui2.PackageID, assetsBagRef *sui2.ObjectRef, coin *sui2.ObjectRef, amount uint64, coinType string) *sui2.ProgrammableTransactionBuilder {
	typeTag, err := sui2.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	splitCoinArg := ptb.Command(
		sui2.Command{
			SplitCoins: &sui2.ProgrammableSplitCoins{
				Coin:    ptb.MustObj(sui2.ObjectArg{ImmOrOwnedObject: coin}),
				Amounts: []sui2.Argument{ptb.MustForceSeparatePure(amount)},
			},
		},
	)
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin",
				TypeArguments: []sui2.TypeTag{*typeTag},
				Arguments: []sui2.Argument{
					ptb.MustObj(sui2.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
					splitCoinArg,
				},
			},
		},
	)
	return ptb
}

func PTBAssetsBagTakeCoinBalance(ptb *sui2.ProgrammableTransactionBuilder, packageID sui2.PackageID, argAssetsBag sui2.Argument, amount uint64, coinType string) *sui2.ProgrammableTransactionBuilder {
	typeTag, err := sui2.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "take_coin_balance",
				TypeArguments: []sui2.TypeTag{*typeTag},
				Arguments: []sui2.Argument{
					argAssetsBag,
					ptb.MustForceSeparatePure(amount),
				},
			},
		},
	)
	return ptb
}

func PTBAssetsDestroyEmpty(ptb *sui2.ProgrammableTransactionBuilder, packageID sui2.PackageID, argAssetsBag sui2.Argument) *sui2.ProgrammableTransactionBuilder {
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "destroy_empty",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
					argAssetsBag,
				},
			},
		},
	)
	return ptb
}
