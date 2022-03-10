package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil"
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
		AddIotas(1)
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
		AddIotas(1)
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

func TestSpamCallViewWasm(t *testing.T) {
	testutil.RunHeavy(t)
	clu := newCluster(t)
	committee := []int{0}
	quorum := uint16(1)
	addr, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)
	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), committee, quorum, addr)
	require.NoError(t, err)

	e := &env{t: t, clu: clu}
	chEnv := &chainEnv{
		env:   e,
		chain: chain,
	}

	chEnv.requestFunds(scOwnerAddr, "client")
	chEnv.deployContract(incName, incDescription, nil)

	{
		// increment counter once
		tx, err := chEnv.chainClient().Post1Request(incHname, iscp.Hn("increment"))
		require.NoError(t, err)
		err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
	}

	const n = 200
	ch := make(chan error, n)

	for i := 0; i < n; i++ {
		go func() {
			cl := chain.SCClient(incHname, nil)
			r, err := cl.CallView("getCounter", nil)
			if err != nil {
				ch <- err
				return
			}

			v, err := codec.DecodeInt64(r.MustGet(inccounter.VarCounter))
			if err == nil && v != 1 {
				err = errors.New("v != 1")
			}
			ch <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-ch
		if err != nil {
			t.Error(err)
		}
	}
}
