package iscmove

import (
	"context"
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
) (*Anchor, error) {
	var err error

	ptb := NewStartNewChainPTB(packageID, initParams, cryptolibSigner.Address())
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		ptb,
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
	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		signer,
		txnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}
	return c.GetAnchorFromSuiTransactionBlockResponse(ctx, txnResponse)
}

func (c *Client) ReceiveAndUpdateStateRootRequest(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	anchor sui.ObjectRef,
	reqObjects []sui.ObjectRef,
	stateRoot []byte,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)
	ptb, err := NewReceiveRequestPTB(packageID, anchor, reqObjects, stateRoot)
	if err != nil {
		return nil, err
	}

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}
	tx := sui.NewProgrammable(
		signer.Address(),
		ptb,
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
	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		signer,
		txnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}
	return txnResponse, nil
}

type MoveAnchor struct {
	ID         *sui.ObjectID
	Assets     Referent[MoveAssetsBag]
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

	anchor, err := c.GetAnchorFromObjectID(ctx, anchorObjRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAnchorFromObjectID: %w", err)
	}

	return anchor, nil
}

func (c *Client) GetAnchorFromObjectID(
	ctx context.Context,
	anchorObjectID *sui.ObjectID,
) (*Anchor, error) {

	getObjectResponse, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: anchorObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}
	anchorBCS := getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes

	_anchor := MoveAnchor{}
	n, err := bcs.Unmarshal(anchorBCS, &_anchor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	if n != len(anchorBCS) {
		// FIXME this currently can't pass
		// return nil, errors.New("cannot decode anchor: excess bytes")
	}

	resGetObject, err := c.GetObject(ctx,
		suiclient.GetObjectRequest{ObjectID: _anchor.ID, Options: &suijsonrpc.SuiObjectDataOptions{ShowType: true}})
	if err != nil {
		return nil, fmt.Errorf("failed to get Anchor object: %w", err)
	}
	anchorRef := resGetObject.Data.Ref()
	assets, err := c.GetAssetsBagFromAnchor(ctx, anchorRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetsBag from Anchor: %w", err)
	}
	anchor := Anchor{
		Ref: &anchorRef,
		Assets: Referent[AssetsBag]{
			ID:    _anchor.Assets.ID,
			Value: assets,
		},
		InitParams: _anchor.InitParams,
		StateRoot:  _anchor.StateRoot,
		StateIndex: _anchor.StateIndex,
	}

	return &anchor, nil
}
