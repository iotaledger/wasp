package tests

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

var (
	testOwner = wallet.KeyPair(1)
	myAddress = ledgerstate.NewED25519Address(testOwner.PublicKey)
)

func setupBlobTest(t *testing.T) *cluster.Chain {
	setup(t, "test_cluster")

	chain1, err := clu.DeployDefaultChain()
	check(err, t)

	chain1.WithSCState(root.Interface.Hname(), func(host string, blockIndex uint32, state dict.Dict) bool {
		require.EqualValues(t, 1, blockIndex)
		checkRoots(t, chain1)
		contractRegistry := collections.NewMapReadOnly(state, root.VarContractRegistry)
		require.EqualValues(t, 4, contractRegistry.MustLen())
		return true
	})

	err = requestFunds(clu, myAddress, "myAddress")
	check(err, t)

	if !clu.VerifyAddressBalances(myAddress, solo.Saldo, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: solo.Saldo,
	}, "myAddress after request funds") {
		t.Fail()
	}
	return chain1
}

func getBlobInfo(t *testing.T, chain *cluster.Chain, hash hashing.HashValue) map[string]uint32 {
	ret, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, blob.Interface.Hname(), blob.FuncGetBlobInfo,
		dict.Dict{
			blob.ParamHash: hash[:],
		})
	check(err, t)
	decoded, err := blob.DecodeSizesMap(ret)
	check(err, t)
	return decoded
}

func getBlobFieldValue(t *testing.T, chain *cluster.Chain, blobHash hashing.HashValue, field string) []byte {
	v, err := chain.Cluster.WaspClient(0).CallView(
		chain.ChainID, blob.Interface.Hname(), blob.FuncGetBlobField,
		dict.Dict{
			blob.ParamHash:  blobHash[:],
			blob.ParamField: []byte(field),
		})
	check(err, t)
	if v.IsEmpty() {
		return nil
	}
	ret, err := v.Get(blob.ParamBytes)
	check(err, t)
	return ret
}

func TestBlobDeployChain(t *testing.T) {
	chain1 := setupBlobTest(t)

	ret := getBlobInfo(t, chain1, hashing.NilHash)
	require.Len(t, ret, 0)
}

func TestBlobStoreSmallBlob(t *testing.T) {
	chain1 := setupBlobTest(t)

	description := "testing the blob"
	fv := codec.MakeDict(map[string]interface{}{
		blob.VarFieldProgramDescription: []byte(description),
	})
	expectedHash := blob.MustGetBlobHash(fv)
	t.Logf("expected hash: %s", expectedHash.String())

	chClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain1.ChainID, testOwner)
	reqTx, err := chClient.Post1Request(
		blob.Interface.Hname(),
		coretypes.Hn(blob.FuncStoreBlob),
		chainclient.PostRequestParams{
			Args: requestargs.New().AddEncodeSimpleMany(fv),
		},
	)
	check(err, t)
	err = chain1.CommitteeMultiClient().WaitUntilAllRequestsProcessed(chain1.ChainID, reqTx, 30*time.Second)
	check(err, t)

	sizes := getBlobInfo(t, chain1, expectedHash)
	require.NotZero(t, len(sizes))

	require.EqualValues(t, len(description), sizes[blob.VarFieldProgramDescription])

	retBin := getBlobFieldValue(t, chain1, expectedHash, blob.VarFieldProgramDescription)
	require.NotNil(t, retBin)
	require.EqualValues(t, []byte(description), retBin)
}

func TestBlobStoreManyBlobsNoEncoding(t *testing.T) {
	chain1 := setupBlobTest(t)

	var err error
	fileNames := []string{"blob_test.go", "deploy_test.go", "inccounter_test.go", "account_test.go"}
	blobs := make([][]byte, len(fileNames))
	for i := range fileNames {
		blobs[i], err = ioutil.ReadFile(fileNames[i])
		check(err, t)
	}
	blobFieldValues := make(map[string]interface{})
	totalSize := 0
	for i, fn := range fileNames {
		blobFieldValues[fn] = blobs[i]
		totalSize += len(blobs[i])
	}
	t.Logf("================= total size: %d. Files: %+v", totalSize, fileNames)

	fv := codec.MakeDict(blobFieldValues)
	chClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain1.ChainID, testOwner)
	expectedHash, tx, err := chClient.UploadBlob(fv, clu.Config.ApiHosts(clu.Config.AllNodes()), int(chain1.Quorum))
	require.NoError(t, err)
	err = chClient.WaspClient.WaitUntilAllRequestsProcessed(chain1.ChainID, tx, 30*time.Second)
	require.NoError(t, err)
	t.Logf("expected hash: %s", expectedHash.String())

	sizes := getBlobInfo(t, chain1, expectedHash)
	require.NotZero(t, len(sizes))

	for i, fn := range fileNames {
		v := sizes[fn]
		require.EqualValues(t, len(blobs[i]), v)
		fmt.Printf("    %s: %d\n", fn, len(blobs[i]))

		fdata := getBlobFieldValue(t, chain1, expectedHash, fn)
		require.NotNil(t, fdata)
		require.EqualValues(t, fdata, blobs[i])
	}
}

func TestBlobRefConsensus(t *testing.T) {
	chain1 := setupBlobTest(t)

	fileNames := []string{"blob_test.go", "deploy_test.go", "inccounter_test.go", "account_test.go"}
	blobs := make([][]byte, len(fileNames))
	for i := range fileNames {
		blobs[i], err = ioutil.ReadFile(fileNames[i])
		check(err, t)
	}
	blobFieldValues := make(map[string]interface{})
	for i, fn := range fileNames {
		blobFieldValues[fn] = blobs[i]
		t.Logf("================= uploading %s: size %d bytes", fn, len(blobs[i]))
	}

	fv := codec.MakeDict(blobFieldValues)
	expectedHash := blob.MustGetBlobHash(fv)

	// optimizing parameters
	argsEncoded, optimizedBlobs := requestargs.NewOptimizedRequestArgs(fv)

	// sending storeBlob request (data is not uploaded yet)
	chClient := chainclient.New(clu.GoshimmerClient(), clu.WaspClient(0), chain1.ChainID, testOwner)
	reqTx, err := chClient.Post1Request(
		blob.Interface.Hname(),
		coretypes.Hn(blob.FuncStoreBlob),
		chainclient.PostRequestParams{
			Args: argsEncoded,
		},
	)
	require.NoError(t, err)
	time.Sleep(10 * time.Second)
	// not waiting for the request to be processed because it is waiting for blob data to be uploaded to the cache
	// Uploading the data
	fieldValues := make([][]byte, 0, len(fv))
	for _, v := range optimizedBlobs {
		fieldValues = append(fieldValues, v)
	}
	nodesMultiAPI := multiclient.New(clu.Config.ApiHosts(clu.Config.AllNodes()))
	err = nodesMultiAPI.UploadData(fieldValues, int(chain1.Quorum))
	require.NoError(t, err)

	// now waiting
	err = chClient.WaspClient.WaitUntilAllRequestsProcessed(chain1.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	sizes := getBlobInfo(t, chain1, expectedHash)
	require.NotZero(t, len(sizes))

	for k, v := range blobFieldValues {
		retBin := getBlobFieldValue(t, chain1, expectedHash, k)
		require.NotNil(t, retBin)
		require.EqualValues(t, v, retBin)
	}
}
