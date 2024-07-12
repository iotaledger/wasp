package iscmove

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func NewStartNewChainPTB(packageID sui.PackageID, initParams []byte, ownerAddress *cryptolib.Address) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()
	arg1 := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        AnchorModuleName,
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

func NewReceiveRequestPTB(packageID sui.PackageID, anchorID sui.ObjectRef, requestObjects []sui.ObjectRef, stateRoot []byte) (sui.ProgrammableTransaction, error) {
	ptb := sui.NewProgrammableTransactionBuilder()

	argAnchor := ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: &anchorID})
	typeReceipt, err := sui.TypeTagFromString(fmt.Sprintf("%s::%s::%s", packageID, AnchorModuleName, ReceiptObjectName))
	if err != nil {
		return sui.ProgrammableTransaction{}, fmt.Errorf("can't parse Receipt's TypeTag: %w", err)
	}

	for i, reqObject := range requestObjects {
		argReqObject := ptb.MustObj(sui.ObjectArg{Receiving: &reqObject})
		ptb.Command(
			sui.Command{
				MoveCall: &sui.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        AnchorModuleName,
					Function:      "receive_request",
					TypeArguments: []sui.TypeTag{},
					Arguments:     []sui.Argument{argAnchor, argReqObject},
				},
			},
		)
		ptb.Command(
			sui.Command{
				TransferObjects: &sui.ProgrammableTransferObjects{
					Objects: []sui.Argument{
						{NestedResult: &sui.NestedResult{Cmd: uint16(i * 2), Result: 1}},
					},
					Address: ptb.MustPure(anchorID.ObjectID),
				},
			},
		)
	}

	var nestedResults []sui.Argument
	for i := 0; i < len(requestObjects); i++ {
		nestedResults = append(nestedResults, sui.Argument{NestedResult: &sui.NestedResult{Cmd: uint16(i * 2), Result: 0}})
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
				Module:        AnchorModuleName,
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
