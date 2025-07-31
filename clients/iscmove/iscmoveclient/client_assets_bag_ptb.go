package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

func PTBAssetsBagNew(ptb *iotago.ProgrammableTransactionBuilder, packageID iotago.PackageID, owner *cryptolib.Address) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "new",
				TypeArguments: []iotago.TypeTag{},
				Arguments:     []iotago.Argument{},
			},
		},
	)
	return ptb
}

func PTBAssetsBagNewAndTransfer(ptb *iotago.ProgrammableTransactionBuilder, packageID iotago.PackageID, owner *cryptolib.Address) *iotago.ProgrammableTransactionBuilder {
	arg1 := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "new",
				TypeArguments: []iotago.TypeTag{},
				Arguments:     []iotago.Argument{},
			},
		},
	)
	ptb.Command(
		iotago.Command{
			TransferObjects: &iotago.ProgrammableTransferObjects{
				Objects: []iotago.Argument{arg1},
				Address: ptb.MustForceSeparatePure(owner.AsIotaAddress()),
			},
		},
	)
	return ptb
}

func PTBAssetsBagPlaceCoin(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	argAssetsBag iotago.Argument,
	argCoin iotago.Argument,
	coinType iotajsonrpc.CoinType,
) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin",
				TypeArguments: []iotago.TypeTag{coinType.TypeTag()},
				Arguments: []iotago.Argument{
					argAssetsBag,
					argCoin,
				},
			},
		},
	)
	return ptb
}

func PTBAssetsBagPlaceCoinBalance(ptb *iotago.ProgrammableTransactionBuilder, packageID iotago.PackageID, argAssetsBag iotago.Argument, argCoinBalance iotago.Argument, coinType string) *iotago.ProgrammableTransactionBuilder {
	typeTag, err := iotago.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin_balance",
				TypeArguments: []iotago.TypeTag{*typeTag},
				Arguments: []iotago.Argument{
					argAssetsBag,
					argCoinBalance,
				},
			},
		},
	)
	return ptb
}

func PTBAssetsBagPlaceCoinWithAmount(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	argAssetsBag iotago.Argument,
	argCoin iotago.Argument,
	amount iotajsonrpc.CoinValue,
	coinType iotajsonrpc.CoinType,
) *iotago.ProgrammableTransactionBuilder {
	splitCoinArg := ptb.Command(
		iotago.Command{
			SplitCoins: &iotago.ProgrammableSplitCoins{
				Coin:    argCoin,
				Amounts: []iotago.Argument{ptb.MustForceSeparatePure(amount.Uint64())},
			},
		},
	)
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin",
				TypeArguments: []iotago.TypeTag{coinType.TypeTag()},
				Arguments: []iotago.Argument{
					argAssetsBag,
					splitCoinArg,
				},
			},
		},
	)
	return ptb
}

func PTBAssetsBagPlaceObject(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	argAssetsBag iotago.Argument,
	argObject iotago.Argument,
	argObjectType iotago.ObjectType,
) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_asset",
				TypeArguments: []iotago.TypeTag{argObjectType.TypeTag()},
				Arguments: []iotago.Argument{
					argAssetsBag,
					argObject,
				},
			},
		},
	)
	return ptb
}

func PTBAssetsBagTakeCoinBalance(ptb *iotago.ProgrammableTransactionBuilder, packageID iotago.PackageID, argAssetsBag iotago.Argument, amount uint64, coinType string) *iotago.ProgrammableTransactionBuilder {
	typeTag, err := iotago.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "take_coin_balance",
				TypeArguments: []iotago.TypeTag{*typeTag},
				Arguments: []iotago.Argument{
					argAssetsBag,
					ptb.MustForceSeparatePure(amount),
				},
			},
		},
	)
	return ptb
}

func PTBAssetsDestroyEmpty(ptb *iotago.ProgrammableTransactionBuilder, packageID iotago.PackageID, argAssetsBag iotago.Argument) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "destroy_empty",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					argAssetsBag,
				},
			},
		},
	)
	return ptb
}

func PTBAssetsBagTakeCoinBalanceMergeTo(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	argAssetsBag iotago.Argument,
	amount uint64,
	coinType iotajsonrpc.CoinType,
) *iotago.ProgrammableTransactionBuilder {
	typeTag, err := iotago.TypeTagFromString(coinType.String())
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	takenBalArg := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "take_coin_balance",
				TypeArguments: []iotago.TypeTag{*typeTag},
				Arguments: []iotago.Argument{
					argAssetsBag,
					ptb.MustForceSeparatePure(amount),
				},
			},
		},
	)
	createdCoin := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       iotago.IotaPackageIDIotaFramework,
				Module:        "coin",
				Function:      "from_balance",
				TypeArguments: []iotago.TypeTag{*typeTag},
				Arguments: []iotago.Argument{
					takenBalArg,
				},
			},
		},
	)
	ptb.Command(
		iotago.Command{
			MergeCoins: &iotago.ProgrammableMergeCoins{
				Destination: iotago.GetArgumentGasCoin(),
				Sources:     []iotago.Argument{createdCoin},
			},
		},
	)
	return ptb
}
