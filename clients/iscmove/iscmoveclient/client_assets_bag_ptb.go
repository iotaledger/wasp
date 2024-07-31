package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func NewAssetsBagNewPTB(packageID sui.PackageID, owner *cryptolib.Address) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()
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

	return ptb.Finish()
}

func NewAssetsBagPlaceCoinPTB(packageID sui.PackageID, assetsBagRef *sui.ObjectRef, coin *sui.ObjectRef, coinType string) (sui.ProgrammableTransaction, error) {
	ptb := sui.NewProgrammableTransactionBuilder()
	typeTag, err := sui.TypeTagFromString(coinType)
	if err != nil {
		return sui.ProgrammableTransaction{}, fmt.Errorf("failed to parse TypeTag: %s: %w", coinType, err)
	}
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin",
				TypeArguments: []sui.TypeTag{*typeTag},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: coin}),
				},
			},
		},
	)

	return ptb.Finish(), nil
}

func NewAssetsDestroyEmptyPTB(packageID sui.PackageID, assetsBagRef *sui.ObjectRef) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "destroy_empty",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
				},
			},
		},
	)

	return ptb.Finish()
}
