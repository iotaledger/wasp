package tests

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/stretchr/testify/require"
)

var (
	testOwner = wallet.KeyPair(1)
	myAddress = ledgerstate.NewED25519Address(testOwner.PublicKey)
)

func setupBlobTest(t *testing.T) *chainEnv {
	e := setupWithNoChain(t)

	chain, err := e.clu.DeployDefaultChain()
	require.NoError(t, err)

	chEnv := newChainEnv(t, e.clu, chain)

	chEnv.checkCoreContracts()
	for _, i := range chain.CommitteeNodes {
		blockIndex, err := chain.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 1, blockIndex)
	}

	e.requestFunds(myAddress, "myAddress")

	if !e.clu.VerifyAddressBalances(myAddress, solo.Saldo,
		colored.NewBalancesForIotas(solo.Saldo),
		"myAddress after request funds") {
		t.Fail()
	}
	return chEnv
}

func (e *chainEnv) getBlobInfo(hash hashing.HashValue) map[string]uint32 {
	ret, err := e.chain.Cluster.WaspClient(0).CallView(
		e.chain.ChainID, blob.Contract.Hname(), blob.FuncGetBlobInfo.Name,
		dict.Dict{
			blob.ParamHash: hash[:],
		})
	require.NoError(e.t, err)
	decoded, err := blob.DecodeSizesMap(ret)
	require.NoError(e.t, err)
	return decoded
}

func (e *chainEnv) getBlobFieldValue(blobHash hashing.HashValue, field string) []byte {
	v, err := e.chain.Cluster.WaspClient(0).CallView(
		e.chain.ChainID, blob.Contract.Hname(), blob.FuncGetBlobField.Name,
		dict.Dict{
			blob.ParamHash:  blobHash[:],
			blob.ParamField: []byte(field),
		})
	require.NoError(e.t, err)
	if v.IsEmpty() {
		return nil
	}
	ret, err := v.Get(blob.ParamBytes)
	require.NoError(e.t, err)
	return ret
}

func TestBlobDeployChain(t *testing.T) {
	e := setupBlobTest(t)
	ret := e.getBlobInfo(hashing.NilHash)
	require.Len(t, ret, 0)
}

func TestBlobStoreSmallBlob(t *testing.T) {
	e := setupBlobTest(t)

	description := "testing the blob"
	fv := codec.MakeDict(map[string]interface{}{
		blob.VarFieldProgramDescription: []byte(description),
	})
	expectedHash := blob.MustGetBlobHash(fv)
	t.Logf("expected hash: %s", expectedHash.String())

	chClient := chainclient.New(e.clu.GoshimmerClient(), e.clu.WaspClient(0), e.chain.ChainID, testOwner)
	reqTx, err := chClient.Post1Request(
		blob.Contract.Hname(),
		blob.FuncStoreBlob.Hname(),
		chainclient.PostRequestParams{
			Args: requestargs.New().AddEncodeSimpleMany(fv),
		},
	)
	require.NoError(t, err)
	err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	sizes := e.getBlobInfo(expectedHash)
	require.NotZero(t, len(sizes))

	require.EqualValues(t, len(description), sizes[blob.VarFieldProgramDescription])

	retBin := e.getBlobFieldValue(expectedHash, blob.VarFieldProgramDescription)
	require.NotNil(t, retBin)
	require.EqualValues(t, []byte(description), retBin)
}

func TestBlobStoreManyBlobsNoEncoding(t *testing.T) {
	e := setupBlobTest(t)

	var err error
	fileNames := []string{"blob_test.go", "deploy_test.go", "inccounter_test.go", "account_test.go"}
	blobs := make([][]byte, len(fileNames))
	for i := range fileNames {
		blobs[i], err = ioutil.ReadFile(fileNames[i])
		require.NoError(t, err)
	}
	blobFieldValues := make(map[string]interface{})
	totalSize := 0
	for i, fn := range fileNames {
		blobFieldValues[fn] = blobs[i]
		totalSize += len(blobs[i])
	}
	t.Logf("================= total size: %d. Files: %+v", totalSize, fileNames)

	fv := codec.MakeDict(blobFieldValues)
	chClient := chainclient.New(e.clu.GoshimmerClient(), e.clu.WaspClient(0), e.chain.ChainID, testOwner)
	expectedHash, tx, err := chClient.UploadBlob(fv, e.clu.Config.APIHosts(e.clu.Config.AllNodes()), int(e.chain.Quorum))
	require.NoError(t, err)
	err = chClient.WaspClient.WaitUntilAllRequestsProcessed(e.chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)
	t.Logf("expected hash: %s", expectedHash.String())

	sizes := e.getBlobInfo(expectedHash)
	require.NotZero(t, len(sizes))

	for i, fn := range fileNames {
		v := sizes[fn]
		require.EqualValues(t, len(blobs[i]), v)
		fmt.Printf("    %s: %d\n", fn, len(blobs[i]))

		fdata := e.getBlobFieldValue(expectedHash, fn)
		require.NotNil(t, fdata)
		require.EqualValues(t, fdata, blobs[i])
	}
}

func TestBlobRefConsensus(t *testing.T) {
	e := setupBlobTest(t)

	fileNames := []string{"blob_test.go", "deploy_test.go", "inccounter_test.go", "account_test.go"}
	blobs := make([][]byte, len(fileNames))
	for i := range fileNames {
		var err error
		blobs[i], err = ioutil.ReadFile(fileNames[i])
		require.NoError(t, err)
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
	chClient := chainclient.New(e.clu.GoshimmerClient(), e.clu.WaspClient(0), e.chain.ChainID, testOwner)
	reqTx, err := chClient.Post1Request(
		blob.Contract.Hname(),
		blob.FuncStoreBlob.Hname(),
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
	nodesMultiAPI := multiclient.New(e.clu.Config.APIHosts(e.clu.Config.AllNodes()))
	err = nodesMultiAPI.UploadData(fieldValues, int(e.chain.Quorum))
	require.NoError(t, err)

	// now waiting
	err = chClient.WaspClient.WaitUntilAllRequestsProcessed(e.chain.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	sizes := e.getBlobInfo(expectedHash)
	require.NotZero(t, len(sizes))

	for k, v := range blobFieldValues {
		retBin := e.getBlobFieldValue(expectedHash, k)
		require.NotNil(t, retBin)
		require.EqualValues(t, v, retBin)
	}
}
