package iscmoveclient

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func (c *Client) FindCoinsForGasPayment(
	ctx context.Context,
	owner *iotago.Address,
	pt iotago.ProgrammableTransaction,
	gasPrice uint64,
	gasBudget uint64,
) ([]*iotago.ObjectRef, error) {
	coinType := iotajsonrpc.IotaCoinType
	coinPage, err := c.GetCoins(ctx, iotaclient.GetCoinsRequest{
		CoinType: &coinType,
		Owner:    owner,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch coins for gas payment: %w", err)
	}
	var gasPayments []*iotago.ObjectRef
	sum := uint64(0)
	for _, coin := range coinPage.Data {
		if pt.IsInInputObjects(coin.CoinObjectID) {
			continue
		}
		gasPayments = append(gasPayments, coin.Ref())
		sum += coin.Balance.Uint64()
		if sum >= gasPrice*gasBudget {
			break
		}
	}
	if sum < gasPrice*gasBudget {
		return nil, fmt.Errorf("not enough coins for gas payment")
	}
	return gasPayments, nil
}

func (c *Client) StartNewChain(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID iotago.PackageID,
	stateMetadata []byte,
	initCoinRef *iotago.ObjectRef,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iscmove.AnchorWithRef, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	var argInitCoin iotago.Argument
	if initCoinRef != nil {
		ptb = PTBOptionSomeIotaCoin(ptb, initCoinRef)
	} else {
		ptb = PTBOptionNoneIotaCoin(ptb)
	}
	argInitCoin = ptb.LastCommandResultArg()

	ptb = PTBStartNewChain(ptb, packageID, stateMetadata, argInitCoin, signer.Address())

	txnResponse, err := c.SignAndExecutePTB(
		ctx,
		signer,
		ptb.Finish(),
		gasPayments,
		gasPrice,
		gasBudget,
	)
	if err != nil {
		return nil, fmt.Errorf("start new chain PTB failed: %w", err)
	}

	anchorRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AnchorModuleName, iscmove.AnchorObjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to GetCreatedObjectInfo: %w", err)
	}
	return c.GetAnchorFromObjectID(ctx, anchorRef.ObjectID)
}

func (c *Client) ReceiveRequestsAndTransition(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID iotago.PackageID,
	anchorRef *iotago.ObjectRef,
	reqs []iotago.ObjectRef,
	stateMetadata []byte,
	gasPayments []*iotago.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
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
	return c.SignAndExecutePTB(
		ctx,
		signer,
		ptb.Finish(),
		gasPayments,
		gasPrice,
		gasBudget,
	)
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
