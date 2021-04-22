package evmchain

import (
	"math/big"
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", Interface.ProgramHash)
	require.NoError(t, err)
}

func TestFaucetBalance(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "ch1")
	err := chain.DeployContract(nil, "evmchain", Interface.ProgramHash)
	require.NoError(t, err)

	ret, err := chain.CallView(Interface.Name, FuncGetBalance, FieldAddress, FaucetAddress.Bytes())
	require.NoError(t, err)

	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(FieldBalance))
	require.Zero(t, FaucetSupply.Cmp(bal))
}
