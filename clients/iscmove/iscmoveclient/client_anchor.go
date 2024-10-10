package iscmoveclient

import (
	"context"
	"fmt"

	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	suiclient2 "github.com/iotaledger/wasp/clients/iota-go/suiclient"
	suijsonrpc2 "github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	"github.com/iotaledger/wasp/packages/util/bcs"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func (c *Client) StartNewChain(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID sui2.PackageID,
	stateMetadata []byte,
	initCoinRef *sui2.ObjectRef,
	gasPayments []*sui2.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*iscmove.AnchorWithRef, error) {
	var err error
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	ptb := sui2.NewProgrammableTransactionBuilder()
	var argInitCoin sui2.Argument
	if initCoinRef != nil {
		ptb = PTBOptionSomeSuiCoin(ptb, initCoinRef)
	} else {
		ptb = PTBOptionNoneSuiCoin(ptb)
	}
	argInitCoin = ptb.LastCommandResultArg()

	ptb = PTBStartNewChain(ptb, packageID, stateMetadata, argInitCoin, cryptolibSigner.Address())
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coinPage, err := c.GetCoins(ctx, suiclient2.GetCoinsRequest{Owner: signer.Address()})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = []*sui2.ObjectRef{coinPage.Data[0].Ref()}
	}

	tx := sui2.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)

	var txnBytes []byte
	if devMode {
		txnBytes, err = bcs.Marshal(&tx.V1.Kind)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	} else {
		txnBytes, err = bcs.Marshal(&tx)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	}
	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		signer,
		txnBytes,
		&suijsonrpc2.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
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
	cryptolibSigner cryptolib.Signer,
	packageID sui2.PackageID,
	anchorRef *sui2.ObjectRef,
	reqs []sui2.ObjectRef,
	stateMetadata []byte,
	gasPayments []*sui2.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	devMode bool,
) (*suijsonrpc2.SuiTransactionBlockResponse, error) {
	var err error
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)

	reqAssetsBagsMap := make(map[sui2.ObjectRef]*iscmove.AssetsBagWithBalances)
	for _, reqRef := range reqs {
		reqWithObj, err := c.GetRequestFromObjectID(ctx, reqRef.ObjectID)
		if err != nil {
			return nil, err
		}
		assetsBag, err := c.GetAssetsBagWithBalances(ctx, &reqWithObj.Object.AssetsBag.ID)
		if err != nil {
			return nil, err
		}
		reqAssetsBagsMap[reqRef] = assetsBag
	}

	ptb := sui2.NewProgrammableTransactionBuilder()
	ptb = PTBReceiveRequestAndTransition(ptb, packageID, ptb.MustObj(sui2.ObjectArg{ImmOrOwnedObject: anchorRef}), reqs, reqAssetsBagsMap, stateMetadata)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coinPage, err := c.GetCoins(ctx, suiclient2.GetCoinsRequest{Owner: signer.Address()})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = []*sui2.ObjectRef{coinPage.Data[0].Ref()}
	}
	tx := sui2.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)

	var txnBytes []byte
	if devMode {
		txnBytes, err = bcs.Marshal(&tx.V1.Kind)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	} else {
		txnBytes, err = bcs.Marshal(&tx)
		if err != nil {
			return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
		}
	}
	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		signer,
		txnBytes,
		&suijsonrpc2.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
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
	anchorObjectID *sui2.ObjectID,
) (*iscmove.AnchorWithRef, error) {
	getObjectResponse, err := c.GetObject(ctx, suiclient2.GetObjectRequest{
		ObjectID: anchorObjectID,
		Options:  &suijsonrpc2.SuiObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}

	var moveAnchor moveAnchor
	err = suiclient2.UnmarshalBCS(getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes, &moveAnchor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	return &iscmove.AnchorWithRef{
		ObjectRef: getObjectResponse.Data.Ref(),
		Object:    moveAnchor.ToAnchor(),
	}, nil
}
