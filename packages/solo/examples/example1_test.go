package examples

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExample1(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "exampleChain")

	chainID, ownerID, coreContracts := chain.GetInfo() // calls view root::GetInfo

	require.EqualValues(t, 4, len(coreContracts)) // 4 core contracts deployed by default

	t.Logf("chainID: %s", chainID)
	t.Logf("chain owner ID: %s", ownerID)
	for _, rec := range coreContracts {
		t.Logf("    Contract: %s", rec.Name)
	}
}
