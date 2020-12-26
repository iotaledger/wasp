package examples

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExample1(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "exampleChain")

	chainInfo, coreContracts := chain.GetInfo() // calls view root::GetInfo

	require.EqualValues(t, 4, len(coreContracts)) // 4 core contracts deployed by default

	t.Logf("chainID: %s", chainInfo.ChainID)
	t.Logf("chain owner ID: %s", chainInfo.ChainOwnerID)
	for hname, rec := range coreContracts {
		t.Logf("    Core contract #%d: %s", hname, rec.Name)
	}
}

func TestExample2(t *testing.T) {
	glb := solo.New(t, false, false)
	userWallet := glb.NewSignatureSchemeWithFunds()
	userAddress := userWallet.Address()
	t.Logf("Address of the userWallet is: %s", userAddress)
	numIotas := glb.GetUtxodbBalance(userAddress, balance.ColorIOTA)
	t.Logf("balance of the userWallet is: %d i", numIotas)
	glb.AssertUtxodbBalance(userAddress, balance.ColorIOTA, 1337)
}
