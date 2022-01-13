package tests

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"os"
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/stretchr/testify/require"
)

var (
	testOwner = cryptolib.NewKeyPairFromSeed(wallet.SubSeed(1))
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
			Args: fv,
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
		blobs[i], err = os.ReadFile(fileNames[i])
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

	reqTx, err := chClient.DepositFunds(100)
	require.NoError(t, err)
	err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessed(e.chain.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	expectedHash, _, err := chClient.UploadBlob(fv)
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
