package tests

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

const incCounterSCName = "inccounter1"

func deployIncCounterSC(t *testing.T, chain *cluster.Chain, counter *cluster.MessageCounter) {
	hname := coretypes.Hn(incCounterSCName)
	description := "testing contract deployment with inccounter" //nolint:goconst
	programHash := inccounter.Interface.ProgramHash
	check(err, t)

	_, err = chain.DeployContract(incCounterSCName, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        incCounterSCName,
	})
	check(err, t)

	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state dict.Dict) bool {
		require.EqualValues(t, 2, blockIndex) //nolint:gomnd
		checkRoots(t, chain)

		contractRegistry := collections.NewMapReadOnly(state, root.VarContractRegistry)
		crBytes := contractRegistry.MustGetAt(hname.Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, cr.Name, incCounterSCName)

		return true
	})

	chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
		require.EqualValues(t, 42, counterValue) //nolint:gomnd
		return true
	})
}
