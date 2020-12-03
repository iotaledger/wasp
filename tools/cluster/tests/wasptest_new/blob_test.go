package wasptest

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

var (
	testOwner   = wallet.WithIndex(1)
	mySigScheme = testOwner.SigScheme()
	myAddress   = testOwner.Address()
)

func setupBlobTest(t *testing.T) *cluster.Chain {
	setup(t, "test_cluster")

	chain, err := clu.DeployDefaultChain()
	check(err, t)

	chain.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state dict.Dict) bool {
		require.EqualValues(t, 1, blockIndex)
		checkRoots(t, chain)
		contractRegistry := datatypes.NewMustMap(state, root.VarContractRegistry)
		require.EqualValues(t, 3, contractRegistry.Len())
		return true
	})

	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	if !clu.VerifyAddressBalances(myAddress, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "myAddrress after request funds") {
		t.Fail()
	}
	return chain
}

func getBlobInfo(t *testing.T, chain *cluster.Chain, hash hashing.HashValue) dict.Dict {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(blob.Interface.Hname()),
		blob.FuncGetBlobInfo,
		dict.FromGoMap(map[kv.Key][]byte{
			blob.ParamHash: hash[:],
		}),
	)
	check(err, t)
	return ret
}

func getBlobFieldValue(t *testing.T, chain *cluster.Chain, blobHash hashing.HashValue, field string) []byte {
	v, err := chain.Cluster.WaspClient(0).CallView(
		chain.ContractID(blob.Interface.Hname()),
		blob.FuncGetBlobField,
		dict.FromGoMap(map[kv.Key][]byte{
			blob.ParamHash:  blobHash[:],
			blob.ParamField: []byte(field),
		}),
	)
	check(err, t)
	if v.IsEmpty() {
		return nil
	}
	ret, err := v.Get(blob.ParamBytes)
	check(err, t)
	return ret
}

func TestBlobDeployChain(t *testing.T) {
	chain := setupBlobTest(t)

	ret := getBlobInfo(t, chain, *hashing.NilHash)
	require.True(t, ret.IsEmpty())
}

func TestBlobStoreSmallBlob(t *testing.T) {
	chain := setupBlobTest(t)

	description := "testing the blob"
	blobFieldValues := map[string]interface{}{
		blob.VarFieldProgramDescription: []byte(description),
	}
	expectedHash := blob.MustGetBlobHash(codec.MakeDict(blobFieldValues))
	t.Logf("expected hash: %s", expectedHash.String())

	chClient := chainclient.New(clu.NodeClient, clu.WaspClient(0), chain.ChainID, mySigScheme)
	reqTx, err := chClient.PostRequest(
		blob.Interface.Hname(),
		coretypes.Hn(blob.FuncStoreBlob),
		chainclient.PostRequestParams{
			Args: codec.MakeDict(blobFieldValues),
		},
	)
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(reqTx, 30*time.Second)
	check(err, t)

	ret := getBlobInfo(t, chain, expectedHash)
	require.False(t, ret.IsEmpty())

	v, ok, err := codec.DecodeInt64(ret.MustGet(blob.VarFieldProgramDescription))
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, len(description), v)

	retBin := getBlobFieldValue(t, chain, expectedHash, blob.VarFieldProgramDescription)
	require.NotNil(t, retBin)
	require.EqualValues(t, []byte(description), retBin)
}

func TestBlobStoreManyBlobs(t *testing.T) {
	chain := setupBlobTest(t)

	var err error
	fileNames := []string{"blob_test.go", "deploy_test.go", "inccounter_test.go"}
	blobs := make([][]byte, len(fileNames))
	for i := range fileNames {
		blobs[i], err = ioutil.ReadFile(fileNames[i])
		check(err, t)
	}
	blobFieldValues := make(map[string]interface{})
	for i, fn := range fileNames {
		blobFieldValues[fn] = blobs[i]
	}

	expectedHash := blob.MustGetBlobHash(codec.MakeDict(blobFieldValues))
	t.Logf("expected hash: %s", expectedHash.String())

	chClient := chainclient.New(clu.NodeClient, clu.WaspClient(0), chain.ChainID, mySigScheme)
	reqTx, err := chClient.PostRequest(
		blob.Interface.Hname(),
		coretypes.Hn(blob.FuncStoreBlob),
		chainclient.PostRequestParams{
			Args: codec.MakeDict(blobFieldValues),
		},
	)
	check(err, t)
	err = chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(reqTx, 30*time.Second)
	check(err, t)

	ret := getBlobInfo(t, chain, expectedHash)
	require.False(t, ret.IsEmpty())

	for i, fn := range fileNames {
		v, err := ret.Get(kv.Key(fn))
		require.NoError(t, err)
		require.NotNil(t, v)
		require.EqualValues(t, len(blobs[i]), util.MustUint64From8Bytes(v))
		fmt.Printf("    %s: %d\n", fn, len(blobs[i]))

		fdata := getBlobFieldValue(t, chain, expectedHash, fn)
		require.NotNil(t, fdata)
		require.EqualValues(t, fdata, blobs[i])
	}
}
