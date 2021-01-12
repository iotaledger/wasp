package examples

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExample1(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "ex1")

	chainInfo, coreContracts := chain.GetInfo()   // calls view root::GetInfo
	require.EqualValues(t, 4, len(coreContracts)) // 4 core contracts deployed by default

	t.Logf("chainID: %s", chainInfo.ChainID)
	t.Logf("chain owner ID: %s", chainInfo.ChainOwnerID)
	for hname, rec := range coreContracts {
		t.Logf("    Core contract '%s': %s", rec.Name, coretypes.NewContractID(chain.ChainID, hname))
	}
}

func TestExample2(t *testing.T) {
	glb := solo.New(t, false, false)
	userWallet := glb.NewSignatureSchemeWithFunds()
	userAddress := userWallet.Address()
	t.Logf("Address of the userWallet is: %s", userAddress)
	numIotas := glb.GetAddressBalance(userAddress, balance.ColorIOTA)
	t.Logf("balance of the userWallet is: %d iota", numIotas)
	glb.AssertAddressBalance(userAddress, balance.ColorIOTA, 1337)
}
