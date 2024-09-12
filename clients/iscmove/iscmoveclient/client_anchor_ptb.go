package iscmoveclient

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func PTBStartNewChain(
	ptb *sui.ProgrammableTransactionBuilder,
	packageID sui.PackageID,
	stateMetadata []byte,
	ownerAddress *cryptolib.Address,
) *sui.ProgrammableTransactionBuilder {
	arg1 := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "start_new_chain",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustPure(stateMetadata),
				},
			},
		},
	)
	ptb.Command(
		sui.Command{
			TransferObjects: &sui.ProgrammableTransferObjects{
				Objects: []sui.Argument{arg1},
				Address: ptb.MustPure(ownerAddress.AsSuiAddress()),
			},
		},
	)
	return ptb
}

func PTBReceiveRequestAndTransition(
	ptb *sui.ProgrammableTransactionBuilder,
	packageID sui.PackageID,
	argAnchor sui.Argument,
	requestRefs []sui.ObjectRef,
	reqAssetsBagsMap map[sui.ObjectRef]*iscmove.AssetsBagWithBalances,
	stateMetadata []byte,
) *sui.ProgrammableTransactionBuilder {
	typeReceipt, err := sui.TypeTagFromString(fmt.Sprintf("%s::%s::%s", packageID, iscmove.AnchorModuleName, iscmove.ReceiptObjectName))
	if err != nil {
		panic(fmt.Sprintf("can't parse Receipt's TypeTag: %s", err))
	}

	var argReceiveRequests []sui.Argument
	argBorrowAssets := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "borrow_assets",
				TypeArguments: []sui.TypeTag{},
				Arguments:     []sui.Argument{argAnchor},
			},
		},
	)
	argAnchorAssets := sui.Argument{NestedResult: &sui.NestedResult{Cmd: *argBorrowAssets.Result, Result: 0}}
	argAnchorBorrow := sui.Argument{NestedResult: &sui.NestedResult{Cmd: *argBorrowAssets.Result, Result: 1}}
	for _, reqObject := range requestRefs {
		reqObject := reqObject
		argReqObject := ptb.MustObj(sui.ObjectArg{Receiving: &reqObject})
		argReceiveRequest := ptb.Command(
			sui.Command{
				MoveCall: &sui.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AnchorModuleName,
					Function:      "receive_request",
					TypeArguments: []sui.TypeTag{},
					Arguments:     []sui.Argument{argAnchor, argReqObject},
				},
			},
		)
		argReceiveRequests = append(argReceiveRequests, argReceiveRequest)

		assetsBag := reqAssetsBagsMap[reqObject]
		argAssetsBag := sui.Argument{NestedResult: &sui.NestedResult{Cmd: *argReceiveRequest.Result, Result: 1}}
		argAllowance := sui.Argument{NestedResult: &sui.NestedResult{Cmd: *argReceiveRequest.Result, Result: 2}}
		for _, bal := range assetsBag.Balances {
			typeTag, err := sui.TypeTagFromString(bal.CoinType)
			if err != nil {
				panic(fmt.Sprintf("can't parse Balance's Coin TypeTag: %s", err))
			}
			argBal := ptb.Command(
				sui.Command{
					MoveCall: &sui.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove.AssetsBagModuleName,
						Function:      "take_all_coin_balance",
						TypeArguments: []sui.TypeTag{*typeTag},
						Arguments:     []sui.Argument{argAssetsBag},
					},
				},
			)
			ptb.Command(
				sui.Command{
					MoveCall: &sui.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove.AssetsBagModuleName,
						Function:      "place_coin_balance",
						TypeArguments: []sui.TypeTag{*typeTag},
						Arguments:     []sui.Argument{argAnchorAssets, argBal},
					},
				},
			)
		}
		ptb.Command(
			sui.Command{
				MoveCall: &sui.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AssetsBagModuleName,
					Function:      "destroy_empty",
					TypeArguments: []sui.TypeTag{},
					Arguments:     []sui.Argument{argAssetsBag},
				},
			},
		)
		ptb.Command(
			sui.Command{
				MoveCall: &sui.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove.AllowanceModuleName,
					Function:      "destroy",
					TypeArguments: []sui.TypeTag{},
					Arguments:     []sui.Argument{argAllowance},
				},
			},
		)
	}

	var nestedResults []sui.Argument
	for _, argReceiveRequest := range argReceiveRequests {
		nestedResults = append(nestedResults, sui.Argument{NestedResult: &sui.NestedResult{Cmd: *argReceiveRequest.Result, Result: 0}})
	}
	argReceipts := ptb.Command(sui.Command{
		MakeMoveVec: &sui.ProgrammableMakeMoveVec{
			Type:    typeReceipt,
			Objects: nestedResults,
		},
	})

	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "return_assets_from_borrow",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					argAnchor,
					argAnchorAssets,
					argAnchorBorrow,
				},
			},
		},
	)

	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.AnchorModuleName,
				Function:      "transition",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					argAnchor,
					ptb.MustPure(stateMetadata),
					argReceipts,
				},
			},
		},
	)
	return ptb
}
