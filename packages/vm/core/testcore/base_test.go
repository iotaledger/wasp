package testcore

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNoContractPost(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams("dummyContract", "dummyEP")
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
}

func TestNoContractView(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	_, err := chain.CallView("dummyContract", "dummyEP")
	require.Error(t, err)
}

func TestNoEPPost(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Interface.Name, "dummyEP")
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
}

func TestNoEPView(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	_, err := chain.CallView(root.Interface.Name, "dummyEP")
	require.Error(t, err)
}
