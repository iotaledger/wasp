// to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
package clients

import (
	"context"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type L2Client interface {
	StartNewChain(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID *sui.PackageID,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) ([]byte, error)
	CreateAndSendRequest(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID *sui.PackageID,
		anchorAddress *sui.ObjectID,
		assetsBagRef *sui.ObjectRef,
		iscContractName string,
		iscFunctionName string,
		args [][]byte,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) ([]byte, error)
	ReceiveAndUpdateStateRootRequest(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID *sui.PackageID,
		anchor *sui.ObjectRef,
		reqObjects []*sui.ObjectRef,
		stateRoot []byte,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) ([]byte, error)
	AssetsBagNew(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID *sui.PackageID,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) ([]byte, error)
	AssetsBagPlaceCoin(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID *sui.PackageID,
		assetsBagRef *sui.ObjectRef,
		coin *sui.ObjectRef,
		coinType string,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) ([]byte, error)
	AssetsDestroyEmpty(
		ctx context.Context,
		cryptolibSigner cryptolib.Signer,
		packageID *sui.PackageID,
		assetsBagRef *sui.ObjectRef,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		devMode bool,
	) ([]byte, error)
	GetAnchorFromSuiTransactionBlockResponse(
		ctx context.Context,
		response *suijsonrpc.SuiTransactionBlockResponse,
	) (*iscmove.Anchor, error)
}

var _ L2Client = &iscmove.Client{}

func NewL2Client(config iscmove.Config) L2Client {
	return iscmove.NewClient(config)
}
