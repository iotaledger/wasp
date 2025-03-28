// to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
package clients

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type L2Client interface {
	RequestFunds(ctx context.Context, address cryptolib.Address) error
	Health(ctx context.Context) error
	SignAndExecutePTB(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		pt iotago.ProgrammableTransaction,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	DevInspectPTB(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		pt iotago.ProgrammableTransaction,
		gasPayments []*iotago.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
	) (*iotajsonrpc.DevInspectResults, error)
	StartNewChain(
		ctx context.Context,
		req *iscmoveclient.StartNewChainRequest,
	) (*iscmove.AnchorWithRef, error)
	ReceiveRequestsAndTransition(
		ctx context.Context,
		req *iscmoveclient.ReceiveRequestsAndTransitionRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetAnchorFromObjectID(
		ctx context.Context,
		anchorObjectID *iotago.ObjectID,
	) (*iscmove.AnchorWithRef, error)
	GetPastAnchorFromObjectID(
		ctx context.Context,
		anchorObjectID *iotago.ObjectID,
		version uint64,
	) (*iscmove.AnchorWithRef, error)
	CreateAndSendRequest(
		ctx context.Context,
		req *iscmoveclient.CreateAndSendRequestRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	CreateAndSendRequestWithAssets(
		ctx context.Context,
		req *iscmoveclient.CreateAndSendRequestWithAssetsRequest,
	) (*iotajsonrpc.IotaTransactionBlockResponse, error)
	GetRequestFromObjectID(
		ctx context.Context,
		reqID *iotago.ObjectID,
	) (*iscmove.RefWithObject[iscmove.Request], error)
	GetRequestsSorted(ctx context.Context, packageID iotago.PackageID, anchorAddress *iotago.ObjectID, maxAmountOfRequests int, cb func(error, *iscmove.RefWithObject[iscmove.Request])) error
	GetRequests(
		ctx context.Context,
		packageID iotago.PackageID,
		anchorAddress *iotago.ObjectID,
		maxAmountOfRequests int,
	) (
		[]*iscmove.RefWithObject[iscmove.Request],
		error,
	)
	GetCoin(
		ctx context.Context,
		coinID *iotago.ObjectID,
	) (*iscmoveclient.MoveCoin, error)
	GetAssetsBagWithBalances(
		ctx context.Context,
		assetsBagID *iotago.ObjectID,
	) (*iscmove.AssetsBagWithBalances, error)
	MustWaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) *iotago.ObjectRef
	WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error)
}

var _ L2Client = &iscmoveclient.Client{}
