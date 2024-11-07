package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func PTBStartNewChain(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	stateMetadata []byte,
	argInitCoin iotago.Argument,
	ownerAddress *cryptolib.Address,
) *iotago.ProgrammableTransactionBuilder {
	arg1 := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "start_new_chain",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					ptb.MustPure(stateMetadata),
					argInitCoin,
				},
			},
		},
	)
	ptb.Command(
		iotago.Command{
			TransferObjects: &iotago.ProgrammableTransferObjects{
				Objects: []iotago.Argument{arg1},
				Address: ptb.MustPure(ownerAddress.AsIotaAddress()),
			},
		},
	)
	return ptb
}

func PTBTakeAndTransferCoinBalance(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	argAnchor iotago.Argument,
	target *iotago.Address,
	assets *isc.Assets,
) *iotago.ProgrammableTransactionBuilder {
	argBorrow := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "borrow_assets",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					argAnchor,
				},
			},
		},
	)
	argAssets := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argBorrow.Result, Result: 0}}
	argB := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argBorrow.Result, Result: 1}}

	for coinType, coinBalance := range assets.Coins {
		argBal := ptb.Command(
			iotago.Command{
				MoveCall: &iotago.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AssetsBagModuleName,
					Function:      "take_coin_balance",
					TypeArguments: []iotago.TypeTag{coinType.TypeTag()},
					Arguments: []iotago.Argument{
						argAssets,
						ptb.MustPure(coinBalance.Uint64()),
					},
				},
			},
		)
		argTransferCoin := ptb.Command(
			iotago.Command{
				MoveCall: &iotago.ProgrammableMoveCall{
					Package:       iotago.IotaPackageIdIotaFramework,
					Module:        "coin",
					Function:      "from_balance",
					TypeArguments: []iotago.TypeTag{coinType.TypeTag()},
					Arguments: []iotago.Argument{
						argBal,
					},
				},
			},
		)
		ptb.Command(
			iotago.Command{
				TransferObjects: &iotago.ProgrammableTransferObjects{
					Objects: []iotago.Argument{argTransferCoin},
					Address: ptb.MustForceSeparatePure(target),
				},
			},
		)
	}
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "return_assets_from_borrow",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
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
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	argAnchor iotago.Argument,
	argAssetsBag iotago.Argument,
	amount uint64,
	coinType string,
) *iotago.ProgrammableTransactionBuilder {
	typeTag, err := iotago.TypeTagFromString(coinType)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TypeTag: %s: %s", coinType, err))
	}
	argBorrow := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "borrow_assets",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					argAnchor,
				},
			},
		},
	)
	argAssets := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argBorrow.Result, Result: 0}}
	argB := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argBorrow.Result, Result: 1}}
	argBal := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "take_coin_balance",
				TypeArguments: []iotago.TypeTag{*typeTag},
				Arguments: []iotago.Argument{
					argAssets,
					ptb.MustPure(amount),
				},
			},
		},
	)
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AssetsBagModuleName,
				Function:      "place_coin_balance",
				TypeArguments: []iotago.TypeTag{*typeTag},
				Arguments: []iotago.Argument{
					argAssetsBag,
					argBal,
				},
			},
		},
	)
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "return_assets_from_borrow",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
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
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	argAnchor iotago.Argument,
	requestRefs []iotago.ObjectRef,
	requestAssets []*iscmove.AssetsBagWithBalances,
	stateMetadata []byte,
) *iotago.ProgrammableTransactionBuilder {
	typeReceipt, err := iotago.TypeTagFromString(fmt.Sprintf("%s::%s::%s", packageID, iscmove.AnchorModuleName, iscmove.ReceiptObjectName))
	if err != nil {
		panic(fmt.Sprintf("can't parse Receipt's TypeTag: %s", err))
	}

	var argReceiveRequests []iotago.Argument
	argBorrowAssets := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "borrow_assets",
				TypeArguments: []iotago.TypeTag{},
				Arguments:     []iotago.Argument{argAnchor},
			},
		},
	)
	argAnchorAssets := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argBorrowAssets.Result, Result: 0}}
	argAnchorBorrow := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argBorrowAssets.Result, Result: 1}}
	for i, reqObject := range requestRefs {
		reqObject := reqObject
		argReqObject := ptb.MustObj(iotago.ObjectArg{Receiving: &reqObject})
		argReceiveRequest := ptb.Command(
			iotago.Command{
				MoveCall: &iotago.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AnchorModuleName,
					Function:      "receive_request",
					TypeArguments: []iotago.TypeTag{},
					Arguments:     []iotago.Argument{argAnchor, argReqObject},
				},
			},
		)
		argReceiveRequests = append(argReceiveRequests, argReceiveRequest)

		assetsBag := requestAssets[i]
		argAssetsBag := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argReceiveRequest.Result, Result: 1}}
		for _, bal := range assetsBag.Balances {
			typeTag, err := iotago.TypeTagFromString(bal.CoinType)
			if err != nil {
				panic(fmt.Sprintf("can't parse Balance's Coin TypeTag: %s", err))
			}
			argBal := ptb.Command(
				iotago.Command{
					MoveCall: &iotago.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove.AssetsBagModuleName,
						Function:      "take_all_coin_balance",
						TypeArguments: []iotago.TypeTag{*typeTag},
						Arguments:     []iotago.Argument{argAssetsBag},
					},
				},
			)
			ptb.Command(
				iotago.Command{
					MoveCall: &iotago.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove.AssetsBagModuleName,
						Function:      "place_coin_balance",
						TypeArguments: []iotago.TypeTag{*typeTag},
						Arguments:     []iotago.Argument{argAnchorAssets, argBal},
					},
				},
			)
		}
		ptb.Command(
			iotago.Command{
				MoveCall: &iotago.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AssetsBagModuleName,
					Function:      "destroy_empty",
					TypeArguments: []iotago.TypeTag{},
					Arguments:     []iotago.Argument{argAssetsBag},
				},
			},
		)
	}

	var nestedResults []iotago.Argument
	for _, argReceiveRequest := range argReceiveRequests {
		nestedResults = append(nestedResults, iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argReceiveRequest.Result, Result: 0}})
	}
	argReceipts := ptb.Command(iotago.Command{
		MakeMoveVec: &iotago.ProgrammableMakeMoveVec{
			Type:    typeReceipt,
			Objects: nestedResults,
		},
	})

	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "return_assets_from_borrow",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					argAnchor,
					argAnchorAssets,
					argAnchorBorrow,
				},
			},
		},
	)

	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "transition",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					argAnchor,
					ptb.MustPure(stateMetadata),
					argReceipts,
				},
			},
		},
	)
	return ptb
}
