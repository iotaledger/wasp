package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func NewAllowanceNewPTB(packageID sui.PackageID, owner *cryptolib.Address) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()
	arg1 := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AllowanceModuleName,
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

func NewAllowanceWithCoinBalancePTB(packageID sui.PackageID, allowanceRef *sui.ObjectRef, balanceVal uint64, coinType string) (sui.ProgrammableTransaction, error) {
	ptb := sui.NewProgrammableTransactionBuilder()
	typeTag, err := sui.TypeTagFromString(coinType)
	if err != nil {
		return sui.ProgrammableTransaction{}, fmt.Errorf("failed to parse TypeTag: %s: %w", coinType, err)
	}

	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AllowanceModuleName,
				Function:      "with_coin_allowance",
				TypeArguments: []sui.TypeTag{*typeTag},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: allowanceRef}),
					ptb.MustPure(balanceVal),
				},
			},
		},
	)

	return ptb.Finish(), nil
}

func NewAllowanceDestroyPTB(packageID sui.PackageID, allowanceRef *sui.ObjectRef) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AllowanceModuleName,
				Function:      "destroy",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: allowanceRef}),
				},
			},
		},
	)

	return ptb.Finish()
}

func NewCreateBalancePTB(coinType suijsonrpc.CoinType, val uint64) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()
	coinTypeTag, err := sui.TypeTagFromString(coinType)
	if err != nil {
		panic(err)
	}
	supplyArg := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       sui.SuiPackageIdSuiFramework,
				Module:        "balance",
				Function:      "create_supply",
				TypeArguments: []sui.TypeTag{*coinTypeTag},
				Arguments:     []sui.Argument{},
			},
		},
	)

	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       sui.SuiPackageIdSuiFramework,
				Module:        "balance",
				Function:      "create_supply",
				TypeArguments: []sui.TypeTag{*coinTypeTag},
				Arguments: []sui.Argument{
					supplyArg,
					ptb.MustPure(val),
				},
			},
		},
	)
	return ptb.Finish()
}
