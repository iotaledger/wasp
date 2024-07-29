package iscmove

import (
	"context"
	"fmt"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/types"
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
) (*types.RefWithObject[types.Anchor], error) {
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

	anchorRef, err := txnResponse.GetCreatedObjectInfo(types.AnchorModuleName, types.AnchorObjectName)
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
	anchorRef *sui.ObjectRef,
	reqs []sui.ObjectRef,
	stateRoot []byte,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	reqAssetsBagsMap := make(map[sui.ObjectRef]*types.AssetsBagWithBalances)
	for _, reqRef := range reqs {
		req, err := c.GetRequestFromObjectID(ctx, reqRef.ObjectID)
		if err != nil {
			return nil, err
		}
		assetsBag, err := c.GetAssetsBagWithBalances(ctx, &req.AssetsBag.Value.ID)
		if err != nil {
			return nil, err
		}
		reqAssetsBagsMap[reqRef] = assetsBag
	}

	ptb, err := NewReceiveRequestPTB(packageID, anchorRef, reqs, reqAssetsBagsMap, stateRoot)
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
) (*types.RefWithObject[types.Anchor], error) {
	getObjectResponse, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: anchorObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}

	var anchor types.Anchor
	err = suiclient.UnmarshalBCS(getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes, &anchor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	return &types.RefWithObject[types.Anchor]{
		ObjectRef: getObjectResponse.Data.Ref(),
		Object:    &anchor,
	}, nil
}

func (c *Client) GetRequestFromObjectID(
	ctx context.Context,
	id *sui.ObjectID,
) (*types.Request, error) {
	getObjectResponse, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: id,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}

	var req types.Request
	err = suiclient.UnmarshalBCS(getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}

	return &req, nil
}
