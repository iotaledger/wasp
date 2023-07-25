package tests

import (
	"context"
	"crypto/ecdsa"
	"runtime/debug"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
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
	programHash := inccounter.Contract.ProgramHash

	tx, err := e.Chain.DeployContract(nativeIncCounterSCName, programHash.String(), map[string]interface{}{
		inccounter.VarCounter: counterStartValue,
		root.ParamName:        nativeIncCounterSCName,
	})
	require.NoError(e.t, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.Chain.ChainID, tx, false, 10*time.Second)
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
			e.t.Logf("deployNativeIncCounterSC: waiting for block=%v on all nodes, nodeIndex=%v, blockIndex=%v, err2=%v", blockIndex, peerIdx, b, err2)
			if retries >= 10 {
				e.t.Fatalf("error on deployIncCounterSC, failed to wait for all peers to be on the same block index after 10 retries. Peer index: %d, stack=%v", peerIdx, string(debug.Stack()))
			}
			// retry (access nodes might take slightly more time to sync)
			retries++
			i--
			time.Sleep(5 * time.Second)
			continue
		}
	}

	e.checkCoreContracts()

	for i := range e.Chain.AllPeers {
		contractRegistry, err2 := e.Chain.ContractRegistry(i)
		require.NoError(e.t, err2)

		cr, ok := lo.Find(contractRegistry, func(item apiclient.ContractInfoResponse) bool {
			return item.HName == nativeIncCounterSCHname.String()
		})
		require.True(e.t, ok)
		require.NotNil(e.t, cr)

		require.EqualValues(e.t, programHash.Hex(), cr.ProgramHash)
		require.EqualValues(e.t, cr.Name, nativeIncCounterSCName)

		counterValue, err2 := e.Chain.GetCounterValue(nativeIncCounterSCHname, i)
		require.NoError(e.t, err2)
		require.EqualValues(e.t, counterStartValue, counterValue)
	}
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, false, 10*time.Second)
	require.NoError(e.t, err)
}

func (e *ChainEnv) expectCounter(hname isc.Hname, counter int64) {
	c := e.getNativeContractCounter(hname)
	require.EqualValues(e.t, counter, c)
}

func (e *ChainEnv) getNativeContractCounter(hname isc.Hname) int64 {
	return e.getCounterForNode(hname, 0)
}

func (e *ChainEnv) getCounterForNode(hname isc.Hname, nodeIndex int) int64 {
	result, _, err := e.Chain.Cluster.WaspClient(nodeIndex).ChainsApi.
		CallView(context.Background(), e.Chain.ChainID.String()).
		ContractCallViewRequest(apiclient.ContractCallViewRequest{
			ContractHName: hname.String(),
			FunctionName:  "getCounter",
		}).Execute()
	require.NoError(e.t, err)

	decodedDict, err := apiextensions.APIJsonDictToDict(*result)
	require.NoError(e.t, err)

	counter, err := codec.DecodeInt64(decodedDict.Get(inccounter.VarCounter), 0)
	require.NoError(e.t, err)

	return counter
}

func (e *ChainEnv) waitUntilCounterEquals(hname isc.Hname, expected int64, duration time.Duration) {
	timeout := time.After(duration)
	var c int64
	allNodesEqualFun := func() bool {
		for _, node := range e.Chain.AllPeers {
			c = e.getCounterForNode(hname, node)
			if c != expected {
				return false
			}
		}
		return true
	}
	for {
		select {
		case <-timeout:
			e.t.Errorf("timeout waiting for inccounter, current: %d, expected: %d", c, expected)
			e.t.Fatal()
		default:
			if allNodesEqualFun() {
				return // success
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func newEthereumAccount() (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return key, crypto.PubkeyToAddress(key.PublicKey)
}
