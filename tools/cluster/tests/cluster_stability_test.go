/**
This test will test the recovery capabilities of the clusterized nodes, where some nodes can break/restart/lay dead at any time.
*/

package tests

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
)

func initializeStabilityTest(t*testing.T, numRequests int, numValidators int, clusterSize int) {
	
	progHash := inccounter.Contract.ProgramHash
	description := "testing with inccounter"

	env := CreateTestEnvironment(t).
					WithCluster(numValidators, clusterSize).
					WithDKG().
					WithDeployChain("chainName").
					WithDeployContract(incCounterSCName, progHash, description, nil).
					Build();
}

func TestOngoingFailureWithoutRecovery(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=6,N=4,req=8", func(t *testing.T) {
		const numRequests = 8
		const quorum = 3
		const numValidators = 20
		const clusterSize = 6

		initializeStabilityTest(t, numRequests, numValidators, clusterSize)
	})
}
