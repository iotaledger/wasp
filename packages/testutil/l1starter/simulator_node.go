package l1starter

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/l1"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/move"
)

// SimulatorNode implements IotaNodeEndpoint backed by an in-memory L1 simulator.
type SimulatorNode struct {
	iscPackageOwner iotasigner.Signer
	iscPackageID    iotago.PackageID
	l1Client        *l1.FakeL1Client
}

func NewSimulatorNode(iscPackageOwner iotasigner.Signer) *SimulatorNode {
	handler := move.NewCompositeHandler(iotago.PackageID{}) // placeholder, updated after deploy
	fakeClient := l1.NewFakeL1Client(l1.MoveCallHandler(handler),
		l1.WithPresetBalance(*iscPackageOwner.Address(), 100_000_000_000), // 100 IOTA
	)
	return &SimulatorNode{
		iscPackageOwner: iscPackageOwner,
		l1Client:        fakeClient,
	}
}

func (s *SimulatorNode) Start(ctx context.Context) {
	// Fund the package owner
	err := s.l1Client.RequestFundsFromFaucet(ctx, *s.iscPackageOwner.Address())
	if err != nil {
		panic(fmt.Errorf("simulator faucet failed: %w", err))
	}

	// Deploy ISC contracts (creates a package object in the simulator)
	packageID, err := s.l1Client.L2().DeployISCContracts(ctx, s.iscPackageOwner)
	if err != nil {
		panic(fmt.Errorf("simulator ISC contract deployment failed: %w", err))
	}
	s.iscPackageID = packageID

	// Update the Move handler with the real package ID
	handler := move.NewCompositeHandler(packageID)
	s.l1Client.UpdateMoveHandler(handler)

	fmt.Printf("Simulator: ISC contracts deployed at package ID: %s\n", packageID.String())
}

func (s *SimulatorNode) ISCPackageID() iotago.PackageID {
	return s.iscPackageID
}

func (s *SimulatorNode) APIURL() string {
	return "simulator://in-memory"
}

func (s *SimulatorNode) FaucetURL() string {
	return "simulator://in-memory"
}

func (s *SimulatorNode) L1Client() clients.L1Client {
	return s.l1Client
}

func (s *SimulatorNode) IsLocal() bool {
	return false
}
