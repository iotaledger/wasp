package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove"
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func PTBStartNewChain(
	ptb *sui2.ProgrammableTransactionBuilder,
	packageID sui2.PackageID,
	stateMetadata []byte,
	argInitCoin sui2.Argument,
	ownerAddress *cryptolib.Address,
) *sui2.ProgrammableTransactionBuilder {
	arg1 := ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "start_new_chain",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
					ptb.MustPure(stateMetadata),
					argInitCoin,
				},
			},
		},
	)
	ptb.Command(
		sui2.Command{
			TransferObjects: &sui2.ProgrammableTransferObjects{
				Objects: []sui2.Argument{arg1},
				Address: ptb.MustPure(ownerAddress.AsSuiAddress()),
			},
		},
	)
	return ptb
}

func PTBTakeAndTransferCoinBalance(
	ptb *sui2.ProgrammableTransactionBuilder,
	packageID sui2.PackageID,
	argAnchor sui2.Argument,
	target *sui2.Address,
	assets *isc.Assets,
) *sui2.ProgrammableTransactionBuilder {
	argBorrow := ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "borrow_assets",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
					argAnchor,
				},
			},
		},
	)
	argAssets := sui2.Argument{NestedResult: &sui2.NestedResult{Cmd: *argBorrow.Result, Result: 0}}
	argB := sui2.Argument{NestedResult: &sui2.NestedResult{Cmd: *argBorrow.Result, Result: 1}}

	for coinType, coinBalance := range assets.Coins {
		argBal := ptb.Command(
			sui2.Command{
				MoveCall: &sui2.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AssetsBagModuleName,
					Function:      "take_coin_balance",
					TypeArguments: []sui2.TypeTag{coinType.TypeTag()},
					Arguments: []sui2.Argument{
						argAssets,
						ptb.MustPure(coinBalance.Uint64()),
					},
				},
			},
		)
		argTransferCoin := ptb.Command(
			sui2.Command{
				MoveCall: &sui2.ProgrammableMoveCall{
					Package:       sui2.SuiPackageIdSuiFramework,
					Module:        "coin",
					Function:      "from_balance",
					TypeArguments: []sui2.TypeTag{coinType.TypeTag()},
					Arguments: []sui2.Argument{
						argBal,
					},
				},
			},
		)
		ptb.Command(
			sui2.Command{
				TransferObjects: &sui2.ProgrammableTransferObjects{
					Objects: []sui2.Argument{argTransferCoin},
					Address: ptb.MustForceSeparatePure(target),
				},
			},
		)
	}
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "return_assets_from_borrow",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
					argAnchor,
					argAssets,
					argB,
				},
			},
		},
	)
	return ptb
}

func PTBTakeAndPlaceToAssetsBag(
	ptb *sui2.ProgrammableTransactionBuilder,
	packageID sui2.PackageID,
	argAnchor sui2.Argument,
	argAssetsBag sui2.Argument,
	amount uint64,
	coinType string,
) *sui2.ProgrammableTransactionBuilder {
	typeTag, err := sui2.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	argBorrow := ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "borrow_assets",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
					argAnchor,
				},
			},
		},
	)
	argAssets := sui2.Argument{NestedResult: &sui2.NestedResult{Cmd: *argBorrow.Result, Result: 0}}
	argB := sui2.Argument{NestedResult: &sui2.NestedResult{Cmd: *argBorrow.Result, Result: 1}}
	argBal := ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "take_coin_balance",
				TypeArguments: []sui2.TypeTag{*typeTag},
				Arguments: []sui2.Argument{
					argAssets,
					ptb.MustPure(amount),
				},
			},
		},
	)
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin_balance",
				TypeArguments: []sui2.TypeTag{*typeTag},
				Arguments: []sui2.Argument{
					argAssetsBag,
					argBal,
				},
			},
		},
	)
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "return_assets_from_borrow",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
					argAnchor,
					argAssets,
					argB,
				},
			},
		},
	)
	return ptb
}

