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

type StartNewChainRequest struct {
	Signer        cryptolib.Signer
	AnchorOwner   *cryptolib.Address
	PackageID     iotago.PackageID
	StateMetadata []byte
	InitCoinRef   *iotago.ObjectRef
	GasPayments   []*iotago.ObjectRef
	GasPrice      uint64
	GasBudget     uint64
}

// StartNewChain is the only exception which doesn't use committee's GasCoin (the one in StateMetadata) for paying gas fee
// this func automatically pick a coin
func (c *Client) StartNewChain(
	ctx context.Context,
	req *StartNewChainRequest,
) (*iscmove.AnchorWithRef, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	var argInitCoin iotago.Argument
	if req.InitCoinRef != nil {
		ptb = PTBOptionSomeIotaCoin(ptb, req.InitCoinRef)
	} else {
		ptb = PTBOptionNoneIotaCoin(ptb)
	}
	argInitCoin = ptb.LastCommandResultArg()

	ptb = PTBStartNewChain(ptb, req.PackageID, req.StateMetadata, argInitCoin, req.AnchorOwner)
	txnResponse, err := c.SignAndExecutePTB(
		ctx,
		req.Signer,
		ptb.Finish(),
		req.GasPayments,
		req.GasPrice,
		req.GasBudget,
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

type ReceiveRequestsAndTransitionRequest struct {
	Signer           cryptolib.Signer
	PackageID        iotago.PackageID
	AnchorRef        *iotago.ObjectRef
	ConsumedRequests []iotago.ObjectRef
	SentAssets       []SentAssets
	StateMetadata    []byte
	TopUpAmount      uint64
	GasPayment       *iotago.ObjectRef
	GasPrice         uint64
	GasBudget        uint64
}

func (c *Client) ReceiveRequestsAndTransition(
	ctx context.Context,
	req *ReceiveRequestsAndTransitionRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	consumed := make([]ConsumedRequest, 0, len(req.ConsumedRequests))
	for _, reqRef := range req.ConsumedRequests {
		reqWithObj, err := c.GetRequestFromObjectID(ctx, reqRef.ObjectID)
		if err != nil {
			return nil, err
		}
		assetsBag, err := c.GetAssetsBagWithBalances(ctx, &reqWithObj.Object.AssetsBag.ID)
		if err != nil {
			return nil, err
		}
		consumed = append(consumed, ConsumedRequest{
			RequestRef: reqRef,
			Assets:     assetsBag,
		})
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb = PTBReceiveRequestsAndTransition(
		ptb,
		req.PackageID,
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: req.AnchorRef}),
		consumed,
		req.SentAssets,
		req.StateMetadata,
		req.TopUpAmount,
	)
	return c.SignAndExecutePTB(
		ctx,
		req.Signer,
		ptb.Finish(),
		[]*iotago.ObjectRef{req.GasPayment},
		req.GasPrice,
		req.GasBudget,
	)
}

func (c *Client) GetAnchorFromObjectID(
	ctx context.Context,
	anchorObjectID *iotago.ObjectID,
) (*iscmove.AnchorWithRef, error) {
	getObjectResponse, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: anchorObjectID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true, ShowOwner: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}
	if getObjectResponse.Error != nil {
		return nil, fmt.Errorf("failed to get anchor content: %s", getObjectResponse.Error.Data.String())
	}
	return decodeAnchorBCS(
		getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes,
		getObjectResponse.Data.Ref(),
		getObjectResponse.Data.Owner.AddressOwner,
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
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true, ShowOwner: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor content: %w", err)
	}
	if getObjectResponse.Data.ObjectDeleted != nil {
		return nil, fmt.Errorf("failed to get anchor content: deleted")
	}
	if getObjectResponse.Data.ObjectNotExists != nil {
		return nil, fmt.Errorf("failed to get anchor content: object does not exist")
	}
	if getObjectResponse.Data.VersionNotFound != nil {
		return nil, fmt.Errorf("failed to get anchor content: version not found")
	}
	if getObjectResponse.Data.VersionTooHigh != nil {
		return nil, fmt.Errorf("failed to get anchor content: version too high")
	}
	return decodeAnchorBCS(
		getObjectResponse.Data.VersionFound.Bcs.Data.MoveObject.BcsBytes,
		getObjectResponse.Data.VersionFound.Ref(),
		getObjectResponse.Data.VersionFound.Owner.AddressOwner,
	)
}

func decodeAnchorBCS(bcsBytes iotago.Base64Data, ref iotago.ObjectRef, owner *iotago.Address) (*iscmove.AnchorWithRef, error) {
	var moveAnchor iscmove.Anchor
	err := iotaclient.UnmarshalBCS(bcsBytes, &moveAnchor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}
	return &iscmove.AnchorWithRef{
		ObjectRef: ref,
		Object:    &moveAnchor,
		Owner:     owner,
	}, nil
}
