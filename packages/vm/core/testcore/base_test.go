package testcore

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNoContractPost(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := solo.NewCall("dummyContract", "dummyEP")
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
}

func TestNoContractView(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	_, err := chain.CallView("dummyContract", "dummyEP")
	require.Error(t, err)
}

func TestNoEPPost(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	req := solo.NewCall(root.Interface.Name, "dummyEP")
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
}

func TestNoEPView(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	_, err := chain.CallView(root.Interface.Name, "dummyEP")
	require.Error(t, err)
}
