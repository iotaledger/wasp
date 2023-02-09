package tests

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

const nativeIncCounterSCName = "NativeIncCounter"

var nativeIncCounterSCHname = isc.Hn(nativeIncCounterSCName)

// TODO deprecate, or refactor to use the WASM-based inccounter
func (e *ChainEnv) deployNativeIncCounterSC(initCounter ...int) {
	counterStartValue := 42
	if len(initCounter) > 0 {
		counterStartValue = initCounter[0]
	}
	description := "testing contract deployment with inccounter" //nolint:goconst
	programHash := inccounter.Contract.ProgramHash

	tx, err := e.Chain.DeployContract(nativeIncCounterSCName, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: counterStartValue,
		root.ParamName:        nativeIncCounterSCName,
	})
	require.NoError(e.t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.Chain.ChainID, tx, 10*time.Second)
	require.NoError(e.t, err)

	blockIndex, err := e.Chain.BlockIndex()
	require.NoError(e.t, err)
	require.Greater(e.t, blockIndex, uint32(1))

	// wait until all nodes (including access nodes) are at least at block `blockIndex`
	retries := 0
	for i := 1; i < len(e.Chain.AllPeers); i++ {
		peerIdx := e.Chain.AllPeers[i]
		b, err2 := e.Chain.BlockIndex(peerIdx)
		if err2 != nil || b < blockIndex {
			if retries >= 10 {
				e.t.Fatalf("error on deployIncCounterSC, failed to wait for all peers to be on the same block index after 10 retries. Peer index: %d", peerIdx)
			}
			// retry (access nodes might take slightly more time to sync)
			retries++
			i--
			time.Sleep(1 * time.Second)
			continue
		}
	}

	e.checkCoreContracts()

	for i := range e.Chain.AllPeers {
		contractRegistry, err2 := e.Chain.ContractRegistry(i)
		require.NoError(e.t, err2)

		cr := contractRegistry[nativeIncCounterSCHname]
		require.NotNil(e.t, cr)

		require.EqualValues(e.t, programHash, cr.ProgramHash)
		require.EqualValues(e.t, description, cr.Description)
		require.EqualValues(e.t, cr.Name, nativeIncCounterSCName)

		counterValue, err2 := e.Chain.GetCounterValue(nativeIncCounterSCHname, i)
		require.NoError(e.t, err2)
		require.EqualValues(e.t, counterStartValue, counterValue)
	}
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 10*time.Second)
	require.NoError(e.t, err)
}
