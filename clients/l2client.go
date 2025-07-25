// Package clients provides layer 2 client functionality to be used by utilities
// like: cluster-tool, wasp-cli, apilib, etc.
package clients

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
)

type L2Client interface {
	StartNewChain(
		ctx context.Context,
		req *iscmoveclient.StartNewChainRequest,
	) (*iscmove.AnchorWithRef, error)
	UpdateAnchorStateMetadata(ctx context.Context, req *iscmoveclient.UpdateAnchorStateMetadataRequest) (bool, error)
	CreateAndSendRequest(
		ctx context.Context,
		req *iscmoveclient.CreateAndSendRequestRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	ReceiveRequestsAndTransition(
		ctx context.Context,
		req *iscmoveclient.ReceiveRequestsAndTransitionRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetAssetsBagWithBalances(
		ctx context.Context,
		assetsBagID *iotago.ObjectID,
	) (*iscmove.AssetsBagWithBalances, error)
	CreateAndSendRequestWithAssets(
		ctx context.Context,
		req *iscmoveclient.CreateAndSendRequestWithAssetsRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetAnchorFromObjectID(
		ctx context.Context,
		anchorObjectID *iotago.ObjectID,
	) (*iscmove.RefWithObject[iscmove.Anchor], error)
	GetRequestFromObjectID(
		ctx context.Context,
		reqID *iotago.ObjectID,
	) (*iscmove.RefWithObject[iscmove.Request], error)
	GetCoin(
		ctx context.Context,
		coinID *iotago.ObjectID,
	) (*iscmoveclient.MoveCoin, error)
}

var _ L2Client = &iscmoveclient.Client{}
