package test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

const (
	incName = "inccounter"
	incFile = "./inccounter_bg.wasm"

	varCounter    = "counter"
	varNumRepeats = "numRepeats"
	varDelay      = "delay"
)

func TestIncSoloInc(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, incName, incFile)
	require.NoError(t, err)
	req := solo.NewCallParams(incName, "increment").
		AddIotas(1 * iscp.Mi).WithMaxAffordableGasBudget()
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	ret, err := chain.CallView(incName, "getCounter")
	require.NoError(t, err)
	counter, err := codec.DecodeInt64(ret.MustGet(varCounter))
	require.NoError(t, err)
	require.EqualValues(t, 1, counter)
}

func TestIncSoloRepeatMany(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, incName, incFile)
	require.NoError(t, err)
	// fill the target contract, so it is able to issue L1 transactions
	chain.SendFromL1ToL2AccountIotas(
		10*iscp.Mi,
		9*iscp.Mi,
		iscp.NewContractAgentID(chain.ChainID, iscp.Hn(incName)),
		nil,
	)
	req := solo.NewCallParams(incName, "repeatMany", varNumRepeats, 2).
		AddIotas(1 * iscp.Mi).WithMaxAffordableGasBudget()
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)
	chain.WaitUntil(
		func(mempool.MempoolInfo) bool {
			ret, err := chain.CallView(incName, "getCounter")
			require.NoError(t, err)
			counter, err := codec.DecodeInt64(ret.MustGet(varCounter))
			return counter == 3
		},
		10*time.Second,
	)
}
