package iscmove

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove/iscmove_types"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func NewStartNewChainPTB(packageID sui.PackageID, initParams []byte, ownerAddress *cryptolib.Address) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()
	arg1 := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove_types.AnchorModuleName,
				Function:      "start_new_chain",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustPure(initParams),
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

	return ptb.Finish()
}

func NewReceiveRequestPTB(packageID sui.PackageID, anchorRef *sui.ObjectRef, requestObjects []sui.ObjectRef, reqAssetsBagsMap map[sui.ObjectRef]*iscmove_types.AssetsBagWithBalances, stateRoot []byte) (sui.ProgrammableTransaction, error) {
	ptb := sui.NewProgrammableTransactionBuilder()

	argAnchor := ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: anchorRef})
	typeReceipt, err := sui.TypeTagFromString(fmt.Sprintf("%s::%s::%s", packageID, iscmove_types.AnchorModuleName, iscmove_types.ReceiptObjectName))
	if err != nil {
		return sui.ProgrammableTransaction{}, fmt.Errorf("can't parse Receipt's TypeTag: %w", err)
	}

	var argReceiveRequests []sui.Argument
	argBorrowAssets := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove_types.AnchorModuleName,
				Function:      "borrow_assets",
				TypeArguments: []sui.TypeTag{},
				Arguments:     []sui.Argument{argAnchor},
			},
		},
	)
	argAnchorAssets := sui.Argument{NestedResult: &sui.NestedResult{Cmd: *argBorrowAssets.Result, Result: 0}}
	argAnchorBorrow := sui.Argument{NestedResult: &sui.NestedResult{Cmd: *argBorrowAssets.Result, Result: 1}}
	for _, reqObject := range requestObjects {
		argReqObject := ptb.MustObj(sui.ObjectArg{Receiving: &reqObject})
		argReceiveRequest := ptb.Command(
			sui.Command{
				MoveCall: &sui.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        iscmove_types.AnchorModuleName,
					Function:      "receive_request",
					TypeArguments: []sui.TypeTag{},
					Arguments:     []sui.Argument{argAnchor, argReqObject},
				},
			},
		)
		argReceiveRequests = append(argReceiveRequests, argReceiveRequest)

		assetsBag := reqAssetsBagsMap[reqObject]
		argAssetsBag := sui.Argument{NestedResult: &sui.NestedResult{Cmd: *argReceiveRequest.Result, Result: 1}}
		for _, bal := range assetsBag.Balances {
			typeTag, err := sui.TypeTagFromString(bal.CoinType)
			if err != nil {
				return sui.ProgrammableTransaction{}, fmt.Errorf("can't parse Balance's Coin TypeTag: %w", err)
			}
			argBal := ptb.Command(
				sui.Command{
					MoveCall: &sui.ProgrammableMoveCall{
						Package:       &packageID,
						Module:        iscmove_types.AssetsBagModuleName,
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
						Module:        iscmove_types.AssetsBagModuleName,
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
					Module:        iscmove_types.AssetsBagModuleName,
					Function:      "destroy_empty",
					TypeArguments: []sui.TypeTag{},
					Arguments:     []sui.Argument{argAssetsBag},
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
				Module:        iscmove_types.AnchorModuleName,
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
				Module:        iscmove_types.AnchorModuleName,
				Function:      "update_state_root",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					argAnchor,
					ptb.MustPure(stateRoot),
					argReceipts,
				},
			},
		},
	)

	return ptb.Finish(), nil
}
