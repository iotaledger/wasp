package tests

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

const incCounterSCName = "inccounter1"

var incCounterSCHname = iscp.Hn(incCounterSCName)

func deployIncCounterSC(t *testing.T, chain *cluster.Chain, counter *cluster.MessageCounter) *ledgerstate.Transaction {
	description := "testing contract deployment with inccounter" //nolint:goconst
	programHash := inccounter.Interface.ProgramHash
	check(err, t)

	tx, err := chain.DeployContract(incCounterSCName, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        incCounterSCName,
	})
	check(err, t)

	if counter != nil && !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	checkCoreContracts(t, chain)

	for i := range chain.CommitteeNodes {
		blockIndex, err := chain.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 2, blockIndex)

		contractRegistry, err := chain.ContractRegistry(i)
		require.NoError(t, err)

		cr := contractRegistry[incCounterSCHname]

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, cr.Name, incCounterSCName)

		counterValue, err := chain.GetCounterValue(incCounterSCHname, i)
		require.NoError(t, err)
		require.EqualValues(t, 42, counterValue)
	}

	return tx
}