func PTBReceiveRequestAndTransition(
	ptb *sui2.ProgrammableTransactionBuilder,
	packageID sui2.PackageID,
	argAnchor sui2.Argument,
	requestRefs []sui2.ObjectRef,
	reqAssetsBagsMap map[sui2.ObjectRef]*iscmove.AssetsBagWithBalances,
	stateMetadata []byte,
) *sui2.ProgrammableTransactionBuilder {
	typeReceipt, err := sui2.TypeTagFromString(fmt.Sprintf("%s::%s::%s", packageID, iscmove.AnchorModuleName, iscmove.ReceiptObjectName))
	if err != nil {
		panic(fmt.Sprintf("can't parse Receipt's TypeTag: %s", err))
	}

	var argReceiveRequests []sui2.Argument
	argBorrowAssets := ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "borrow_assets",
				TypeArguments: []sui2.TypeTag{},
				Arguments:     []sui2.Argument{argAnchor},
			},
		},
	)
	argAnchorAssets := sui2.Argument{NestedResult: &sui2.NestedResult{Cmd: *argBorrowAssets.Result, Result: 0}}
	argAnchorBorrow := sui2.Argument{NestedResult: &sui2.NestedResult{Cmd: *argBorrowAssets.Result, Result: 1}}
	for _, reqObject := range requestRefs {
		reqObject := reqObject
		argReqObject := ptb.MustObj(sui2.ObjectArg{Receiving: &reqObject})
		argReceiveRequest := ptb.Command(
			sui2.Command{
				MoveCall: &sui2.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AnchorModuleName,
					Function:      "receive_request",
					TypeArguments: []sui2.TypeTag{},
					Arguments:     []sui2.Argument{argAnchor, argReqObject},
				},
			},
		)
		argReceiveRequests = append(argReceiveRequests, argReceiveRequest)

		assetsBag := reqAssetsBagsMap[reqObject]
		argAssetsBag := sui2.Argument{NestedResult: &sui2.NestedResult{Cmd: *argReceiveRequest.Result, Result: 1}}
		for _, bal := range assetsBag.Balances {
			typeTag, err := sui2.TypeTagFromString(bal.CoinType)
			if err != nil {
				panic(fmt.Sprintf("can't parse Balance's Coin TypeTag: %s", err))
			}
			argBal := ptb.Command(
				sui2.Command{
					MoveCall: &sui2.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove.AssetsBagModuleName,
						Function:      "take_all_coin_balance",
						TypeArguments: []sui2.TypeTag{*typeTag},
						Arguments:     []sui2.Argument{argAssetsBag},
					},
				},
			)
			ptb.Command(
				sui2.Command{
					MoveCall: &sui2.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove.AssetsBagModuleName,
						Function:      "place_coin_balance",
						TypeArguments: []sui2.TypeTag{*typeTag},
						Arguments:     []sui2.Argument{argAnchorAssets, argBal},
					},
				},
			)
		}
		ptb.Command(
			sui2.Command{
				MoveCall: &sui2.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AssetsBagModuleName,
					Function:      "destroy_empty",
					TypeArguments: []sui2.TypeTag{},
					Arguments:     []sui2.Argument{argAssetsBag},
				},
			},
		)
	}

	var nestedResults []sui2.Argument
	for _, argReceiveRequest := range argReceiveRequests {
		nestedResults = append(nestedResults, sui2.Argument{NestedResult: &sui2.NestedResult{Cmd: *argReceiveRequest.Result, Result: 0}})
	}
	argReceipts := ptb.Command(sui2.Command{
		MakeMoveVec: &sui2.ProgrammableMakeMoveVec{
			Type:    typeReceipt,
			Objects: nestedResults,
		},
	})

	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "return_assets_from_borrow",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
					argAnchor,
					argAnchorAssets,
					argAnchorBorrow,
				},
			},
		},
	)

	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "transition",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
					argAnchor,
					ptb.MustPure(stateMetadata),
					argReceipts,
				},
			},
		},
	)
	return ptb
}
