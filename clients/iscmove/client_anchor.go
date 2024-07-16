package iscmove

import (
	"context"
	"encoding/json"
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
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := NewStartNewChainPTB(packageID, initParams, cryptolibSigner.Address())

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

	anchorRef, err := txnResponse.GetCreatedObjectInfo(AnchorModuleName, AnchorObjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to GetCreatedObjectInfo: %w", err)
	}
	anchor, err := c.GetAnchorFromObjectID(ctx, anchorRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAnchorFromObjectID: %w", err)
	}

	return anchor, nil
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

	reqAssetsBagsMap := make(map[sui.ObjectRef]*AssetsBag)
	for _, req := range reqObjects {
		assetsBag, err := c.GetAssetsBagFromRequestID(ctx, req.ObjectID)
		if err != nil {
			panic(err)
		}
		reqAssetsBagsMap[req] = assetsBag
	}

	ptb, err := NewReceiveRequestPTB(packageID, anchor, reqObjects, reqAssetsBagsMap, stateRoot)
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

func (c *Client) GetAnchorFromObjectID(
	ctx context.Context,
	anchorObjectID *sui.ObjectID,
) (*Anchor, error) {

	getObjectResponseAnchor, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: anchorObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowBcs: true, ShowContent: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}

	var tmpAnchorJsonObject anchorJsonObject
	err = json.Unmarshal(getObjectResponseAnchor.Data.Content.Data.MoveObject.Fields, &tmpAnchorJsonObject)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields in Anchor: %w", err)
	}

	resGetObject, err := c.GetObject(ctx,
		suiclient.GetObjectRequest{ObjectID: tmpAnchorJsonObject.ID.ID, Options: &suijsonrpc.SuiObjectDataOptions{ShowType: true}})
	if err != nil {
		return nil, fmt.Errorf("failed to get Anchor object: %w", err)
	}
	anchorRef := resGetObject.Data.Ref()
	assets, err := c.GetAssetsBagFromAnchorID(ctx, anchorRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetsBag from Anchor: %w", err)
	}
	anchor := Anchor{
		Ref: &anchorRef,
		Assets: Referent[AssetsBag]{
			ID:    *tmpAnchorJsonObject.Assets.Fields.ID,
			Value: assets,
		},
		InitParams: tmpAnchorJsonObject.InitParams,
		StateRoot:  tmpAnchorJsonObject.StateRoot,
		StateIndex: tmpAnchorJsonObject.StateIndex,
	}

	return &anchor, nil
}
