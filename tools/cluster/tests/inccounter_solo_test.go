package tests

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

var incFile = "wasm/inccounter_bg.wasm"

const (
	incName        = "inccounter"
	incDescription = "IncCounter, a PoC smart contract"
)

var incHname = iscp.Hn(incName)

const (
	varCounter    = "counter"
	varNumRepeats = "numRepeats"
	varDelay      = "delay"
)

func TestIncSoloInc(t *testing.T) {
	al := solo.New(t, false, false)
	chain := al.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, incName, incFile)
	require.NoError(t, err)
	req := solo.NewCallParams(incName, "increment").
		WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	ret, err := chain.CallView(incName, "getCounter")
	require.NoError(t, err)
	counter, err := codec.DecodeInt64(ret.MustGet(varCounter))
	require.NoError(t, err)
	require.EqualValues(t, 1, counter)
}

func TestIncSoloRepeatMany(t *testing.T) {
	al := solo.New(t, false, false)
	chain := al.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, incName, incFile)
	require.NoError(t, err)
	req := solo.NewCallParams(incName, "repeatMany", varNumRepeats, 2).
		WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	require.True(t, chain.WaitForRequestsThrough(6))
	// chain.WaitForEmptyBacklog()
	ret, err := chain.CallView(incName, "getCounter")
	require.NoError(t, err)
	counter, err := codec.DecodeInt64(ret.MustGet(varCounter))
	require.NoError(t, err)
	require.EqualValues(t, 3, counter)
}
