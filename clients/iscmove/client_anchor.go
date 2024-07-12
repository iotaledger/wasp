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
	packageID sui.PackageID,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	initParams []byte,
	devMode bool,
) ([]byte, error) {
	var err error
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

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

	var txnBytes []byte
	if devMode {
		txnBytes, err = bcs.Marshal(tx.V1.Kind)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	} else {
		txnBytes, err = bcs.Marshal(tx)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	}
	return txnBytes, nil
}

func (c *Client) ReceiveAndUpdateStateRootRequest(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	anchor *sui.ObjectRef,
	reqObjects []*sui.ObjectRef,
	stateRoot []byte,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) ([]byte, error) {
	panic("impl is wrong")
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)
	ptb := sui.NewProgrammableTransactionBuilder()

	argAnchor := ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: anchor})
	typeReceipt, err := sui.TypeTagFromString(fmt.Sprintf("%s::%s::%s", packageID, AnchorModuleName, ReceiptObjectName))
	if err != nil {
		return nil, fmt.Errorf("can't parse Receipt's TypeTag: %w", err)
	}

	for i, reqObject := range reqObjects {
		argReqObject := ptb.MustObj(sui.ObjectArg{Receiving: reqObject})
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
					Address: ptb.MustPure(anchor.ObjectID),
				},
			},
		)
	}

	var nestedResults []sui.Argument
	for i := 0; i < len(reqObjects); i++ {
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

	var txnBytes []byte
	if devMode {
		txnBytes, err = bcs.Marshal(tx.V1.Kind)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	} else {
		txnBytes, err = bcs.Marshal(tx)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	}
	return txnBytes, nil
}

type bcsAnchor struct {
	ID         *sui.ObjectID
	Assets     Referent[AssetBag]
	InitParams []byte
	StateRoot  sui.Bytes
	StateIndex uint32
}

func (c *Client) GetAnchorFromSuiTransactionBlockResponse(
	ctx context.Context,
	response *suijsonrpc.SuiTransactionBlockResponse,
) (*Anchor, error) {
	anchorObjRef, err := response.GetCreatedObjectInfo(AnchorModuleName, AnchorObjectName)
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

	_anchor := bcsAnchor{}
	n, err := bcs.Unmarshal(anchorBCS, &_anchor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	if n != len(anchorBCS) {
		return nil, errors.New("cannot decode anchor: excess bytes")
	}

	resGetObject, err := c.GetObject(ctx,
		suiclient.GetObjectRequest{ObjectID: _anchor.ID, Options: &suijsonrpc.SuiObjectDataOptions{ShowType: true}})
	if err != nil {
		return nil, fmt.Errorf("failed to get Anchor object: %w", err)
	}
	anchorRef := resGetObject.Data.Ref()
	anchor := Anchor{
		Ref:        &anchorRef,
		Assets:     _anchor.Assets,
		InitParams: _anchor.InitParams,
		StateRoot:  _anchor.StateRoot,
		StateIndex: _anchor.StateIndex,
	}

	return &anchor, nil
}
