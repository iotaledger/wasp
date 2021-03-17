package examples

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestExample1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ex1")

	chainID, chainOwner, coreContracts := chain.GetInfo() // calls view root::GetInfo
	require.EqualValues(t, 4, len(coreContracts))         // 4 core contracts deployed by default

	t.Logf("chainID: %s", chainID.String())
	t.Logf("chain owner ID: %s", chainOwner.String())
	for hname, rec := range coreContracts {
		cid := coretypes.NewContractID(chain.ChainID, hname)
		t.Logf("    Core contract '%s': %s", rec.Name, cid.String())
	}
}

func TestExample2(t *testing.T) {
	env := solo.New(t, false, false)
	_, userAddress := env.NewKeyPair()
	t.Logf("Address of the userWallet is: %s", userAddress.Base58())
	numIotas := env.GetAddressBalance(userAddress, ledgerstate.ColorIOTA)
	t.Logf("balance of the userWallet is: %d iota", numIotas)
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, 0)
}

func TestExample3(t *testing.T) {
	env := solo.New(t, false, false)
	_, userAddress := env.NewKeyPairWithFunds()
	t.Logf("Address of the userWallet is: %s", userAddress.Base58())
	numIotas := env.GetAddressBalance(userAddress, ledgerstate.ColorIOTA)
	t.Logf("balance of the userWallet is: %d iota", numIotas)
	env.AssertAddressBalance(userAddress, ledgerstate.ColorIOTA, solo.Saldo)
}
