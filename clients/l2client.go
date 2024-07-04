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
		signer cryptolib.Signer,
		packageID *sui.PackageID,
		gasPayments []*sui.ObjectRef, // optional
		gasPrice uint64,
		gasBudget uint64,
		execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
		treasuryCap *suijsonrpc.SuiObjectResponse,
	) (*iscmove.Anchor, error)
}

var _ L2Client = &iscmove.Client{}

func NewL2Client(config iscmove.Config) L2Client {
	return iscmove.NewClient(config)
}
