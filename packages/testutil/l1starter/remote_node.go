package l1starter

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/clients/l1simulator"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/samber/lo"
)

type RemoteIotaNode struct {
	faucetURL       string
	apiURL          string
	iscPackageOwner iotasigner.Signer
	iscPackageID    iotago.PackageID
}

func NewRemoteIotaNode(apiURL string, faucetURL string, iscPackageOwner iotasigner.Signer) *RemoteIotaNode {
	return &RemoteIotaNode{
		faucetURL:       faucetURL,
		apiURL:          apiURL,
		iscPackageOwner: iscPackageOwner,
	}
}

func (r *RemoteIotaNode) ISCPackageID() iotago.PackageID {
	return r.iscPackageID
}

func (r *RemoteIotaNode) APIURL() string {
	return r.apiURL
}

func (r *RemoteIotaNode) FaucetURL() string {
	return r.faucetURL
}

func (r *RemoteIotaNode) L1Client() clients.L1Client {
	return l1simulator.NewClient(clients.L1Config{
		APIURL:    r.APIURL(),
		FaucetURL: r.FaucetURL(),
	}, "http://localhost:30003")
}

func (in *RemoteIotaNode) IsLocal() bool {
	return false
}

func (r *RemoteIotaNode) start(ctx context.Context) {
	client := r.L1Client()

	err := client.RequestFunds(ctx, *cryptolib.NewAddressFromIota(r.iscPackageOwner.Address()))
	if err != nil {
		panic(fmt.Errorf("faucet request failed: %w for url: %s", err, r.faucetURL))
	}

	fmt.Println(lo.Must(client.GetAllBalances(ctx, r.iscPackageOwner.Address())))
	fmt.Println(lo.Must(client.GetAllBalances(ctx, r.iscPackageOwner.Address())))

	r.iscPackageID, err = client.DeployISCContracts(ctx, r.iscPackageOwner)
	if err != nil {
		panic(fmt.Errorf("isc contract deployment failed: %w", err))
	}
}
