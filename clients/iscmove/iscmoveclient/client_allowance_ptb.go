package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func PTBAllowanceNew(ptb *sui.ProgrammableTransactionBuilder, packageID sui.PackageID, owner *cryptolib.Address) *sui.ProgrammableTransactionBuilder {
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
	return ptb
}

func PTBAllowanceWithCoinBalance(ptb *sui.ProgrammableTransactionBuilder, packageID sui.PackageID, argAllowance sui.Argument, balanceVal uint64, coinType string) *sui.ProgrammableTransactionBuilder {
	typeTag, err := sui.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}

	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AllowanceModuleName,
				Function:      "with_coin_allowance",
				TypeArguments: []sui.TypeTag{*typeTag},
				Arguments: []sui.Argument{
					argAllowance,
					ptb.MustPure(balanceVal),
				},
			},
		},
	)
	return ptb
}

func PTBAllowanceDestroy(ptb *sui.ProgrammableTransactionBuilder, packageID sui.PackageID, argAllowance sui.Argument) *sui.ProgrammableTransactionBuilder {
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AllowanceModuleName,
				Function:      "destroy",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					argAllowance,
				},
			},
		},
	)
	return ptb
}

func PTBCreateBalance(ptb *sui.ProgrammableTransactionBuilder, coinType suijsonrpc.CoinType, val uint64) *sui.ProgrammableTransactionBuilder {
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
	return ptb
}
