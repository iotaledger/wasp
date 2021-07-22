/**
This test will test the recovery capabilities of the clusterized nodes, where some nodes can break/restart/lay dead at any time.
*/

package tests

import (
	"testing"
)

func initializeStabilityTest(numRequests int, quorum int, clusterSize int) {
}

func TestOngoingFailureWithoutRecovery(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=6,N=4,req=8", func(t *testing.T) {
		const numRequests = 8
		const quorum = 3
		const clusterSize = 6
	})
}
