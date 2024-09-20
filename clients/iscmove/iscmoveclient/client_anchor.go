package iscmoveclient

import (
	"context"
	"fmt"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func (c *Client) StartNewChain(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	stateMetadata []byte,
	initCoinRef *sui.ObjectRef,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*iscmove.AnchorWithRef, error) {
	var err error
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := sui.NewProgrammableTransactionBuilder()
	var argInitCoin sui.Argument
	if initCoinRef != nil {
		ptb = PTBOptionSomeSuiCoin(ptb, initCoinRef)
	} else {
		ptb = PTBOptionNoneSuiCoin(ptb)
	}
	argInitCoin = ptb.LastCommandResultArg()

	ptb = PTBStartNewChain(ptb, packageID, stateMetadata, argInitCoin, cryptolibSigner.Address())
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coinPage, err := c.GetCoins(ctx, suiclient.GetCoinsRequest{Owner: signer.Address()})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = []*sui.ObjectRef{coinPage.Data[0].Ref()}
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

	anchorRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AnchorModuleName, iscmove.AnchorObjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to GetCreatedObjectInfo: %w", err)
	}
	anchor, err := c.GetAnchorFromObjectID(ctx, anchorRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetAnchorFromObjectID: %w", err)
	}

	return anchor, nil
}

func (c *Client) ReceiveRequestAndTransition(
	ctx context.Context,
	ptb *sui.ProgrammableTransactionBuilder,
	cryptolibSigner cryptolib.Signer,
	packageID sui.PackageID,
	anchorRef *sui.ObjectRef,
	reqs []sui.ObjectRef,
	stateMetadata []byte,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	var err error
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	reqAssetsBagsMap := make(map[sui.ObjectRef]*iscmove.AssetsBagWithBalances)
	for _, reqRef := range reqs {
		reqWithObj, err := c.GetRequestFromObjectID(ctx, reqRef.ObjectID)
		if err != nil {
			return nil, err
		}
		assetsBag, err := c.GetAssetsBagWithBalances(ctx, &reqWithObj.Object.AssetsBag.Value.ID)
		if err != nil {
			return nil, err
		}
		reqAssetsBagsMap[reqRef] = assetsBag
	}

	ptb = PTBReceiveRequestAndTransition(ptb, packageID, ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: anchorRef}), reqs, reqAssetsBagsMap, stateMetadata)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coinPage, err := c.GetCoins(ctx, suiclient.GetCoinsRequest{Owner: signer.Address()})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = []*sui.ObjectRef{coinPage.Data[0].Ref()}
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
) (*iscmove.AnchorWithRef, error) {
	getObjectResponse, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: anchorObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}

	var anchor iscmove.Anchor
	err = suiclient.UnmarshalBCS(getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes, &anchor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	return &iscmove.AnchorWithRef{
		ObjectRef: getObjectResponse.Data.Ref(),
		Object:    &anchor,
	}, nil
}
