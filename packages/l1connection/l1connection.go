// to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
package l1connection

import (
	"context"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type Config struct {
	APIAddress    string
	INXAddress    string
	FaucetAddress string
	FaucetKey     *cryptolib.KeyPair
	UseRemotePoW  bool
}

type Client interface {
	RequestFunds(ctx context.Context, address cryptolib.Address) error
	Health(ctx context.Context) error

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

var _ Client = &iscmove.Client{}
