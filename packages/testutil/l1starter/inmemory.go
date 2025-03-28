package l1starter

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/cryptolib"
	`github.com/iotaledger/wasp/packages/testutil/l1starter/inmemory`
	"github.com/lmoe/iota-simulator"
)

type InMemory struct {
	iscPackageOwner iotasigner.Signer
	iscPackageID    iotago.PackageID

	sim      *iotasimulator.Simulator
	l1Client clients.L1Client
}

func NewInMemory(iscPackageOwner iotasigner.Signer) *InMemory {
	return &InMemory{
		iscPackageOwner: iscPackageOwner,
	}
}

func (m *InMemory) ISCPackageID() iotago.PackageID {
	return m.iscPackageID
}

func (m *InMemory) APIURL() string {
	return ""
}

func (m *InMemory) FaucetURL() string {
	return ""
}

func (m *InMemory) L1Client() clients.L1Client {
	return m.l1Client
}

func (m *InMemory) L2Client() clients.L2Client {
	return m.l1Client.L2()
}

func (m *InMemory) IsLocal() bool {
	return false
}

func (m *InMemory) start(ctx context.Context) {
	sim, err := iotasimulator.NewSimulator()
	if err != nil {
		panic(err)
	}

	m.sim = sim
	m.l1Client = inmemory.NewInMemoryClient(m.sim)

	client := m.L1Client()

	err = client.RequestFunds(ctx, *cryptolib.NewAddressFromIota(m.iscPackageOwner.Address()))
	if err != nil {
		panic(fmt.Errorf("faucet request failed: %w", err))
	}

	m.iscPackageID, err = client.DeployISCContracts(ctx, m.iscPackageOwner)
	if err != nil {
		panic(fmt.Errorf("isc contract deployment failed: %w", err))
	}
}

func (m *InMemory) stop() error {
	return nil
}
