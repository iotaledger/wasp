package tests

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/tools/cluster"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"

	"github.com/iotaledger/wasp/contracts/native/inccounter"

	"github.com/stretchr/testify/require"

	clutest "github.com/iotaledger/wasp/tools/cluster/testutil"
)

var (
	contractName  = "inccounter"
	contractHname = coretypes.Hn(contractName)
)

func TestAccessNode(t *testing.T) {
	//core.PrintWellKnownHnames()
	//t.Logf("contract: name = %s, hname = %s", contractName, contractHname.String())
	clu := clutest.NewCluster(t, 10)

	numRequests := 8
	cmt1 := []int{0, 1, 2, 3}

	addr1, err := clu.RunDKG(cmt1, 3)
	require.NoError(t, err)

	t.Logf("addr1: %s", addr1.Base58())

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), cmt1, 3, addr1)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID.Base58())

	description := "testing with inccounter"
	programHash = inccounter.Interface.ProgramHash

	_, err = chain.DeployContract(contractName, programHash.String(), description, nil)
	require.NoError(t, err)

	rec, err := findContract(chain, contractName)
	require.NoError(t, err)
	require.EqualValues(t, contractName, rec.Name)

	kp := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(kp.PublicKey)
	err = requestFunds(clu, myAddress, "myAddress")
	require.NoError(t, err)

	myClient := chain.SCClient(contractHname, kp)

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostRequest(inccounter.FuncIncCounter)
		require.NoError(t, err)
	}

	//err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
	//require.NoError(t, err)

	require.True(t, waitCounter(t, chain, 7, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 8, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 9, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 4, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 5, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 6, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 1, numRequests, 5*time.Second))
}

func TestRotation(t *testing.T) {
	cmt1 := []int{0, 1, 2, 3}
	cmt2 := []int{2, 3, 4, 5}

	clu := clutest.NewCluster(t, 6)
	addr1, err := clu.RunDKG(cmt1, 3)
	require.NoError(t, err)
	addr2, err := clu.RunDKG(cmt2, 3)
	require.NoError(t, err)

	t.Logf("addr1: %s", addr1.Base58())
	t.Logf("addr2: %s", addr2.Base58())

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), cmt1, 3, addr1)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID.Base58())

	description := "testing contract deployment with inccounter"
	programHash = inccounter.Interface.ProgramHash

	_, err = chain.DeployContract(contractName, programHash.String(), description, nil)
	require.NoError(t, err)

	rec, err := findContract(chain, contractName)
	require.NoError(t, err)
	require.EqualValues(t, contractName, rec.Name)

	kp := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(kp.PublicKey)
	err = requestFunds(clu, myAddress, "myAddress")
	require.NoError(t, err)

	myClient := chain.SCClient(contractHname, kp)

	_, err = myClient.PostRequest(inccounter.FuncIncCounter)
	require.NoError(t, err)
	_, err = myClient.PostRequest(inccounter.FuncIncCounter)
	require.NoError(t, err)
	_, err = myClient.PostRequest(inccounter.FuncIncCounter)
	require.NoError(t, err)

	// err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain.ChainID, tx, 30*time.Second)
	// require.NoError(t, err)

	require.True(t, waitCounter(t, chain, 0, 3, 5*time.Second))
	// require.True(t, waitCounter(t, chain, 5, 3, 5*time.Second))

	// TODO not finished with node config
}

func waitCounter(t *testing.T, chain *cluster.Chain, nodeIndex, counter int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		c, err := callGetCounter(t, chain, nodeIndex)
		if err == nil && c >= int64(counter) {
			return true
		}
		time.Sleep(30 * time.Millisecond)
		if time.Now().After(deadline) {
			return false
		}
	}
}

func callGetCounter(t *testing.T, chain *cluster.Chain, nodeIndex int) (int64, error) {
	ret, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID, contractHname, "getCounter",
	)
	if err != nil {
		return 0, err
	}

	counter, _, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
	require.NoError(t, err)

	return counter, nil
}
