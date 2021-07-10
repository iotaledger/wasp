package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/cluster"
	clutest "github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

// cluster of 10 access nodes with the committee of 4 nodes. tested if all nodes are synced
func TestAccessNodeOne(t *testing.T) {
	// core.PrintWellKnownHnames()
	// t.Logf("contract: name = %s, hname = %s", contractName, contractHname.String())
	clu1 := clutest.NewCluster(t, 10)

	numRequests := 8
	cmt1 := []int{0, 1, 2, 3}

	addr1, err := clu1.RunDKG(cmt1, 3)
	require.NoError(t, err)

	t.Logf("addr1: %s", addr1.Base58())

	chain1, err := clu1.DeployChain("chain", clu1.Config.AllNodes(), cmt1, 3, addr1)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain1.ChainID.Base58())

	description := "testing with inccounter"
	programHash = inccounter.Interface.ProgramHash

	_, err = chain1.DeployContract(incCounterSCName, programHash.String(), description, nil)
	require.NoError(t, err)

	waitUntil(t, contractIsDeployed(chain1, incCounterSCName), clu1.Config.AllNodes(), 30*time.Second)

	kp := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(kp.PublicKey)
	err = requestFunds(clu1, myAddress, "myAddress")
	require.NoError(t, err)

	myClient := chain1.SCClient(incCounterSCHname, kp)

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostRequest(inccounter.FuncIncCounter)
		require.NoError(t, err)
	}

	waitUntil(t, counterEquals(chain1, int64(numRequests)), []int{7, 8, 9, 4, 5, 6, 1}, 5*time.Second)
}

// cluster of 10 access nodes with the committee of 4 nodes. tested if all nodes are synced
func TestAccessNodeMany(t *testing.T) {
	const clusterSize = 15
	committee := []int{0, 1, 2, 3, 4, 5, 6}
	const quorum = 5
	const requestsCountInitial = 8
	const requestsCountIncrement = 8
	const iterationCount = 4

	clu1 := clutest.NewCluster(t, clusterSize)

	addr1, err := clu1.RunDKG(committee, uint16(quorum))
	require.NoError(t, err)

	t.Logf("addr1: %s", addr1.Base58())

	chain1, err := clu1.DeployChain("chain", clu1.Config.AllNodes(), committee, uint16(quorum), addr1)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain1.ChainID.Base58())

	description := "testing with inccounter"
	programHash = inccounter.Interface.ProgramHash

	_, err = chain1.DeployContract(incCounterSCName, programHash.String(), description, nil)
	require.NoError(t, err)

	waitUntil(t, contractIsDeployed(chain1, incCounterSCName), clu1.Config.AllNodes(), 30*time.Second, "contract to be deployed")

	kp := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(kp.PublicKey)
	err = requestFunds(clu1, myAddress, "myAddress")
	require.NoError(t, err)

	myClient := chain1.SCClient(incCounterSCHname, kp)

	requestsCount := requestsCountInitial
	requestsCummulative := 0
	for i := 0; i < iterationCount; i++ {
		logMsg := fmt.Sprintf("iteration %v of %v requests", i, requestsCount)
		t.Logf("Running %s", logMsg)
		for j := 0; j < requestsCount; j++ {
			_, err = myClient.PostRequest(inccounter.FuncIncCounter)
			require.NoError(t, err)
		}
		requestsCummulative += requestsCount
		waitUntil(t, counterEquals(chain1, int64(requestsCummulative)), clu1.Config.AllNodes(), 60*time.Second, logMsg)
		requestsCount *= requestsCountIncrement
	}
}

