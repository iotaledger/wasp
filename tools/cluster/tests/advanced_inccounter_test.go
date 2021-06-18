package tests

import (
	"testing"
	"time"

	"golang.org/x/xerrors"

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

	require.True(t, waitCounter(t, chain, 7, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 8, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 9, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 4, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 5, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 6, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 1, numRequests, 5*time.Second))
}

func TestRotation(t *testing.T) {
	numRequests := 8

	cmt1 := []int{0, 1, 2, 3}
	cmt2 := []int{2, 3, 4, 5}

	clu := clutest.NewCluster(t, 10)
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

	require.True(t, waitStateController(t, chain, 0, addr1, 5*time.Second))
	require.True(t, waitStateController(t, chain, 9, addr1, 5*time.Second))

	kp := wallet.KeyPair(1)
	myAddress := ledgerstate.NewED25519Address(kp.PublicKey)
	err = requestFunds(clu, myAddress, "myAddress")
	require.NoError(t, err)

	myClient := chain.SCClient(contractHname, kp)

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostRequest(inccounter.FuncIncCounter)
		require.NoError(t, err)
	}

	require.True(t, waitCounter(t, chain, 0, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 3, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 8, numRequests, 5*time.Second))
	require.True(t, waitCounter(t, chain, 9, numRequests, 5*time.Second))

	govClient := chain.SCClient(governance.Interface.Hname(), chain.OriginatorKeyPair())

	params := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addr2).WithIotas(1)
	tx, err := govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress, *params)

	require.NoError(t, err)

	require.True(t, waitBlockIndex(t, chain, 9, 4, 5*time.Second))
	require.True(t, waitBlockIndex(t, chain, 0, 4, 5*time.Second))
	require.True(t, waitBlockIndex(t, chain, 6, 4, 5*time.Second))

	reqid := coretypes.NewRequestID(tx.ID(), 0)
	require.EqualValues(t, "", waitRequest(t, chain, 0, reqid, 5*time.Second))
	require.EqualValues(t, "", waitRequest(t, chain, 9, reqid, 5*time.Second))

	require.NoError(t, err)
	require.True(t, isAllowedStateControllerAddresses(t, chain, 0, addr2))

	require.True(t, waitStateController(t, chain, 0, addr1, 5*time.Second))
	require.True(t, waitStateController(t, chain, 9, addr1, 5*time.Second))

	params = chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addr2).WithIotas(1)
	tx, err = govClient.PostRequest(governance.FuncRotateStateController, *params)

	require.True(t, waitStateController(t, chain, 0, addr2, 5*time.Second))
	require.True(t, waitStateController(t, chain, 9, addr2, 5*time.Second))

	require.True(t, waitBlockIndex(t, chain, 9, 5, 5*time.Second))
	require.True(t, waitBlockIndex(t, chain, 0, 5, 5*time.Second))
	require.True(t, waitBlockIndex(t, chain, 6, 5, 5*time.Second))

	reqid = coretypes.NewRequestID(tx.ID(), 0)
	require.EqualValues(t, "", waitRequest(t, chain, 0, reqid, 5*time.Second))
	require.EqualValues(t, "", waitRequest(t, chain, 9, reqid, 5*time.Second))
}

func waitTrue(timeout time.Duration, fun func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if fun() {
			return true
		}
		time.Sleep(30 * time.Millisecond)
		if time.Now().After(deadline) {
			return false
		}
	}
}

func waitRequest(t *testing.T, chain *cluster.Chain, nodeIndex int, reqid coretypes.RequestID, timeout time.Duration) string {
	var ret string
	var err error
	succ := waitTrue(timeout, func() bool {
		ret, err = callGetRequestRecord(t, chain, nodeIndex, reqid)
		return err == nil
	})
	if !succ {
		return "(timeout)"
	}
	return ret
}

func waitCounter(t *testing.T, chain *cluster.Chain, nodeIndex int, counter int, timeout time.Duration) bool {
	return waitTrue(timeout, func() bool {
		c, err := callGetCounter(t, chain, nodeIndex)
		return err == nil && c >= int64(counter)
	})
}

func waitBlockIndex(t *testing.T, chain *cluster.Chain, nodeIndex int, blockIndex uint32, timeout time.Duration) bool {
	return waitTrue(timeout, func() bool {
		i, err := callGetBlockIndex(t, chain, nodeIndex)
		return err == nil && i >= blockIndex
	})
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

func callGetBlockIndex(t *testing.T, chain *cluster.Chain, nodeIndex int) (uint32, error) {
	ret, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID,
		blocklog.Interface.Hname(),
		blocklog.FuncGetLatestBlockInfo,
	)
	if err != nil {
		return 0, err
	}
	v, ok, err := codec.DecodeUint64(ret.MustGet(blocklog.ParamBlockIndex))
	require.NoError(t, err)
	require.True(t, ok)
	return uint32(v), nil
}

func callGetRequestRecord(t *testing.T, chain *cluster.Chain, nodeIndex int, reqid coretypes.RequestID) (string, error) {
	args := dict.New()
	args.Set(blocklog.ParamRequestID, reqid.Bytes())

	res, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID,
		blocklog.Interface.Hname(),
		blocklog.FuncGetRequestLogRecord,
		args,
	)
	if err != nil {
		return "", xerrors.New("not found")
	}
	if len(res) == 0 {
		return "", nil
	}
	recBin := res.MustGet(blocklog.ParamRequestRecord)
	rec, err := blocklog.RequestLogRecordFromBytes(recBin)
	require.NoError(t, err)
	return string(rec.LogData), nil
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

func isAllowedStateControllerAddresses(t *testing.T, chain *cluster.Chain, nodeIndex int, addr ledgerstate.Address) bool {
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
