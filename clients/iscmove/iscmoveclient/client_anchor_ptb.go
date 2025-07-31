package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

func PTBUpdateAnchorStateMetadata(ptb *iotago.ProgrammableTransactionBuilder, packageID iotago.PackageID, anchorRef iotago.Argument, stateMetadata []byte, stateIndex uint32) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "update_anchor_state_for_migration",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					anchorRef,
					ptb.MustPure(stateMetadata),
					ptb.MustPure(stateIndex),
				},
			},
		},
	)
	return ptb
}

func PTBStartNewChain(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	stateMetadata []byte,
	argInitCoin iotago.Argument,
	anchorOwner *cryptolib.Address,
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
				Address: ptb.MustPure(anchorOwner.AsIotaAddress()),
			},
		},
	)
	return ptb
}

func PTBTakeAndTransferAssets(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	argAnchor iotago.Argument,
	target *iotago.Address,
	assets *iscmove.Assets,
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

	for coinType, coinAmount := range assets.Coins.Iterate() {
		argBal := ptb.Command(
			iotago.Command{
				MoveCall: &iotago.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AssetsBagModuleName,
					Function:      "take_coin_balance",
					TypeArguments: []iotago.TypeTag{coinType.TypeTag()},
					Arguments: []iotago.Argument{
						argAssets,
						ptb.MustPure(coinAmount),
					},
				},
			},
		)
		argTransferCoin := ptb.Command(
			iotago.Command{
				MoveCall: &iotago.ProgrammableMoveCall{
					Package:       iotago.IotaPackageIDIotaFramework,
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
	for objID, objType := range assets.Objects.Iterate() {
		argObj := ptb.Command(
			iotago.Command{
				MoveCall: &iotago.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AssetsBagModuleName,
					Function:      "take_asset",
					TypeArguments: []iotago.TypeTag{objType.TypeTag()},
					Arguments: []iotago.Argument{
						argAssets,
						ptb.MustForceSeparatePure(objID),
					},
				},
			},
		)
		ptb.Command(
			iotago.Command{
				TransferObjects: &iotago.ProgrammableTransferObjects{
					Objects: []iotago.Argument{argObj},
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

type ConsumedRequest struct {
	RequestRef iotago.ObjectRef
	Assets     *iscmove.AssetsBagWithBalances
}

type SentAssets struct {
	Target iotago.Address
	Assets iscmove.Assets
}

//nolint:funlen
func PTBReceiveRequestsAndTransition(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	argAnchor iotago.Argument,
	consumedRequests []ConsumedRequest,
	sentAssets []SentAssets,
	stateMetadata []byte,
	topUpAmount uint64,
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
	for _, consumed := range consumedRequests {
		argReqObject := ptb.MustObj(iotago.ObjectArg{Receiving: &consumed.RequestRef})
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

		argAssetsBag := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argReceiveRequest.Result, Result: 1}}
		for coinType := range consumed.Assets.Coins.Iterate() {
			argBal := ptb.Command(
				iotago.Command{
					MoveCall: &iotago.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove.AssetsBagModuleName,
						Function:      "take_all_coin_balance",
						TypeArguments: []iotago.TypeTag{coinType.TypeTag()},
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
						TypeArguments: []iotago.TypeTag{coinType.TypeTag()},
						Arguments:     []iotago.Argument{argAnchorAssets, argBal},
					},
				},
			)
		}
		for objID, objType := range consumed.Assets.Objects.Iterate() {
			obj := ptb.Command(
				iotago.Command{
					MoveCall: &iotago.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove.AssetsBagModuleName,
						Function:      "take_asset",
						TypeArguments: []iotago.TypeTag{objType.TypeTag()},
						Arguments:     []iotago.Argument{argAssetsBag, ptb.MustPure(objID)},
					},
				},
			)
			ptb.Command(
				iotago.Command{
					MoveCall: &iotago.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove.AssetsBagModuleName,
						Function:      "place_asset",
						TypeArguments: []iotago.TypeTag{objType.TypeTag()},
						Arguments:     []iotago.Argument{argAnchorAssets, obj},
					},
				},
			)
		}
		PTBAssetsDestroyEmpty(ptb, packageID, argAssetsBag)
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
	// top up gas coin

	if topUpAmount > 0 {
		ptb = PTBAssetsBagTakeCoinBalanceMergeTo(
			ptb,
			packageID,
			argAnchorAssets,
			topUpAmount,
			iotajsonrpc.IotaCoinType,
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
					argAnchorAssets,
					argAnchorBorrow,
				},
			},
		},
	)

	for _, sent := range sentAssets {
		ptb = PTBTakeAndTransferAssets(ptb, packageID, argAnchor, &sent.Target, &sent.Assets)
	}

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
