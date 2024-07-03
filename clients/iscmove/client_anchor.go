package iscmove

import (
	"context"
	"errors"
	"fmt"

	"github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (c *Client) StartNewChain(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID *sui.PackageID,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
	treasuryCap *suijsonrpc.SuiObjectResponse,
) (*Anchor, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := sui.NewProgrammableTransactionBuilder()
	arguments := []sui.Argument{}
	if treasuryCap != nil {
		ref := treasuryCap.Data.Ref()
		arguments = []sui.Argument{
			ptb.MustObj(
				sui.ObjectArg{
					ImmOrOwnedObject: &ref,
				},
			),
		}
	}
	arg1 := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "anchor",
				Function:      "start_new_chain",
				TypeArguments: []sui.TypeTag{},
				Arguments:     arguments,
			},
		},
	)
	ptb.Command(
		sui.Command{
			TransferObjects: &sui.ProgrammableTransferObjects{
				Objects: []sui.Argument{arg1},
				Address: ptb.MustPure(signer.Address()),
			},
		},
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}

	return c.getAnchorFromSuiTransactionBlockResponse(ctx, txnResponse)
}

func (c *Client) ReceiveAndUpdateStateRootRequest(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID *sui.PackageID,
	anchor *sui.ObjectRef,
	reqObject *sui.ObjectRef,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)
	ptb := sui.NewProgrammableTransactionBuilder()

	argAnchor := ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: anchor})
	argReqObject := ptb.MustObj(sui.ObjectArg{Receiving: reqObject})
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "anchor",
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
					{NestedResult: &sui.NestedResult{Cmd: 0, Result: 1}},
				},
				Address: ptb.MustPure(signer.Address()),
			},
		},
	)
	typeReceipt, err := sui.TypeTagFromString(fmt.Sprintf("%s::anchor::Receipt", packageID))
	if err != nil {
		return nil, fmt.Errorf("can't parse Receipt's TypeTag: %w", err)
	}
	argReceipts := ptb.Command(sui.Command{
		MakeMoveVec: &sui.ProgrammableMakeMoveVec{
			Type: typeReceipt,
			Objects: []sui.Argument{
				{NestedResult: &sui.NestedResult{Cmd: 0, Result: 0}},
			},
		},
	})
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "anchor",
				Function:      "update_state_root",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					argAnchor,
					ptb.MustPure([]byte{1, 2, 3}),
					argReceipts,
				},
			},
		},
	)
	pt := ptb.Finish()
	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}
	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	return txnResponse, nil
}

func (c *Client) getAnchorFromSuiTransactionBlockResponse(
	ctx context.Context,
	response *suijsonrpc.SuiTransactionBlockResponse,
) (*Anchor, error) {
	anchorObjRef, err := response.GetCreatedObjectInfo("anchor", "Anchor")
	if err != nil {
		return nil, fmt.Errorf("failed to GetCreatedObjectInfo: %w", err)
	}

	getObjectResponse, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: anchorObjRef.ObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}
	anchorBCS := getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes

	anchor := Anchor{}
	n, err := bcs.Unmarshal(anchorBCS, &anchor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	if n != len(anchorBCS) {
		return nil, errors.New("cannot decode anchor: excess bytes")
	}
	return &anchor, nil
}