// cluster of 10 access nodes and two overlapping committees
func TestRotation(t *testing.T) {
	numRequests := 8

	cmt1 := []int{0, 1, 2, 3}
	cmt2 := []int{2, 3, 4, 5}

	clu1 := clutest.NewCluster(t, 10)
	addr1, err := clu1.RunDKG(cmt1, 3)
	require.NoError(t, err)
	addr2, err := clu1.RunDKG(cmt2, 3)
	require.NoError(t, err)

	t.Logf("addr1: %s", addr1.Base58())
	t.Logf("addr2: %s", addr2.Base58())

	chain1, err := clu1.DeployChain("chain", clu1.Config.AllNodes(), cmt1, 3, addr1)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain1.ChainID.Base58())

	description := "inccounter testing contract"
	programHash = inccounter.Interface.ProgramHash

	_, err = chain1.DeployContract(incCounterSCName, programHash.String(), description, nil)
	require.NoError(t, err)

	waitUntil(t, contractIsDeployed(chain1, incCounterSCName), clu1.Config.AllNodes(), 30*time.Second)

	require.True(t, waitStateController(t, chain1, 0, addr1, 5*time.Second))
	require.True(t, waitStateController(t, chain1, 9, addr1, 5*time.Second))

	kp := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(kp.PublicKey)
	err = requestFunds(clu1, myAddress, "myAddress")
	require.NoError(t, err)

	myClient := chain1.SCClient(incCounterSCHname, kp)

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostRequest(inccounter.FuncIncCounter)
		require.NoError(t, err)
	}

	waitUntil(t, counterEquals(chain1, int64(numRequests)), []int{0, 3, 8, 9}, 5*time.Second)

	govClient := chain1.SCClient(governance.Interface.Hname(), chain1.OriginatorKeyPair())

	params := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addr2).WithIotas(1)
	tx, err := govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress, *params)

	require.NoError(t, err)

	require.True(t, waitBlockIndex(t, chain1, 9, 4, 5*time.Second))
	require.True(t, waitBlockIndex(t, chain1, 0, 4, 5*time.Second))
	require.True(t, waitBlockIndex(t, chain1, 6, 4, 5*time.Second))

	reqid := coretypes.NewRequestID(tx.ID(), 0)

	require.EqualValues(t, "", waitRequest(t, chain1, 0, reqid, 5*time.Second))
	require.EqualValues(t, "", waitRequest(t, chain1, 9, reqid, 5*time.Second))

	require.NoError(t, err)
	require.True(t, isAllowedStateControllerAddress(t, chain1, 0, addr2))

	require.True(t, waitStateController(t, chain1, 0, addr1, 5*time.Second))
	require.True(t, waitStateController(t, chain1, 9, addr1, 5*time.Second))

	params = chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addr2).WithIotas(1)
	tx, err = govClient.PostRequest(governance.FuncRotateStateController, *params)
	require.NoError(t, err)

	require.True(t, waitStateController(t, chain1, 0, addr2, 5*time.Second))
	require.True(t, waitStateController(t, chain1, 9, addr2, 5*time.Second))

	require.True(t, waitBlockIndex(t, chain1, 9, 5, 5*time.Second))
	require.True(t, waitBlockIndex(t, chain1, 0, 5, 5*time.Second))
	require.True(t, waitBlockIndex(t, chain1, 6, 5, 5*time.Second))

	reqid = coretypes.NewRequestID(tx.ID(), 0)
	require.EqualValues(t, "", waitRequest(t, chain1, 0, reqid, 5*time.Second))
	require.EqualValues(t, "", waitRequest(t, chain1, 9, reqid, 5*time.Second))

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostRequest(inccounter.FuncIncCounter)
		require.NoError(t, err)
	}

	waitUntil(t, counterEquals(chain1, int64(2*numRequests)), clu1.Config.AllNodes(), 5*time.Second)
}

func TestRotationMany(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	const numRequests = 2
	const numCmt = 3
	const numRotations = 5
	const waitTimeout = 180 * time.Second

	cmtPredef := [][]int{
		{0, 1, 2, 3},
		{2, 3, 4, 5},
		{3, 4, 5, 6, 7, 8},
		{9, 4, 5, 6, 7, 8, 3},
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}
	quorumPredef := []uint16{3, 3, 5, 5, 7}
	cmt := cmtPredef[:numCmt]
	quorum := quorumPredef[:numCmt]
	addrs := make([]ledgerstate.Address, numCmt)

	var err error
	clu1 := clutest.NewCluster(t, 10)
	for i := range cmt {
		addrs[i], err = clu1.RunDKG(cmt[i], quorum[i])
		require.NoError(t, err)
		t.Logf("addr[%d]: %s", i, addrs[i].Base58())
	}

	chain1, err := clu1.DeployChain("chain", clu1.Config.AllNodes(), cmt[0], quorum[0], addrs[0])
	require.NoError(t, err)
	t.Logf("chainID: %s", chain1.ChainID.Base58())

	govClient := chain1.SCClient(governance.Interface.Hname(), chain1.OriginatorKeyPair())

	for i := range addrs {
		par := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addrs[i]).WithIotas(1)
		tx, err := govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress, *par)
		require.NoError(t, err)
		reqid := coretypes.NewRequestID(tx.ID(), 0)
		require.EqualValues(t, "", waitRequest(t, chain1, 0, reqid, waitTimeout))
		require.EqualValues(t, "", waitRequest(t, chain1, 5, reqid, waitTimeout))
		require.EqualValues(t, "", waitRequest(t, chain1, 9, reqid, waitTimeout))
		require.True(t, isAllowedStateControllerAddress(t, chain1, 0, addrs[i]))
		require.True(t, isAllowedStateControllerAddress(t, chain1, 5, addrs[i]))
		require.True(t, isAllowedStateControllerAddress(t, chain1, 9, addrs[i]))
	}

	description := "inccounter testing contract"
	programHash = inccounter.Interface.ProgramHash

	_, err = chain1.DeployContract(incCounterSCName, programHash.String(), description, nil)
	require.NoError(t, err)

	waitUntil(t, contractIsDeployed(chain1, incCounterSCName), clu1.Config.AllNodes(), 30*time.Second)

	addrIndex := 0
	kp := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(kp.PublicKey)
	err = requestFunds(clu1, myAddress, "myAddress")
	require.NoError(t, err)

	myClient := chain1.SCClient(incCounterSCHname, kp)

	for i := 0; i < numRotations; i++ {
		require.True(t, waitStateController(t, chain1, 0, addrs[addrIndex], waitTimeout))
		require.True(t, waitStateController(t, chain1, 4, addrs[addrIndex], waitTimeout))
		require.True(t, waitStateController(t, chain1, 9, addrs[addrIndex], waitTimeout))

		for j := 0; j < numRequests; j++ {
			_, err = myClient.PostRequest(inccounter.FuncIncCounter)
			require.NoError(t, err)
		}

		waitUntil(t, counterEquals(chain1, int64(numRequests*(i+1))), []int{0, 3, 8, 9}, 30*time.Second)

		addrIndex = (addrIndex + 1) % numCmt

		par := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addrs[addrIndex]).WithIotas(1)
		tx, err := govClient.PostRequest(governance.FuncRotateStateController, *par)
		require.NoError(t, err)
		reqid := coretypes.NewRequestID(tx.ID(), 0)
		require.EqualValues(t, "", waitRequest(t, chain1, 0, reqid, waitTimeout))
		require.EqualValues(t, "", waitRequest(t, chain1, 4, reqid, waitTimeout))
		require.EqualValues(t, "", waitRequest(t, chain1, 9, reqid, waitTimeout))

		require.True(t, waitStateController(t, chain1, 0, addrs[addrIndex], waitTimeout))
		require.True(t, waitStateController(t, chain1, 4, addrs[addrIndex], waitTimeout))
		require.True(t, waitStateController(t, chain1, 9, addrs[addrIndex], waitTimeout))
	}
}

