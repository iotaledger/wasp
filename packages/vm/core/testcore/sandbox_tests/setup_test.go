package sandbox_tests

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	RUN_WASM  = false
	WASM_FILE = "testcore_bg.wasm"
)

var SCName = "test_sandbox"

func setupChain(t *testing.T, sigSchemeChain signaturescheme.SignatureScheme) (*solo.Solo, *solo.Chain) {
	env := solo.New(t, false, false)
	chain := env.NewChain(sigSchemeChain, "ch1")
	return env, chain
}

func setupDeployer(t *testing.T, chain *solo.Chain) signaturescheme.SignatureScheme {
	user := chain.Env.NewSignatureSchemeWithFunds()
	chain.Env.AssertAddressBalance(user.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)

	req := solo.NewCall(root.Interface.Name, root.FuncGrantDeploy,
		root.ParamDeployer, coretypes.NewAgentIDFromAddress(user.Address()),
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)
	return user
}

func setupSC(t *testing.T, chain *solo.Chain, user signaturescheme.SignatureScheme) coretypes.ContractID {
	var err error
	if RUN_WASM {
		err = chain.DeployWasmContract(user, SCName, WASM_FILE)
	} else {
		err = chain.DeployContract(user, SCName, test_sandbox.Interface.ProgramHash)
	}
	require.NoError(t, err)

	deployed := coretypes.NewContractID(chain.ChainID, coretypes.Hn(test_sandbox.Interface.Name))
	req := solo.NewCall(SCName, test_sandbox.FuncDoNothing)
	_, err = chain.PostRequest(req, user)
	require.NoError(t, err)
	t.Logf("deployed '%s': %s", SCName, coretypes.Hn(SCName))
	return deployed
}

func TestSetup1(t *testing.T) {
	_, chain := setupChain(t, nil)
	setupSC(t, chain, nil)
}

func TestSetup2(t *testing.T) {
	_, chain := setupChain(t, nil)
	user := setupDeployer(t, chain)
	setupSC(t, chain, user)
}
