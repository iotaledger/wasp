package tests

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

const incCounterSCName = "inccounter1"

var incCounterSCHname = iscp.Hn(incCounterSCName)

func (e *chainEnv) deployIncCounterSC(counter *cluster.MessageCounter) *iotago.Transaction {
	description := "testing contract deployment with inccounter" //nolint:goconst
	programHash := inccounter.Contract.ProgramHash

	tx, err := e.chain.DeployContract(incCounterSCName, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        incCounterSCName,
	})
	require.NoError(e.t, err)

	if counter != nil && !counter.WaitUntilExpectationsMet() {
		e.t.Fatal()
	}

	e.checkCoreContracts()

	for i := range e.chain.AllPeers {
		blockIndex, err := e.chain.BlockIndex(i)
		require.NoError(e.t, err)
		require.Greater(e.t, blockIndex, uint32(1))

		contractRegistry, err := e.chain.ContractRegistry(i)
		require.NoError(e.t, err)

		cr := contractRegistry[incCounterSCHname]

		require.EqualValues(e.t, programHash, cr.ProgramHash)
		require.EqualValues(e.t, description, cr.Description)
		require.EqualValues(e.t, cr.Name, incCounterSCName)

		counterValue, err := e.chain.GetCounterValue(incCounterSCHname, i)
		require.NoError(e.t, err)
		require.EqualValues(e.t, 42, counterValue)
	}

	return tx
}