func waitRequest(t *testing.T, chain *cluster.Chain, nodeIndex int, reqid coretypes.RequestID, timeout time.Duration) string {
	var ret string
	succ := waitTrue(timeout, func() bool {
		rec, err := callGetRequestRecord(t, chain, nodeIndex, reqid)
		if err == nil && rec != nil {
			ret = string(rec.LogData)
			return true
		}
		return false
	})
	if !succ {
		return "(timeout)"
	}
	return ret
}

func waitBlockIndex(t *testing.T, chain *cluster.Chain, nodeIndex int, blockIndex uint32, timeout time.Duration) bool { //nolint:unparam // (timeout is always 5s)
	return waitTrue(timeout, func() bool {
		i, err := callGetBlockIndex(t, chain, nodeIndex)
		return err == nil && i >= blockIndex
	})
}

func callGetBlockIndex(t *testing.T, chain *cluster.Chain, nodeIndex int) (uint32, error) {
	ret, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID,
		blocklog.Interface.Hname(),
		blocklog.FuncGetLatestBlockInfo,
	)
	if err != nil {
		return 0, err
	}
	v, ok, err := codec.DecodeUint32(ret.MustGet(blocklog.ParamBlockIndex))
	require.NoError(t, err)
	require.True(t, ok)
	return v, nil
}

func callGetRequestRecord(t *testing.T, chain *cluster.Chain, nodeIndex int, reqid coretypes.RequestID) (*blocklog.RequestLogRecord, error) {
	args := dict.New()
	args.Set(blocklog.ParamRequestID, reqid.Bytes())

	res, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID,
		blocklog.Interface.Hname(),
		blocklog.FuncGetRequestLogRecord,
		args,
	)
	if err != nil {
		return nil, xerrors.New("not found")
	}
	if len(res) == 0 {
		return nil, nil
	}
	recBin := res.MustGet(blocklog.ParamRequestRecord)
	rec, err := blocklog.RequestLogRecordFromBytes(recBin)
	require.NoError(t, err)
	return rec, nil
}

func waitStateController(t *testing.T, chain *cluster.Chain, nodeIndex int, addr ledgerstate.Address, timeout time.Duration) bool {
	return waitTrue(timeout, func() bool {
		a, err := callGetStateController(t, chain, nodeIndex)
		return err == nil && a.Equals(addr)
	})
}

func callGetStateController(t *testing.T, chain *cluster.Chain, nodeIndex int) (ledgerstate.Address, error) {
	ret, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID,
		blocklog.Interface.Hname(),
		blocklog.FuncControlAddresses,
	)
	if err != nil {
		return nil, err
	}
	addr, ok, err := codec.DecodeAddress(ret.MustGet(blocklog.ParamStateControllerAddress))
	require.NoError(t, err)
	require.True(t, ok)
	return addr, nil
}

func isAllowedStateControllerAddress(t *testing.T, chain *cluster.Chain, nodeIndex int, addr ledgerstate.Address) bool {
	ret, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID,
		governance.Interface.Hname(),
		governance.FuncGetAllowedStateControllerAddresses,
	)
	require.NoError(t, err)
	arr := collections.NewArray16ReadOnly(ret, governance.ParamAllowedStateControllerAddresses)
	arrlen := arr.MustLen()
	if arrlen == 0 {
		return false
	}
	for i := uint16(0); i < arrlen; i++ {
		a, ok, err := codec.DecodeAddress(arr.MustGetAt(i))
		require.NoError(t, err)
		require.True(t, ok)
		if a.Equals(addr) {
			return true
		}
	}
	return false
}
