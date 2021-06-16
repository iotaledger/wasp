package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes/requestargs"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestDeployChain(t *testing.T) {
	setup(t, "test_cluster")

	counter, err := clu.StartMessageCounter(map[string]int{
		"dismissed_chain": 0,
		"state":           2,
		"request_out":     1,
	})
	check(err, t)
	defer counter.Close()

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	if !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}
	chainID, chainOwnerID := getChainInfo(t, chain)
	require.Equal(t, chainID, chain.ChainID)
	require.Equal(t, chainOwnerID, *coretypes.NewAgentID(chain.OriginatorAddress(), 0))
	t.Logf("--- chainID: %s", chainID.String())
	t.Logf("--- chainOwnerID: %s", chainOwnerID.String())

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state dict.Dict) bool {
		require.EqualValues(t, 1, blockIndex)
		checkRoots(t, chain)
		contractRegistry := collections.NewMapReadOnly(state, root.VarContractRegistry)
		require.EqualValues(t, 4, contractRegistry.MustLen())
		return true
	})
	checkRootsOutside(t, chain)
}

func TestDeployContractOnly(t *testing.T) {
	setup(t, "test_cluster")

	counter, err := clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	check(err, t)
	defer counter.Close()

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	deployIncCounterSC(t, chain, counter)

	// test calling root.FuncFindContractByName view function using client
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, root.Interface.Hname(), root.FuncFindContract,
		dict.Dict{
			root.ParamHname: coretypes.Hn("inncounter1").Bytes(),
		})
	check(err, t)
	recb, err := ret.Get(root.VarData)
	check(err, t)
	rec, err := root.DecodeContractRecord(recb)
	check(err, t)
	require.EqualValues(t, "testing contract deployment with inccounter", rec.Description)
}

func TestDeployContractAndSpawn(t *testing.T) {
	setup(t, "test_cluster")

	counter, err := clu.StartMessageCounter(map[string]int{
		"dismissed_committee": 0,
		"state":               2,
		"request_out":         1,
	})
	check(err, t)
	defer counter.Close()

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	deployIncCounterSC(t, chain, counter)

	hname := coretypes.Hn("inncounter1")

	nameNew := "spawnedContract"
	dscrNew := "spawned contract it is"
	hnameNew := coretypes.Hn(nameNew)
	// send 'spawn' request to the SC which was just deployed
	tx, err := chain.OriginatorClient().Post1Request(hname, coretypes.Hn(inccounter.FuncSpawn), chainclient.PostRequestParams{
		Args: requestargs.New().AddEncodeSimpleMany(codec.MakeDict(map[string]interface{}{
			inccounter.VarName:        nameNew,
			inccounter.VarDescription: dscrNew,
		})),
	})
	check(err, t)

	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
	check(err, t)

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state dict.Dict) bool {
		require.EqualValues(t, 3, blockIndex)
		checkRoots(t, chain)

		contractRegistry := collections.NewMapReadOnly(state, root.VarContractRegistry)
		require.EqualValues(t, 6, contractRegistry.MustLen())
		//--
		crBytes := contractRegistry.MustGetAt(accounts.Interface.Hname().Bytes())
		require.NotNil(t, crBytes)
		cr, err := root.DecodeContractRecord(crBytes)
		check(err, t)
		require.EqualValues(t, accounts.Interface.ProgramHash, cr.ProgramHash)
		require.EqualValues(t, accounts.Interface.Description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, accounts.Interface.Name, cr.Name)

		//--
		crBytes = contractRegistry.MustGetAt(hnameNew.Bytes())
		require.NotNil(t, crBytes)
		cr, err = root.DecodeContractRecord(crBytes)
		check(err, t)
		// TODO check program hash
		require.EqualValues(t, dscrNew, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, nameNew, cr.Name)
		return true
	})
	chain.WithSCState(hname, func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
		require.EqualValues(t, 42, counterValue)
		return true
	})
	chain.WithSCState(coretypes.Hn(nameNew), func(host string, blockIndex uint32, state dict.Dict) bool {
		counterValue, _, _ := codec.DecodeInt64(state.MustGet(inccounter.VarCounter))
		require.EqualValues(t, 44, counterValue)
		return true
	})

}
