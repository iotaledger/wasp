package tests

import (
	"testing"
)

func TestOffledgerRequests(t *testing.T) {

	setup(t, "test_cluster")

	counter, err := clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	check(err, t)
	defer counter.Close()

	nodes := clu.Config.AllNodes()
	committeeNodes := nodes[0:3]
	accessNodes := nodes[3:len(nodes)]
	minQuorum := len(committeeNodes)/2 + 1
	quorum := len(committeeNodes) * 3 / 4
	// deploy custom chain with 3 committee nodes and 1 access node
	chain, err := clu.DeployChain("3 committee nodes", committeeNodes, quorum)
	check(err, t)

	deployIncCounterSC(t, chain, counter)

	//TODO send offledger request
}
