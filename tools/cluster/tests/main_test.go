package tests

import "testing"

func TestCluster(t *testing.T) {
	tangle := privtangle.New(t)
	// deploy_test.go
	testDeployChain(t)
	testDeployContractOnly(t)
	testDeployContractAndSpawn(t)
}
