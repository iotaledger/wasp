package iscmoveclient

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/util/bcs"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func (c *Client) StartNewChain(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID iotago.PackageID,
	stateMetadata []byte,
	initCoinRef *iotago.ObjectRef,
	gasObjAddr *iotago.Address,
	txFeePerReq uint64,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iscmove.AnchorWithRef, error) {
	var err error
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)

	ptb := iotago.NewProgrammableTransactionBuilder()
	var argInitCoin iotago.Argument
	if initCoinRef != nil {
		ptb = PTBOptionSomeIotaCoin(ptb, initCoinRef)
	} else {
		ptb = PTBOptionNoneIotaCoin(ptb)
	}
	argInitCoin = ptb.LastCommandResultArg()

	ptb = PTBStartNewChain(ptb, packageID, stateMetadata, argInitCoin, gasObjAddr, txFeePerReq, cryptolibSigner.Address())
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coinPage, err := c.GetCoins(ctx, iotaclient.GetCoinsRequest{Owner: signer.Address()})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = []*iotago.ObjectRef{coinPage.Data[0].Ref()}
	}

	tx := iotago.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)

	txnBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}
	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		signer,
		txnBytes,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
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

func (c *Client) ReceiveRequestsAndTransition(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID iotago.PackageID,
	anchorRef *iotago.ObjectRef,
	reqs []iotago.ObjectRef,
	stateMetadata []byte,
	gasPayments []*iotago.ObjectRef,
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	var err error
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)

	var reqAssetsBags []*iscmove.AssetsBagWithBalances
	for _, reqRef := range reqs {
		reqWithObj, err := c.GetRequestFromObjectID(ctx, reqRef.ObjectID)
		if err != nil {
			return nil, err
		}
		assetsBag, err := c.GetAssetsBagWithBalances(ctx, &reqWithObj.Object.AssetsBag.ID)
		if err != nil {
			return nil, err
		}
		reqAssetsBags = append(reqAssetsBags, assetsBag)
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBReceiveRequestsAndTransition(
		ptb,
		packageID,
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: anchorRef}),
		reqs,
		reqAssetsBags,
		stateMetadata,
	)
	pt := ptb.Finish()

	tx := iotago.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)

	txnBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}
	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		signer,
		txnBytes,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
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
	anchorObjectID *iotago.ObjectID,
) (*iscmove.AnchorWithRef, error) {
	getObjectResponse, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: anchorObjectID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}
	return decodeAnchorBCS(
		getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes,
		getObjectResponse.Data.Ref(),
	)
}

func (c *Client) GetPastAnchorFromObjectID(
	ctx context.Context,
	anchorObjectID *iotago.ObjectID,
	version uint64,
) (*iscmove.AnchorWithRef, error) {
	getObjectResponse, err := c.TryGetPastObject(ctx, iotaclient.TryGetPastObjectRequest{
		ObjectID: anchorObjectID,
		Version:  version,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}
	if getObjectResponse.Data.VersionFound == nil {
		return nil, fmt.Errorf("failed to get anchor content")
	}
	return decodeAnchorBCS(
		getObjectResponse.Data.VersionFound.Bcs.Data.MoveObject.BcsBytes,
		getObjectResponse.Data.VersionFound.Ref(),
	)
}

func decodeAnchorBCS(bcsBytes iotago.Base64Data, ref iotago.ObjectRef) (*iscmove.AnchorWithRef, error) {
	var moveAnchor moveAnchor
	err := iotaclient.UnmarshalBCS(bcsBytes, &moveAnchor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	return &iscmove.AnchorWithRef{
		ObjectRef: ref,
		Object:    moveAnchor.ToAnchor(),
	}, nil
}
