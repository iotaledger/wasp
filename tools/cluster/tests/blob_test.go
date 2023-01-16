package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
)

func (e *ChainEnv) getBlobInfo(hash hashing.HashValue) map[string]uint32 {
	ret, err := e.Chain.Cluster.WaspClient(0).CallView(
		e.Chain.ChainID, blob.Contract.Hname(), blob.ViewGetBlobInfo.Name,
		dict.Dict{
			blob.ParamHash: hash[:],
		})
	require.NoError(e.t, err)
	decoded, err := blob.DecodeSizesMap(ret)
	require.NoError(e.t, err)
	return decoded
}

func (e *ChainEnv) getBlobFieldValue(blobHash hashing.HashValue, field string) []byte {
	v, err := e.Chain.Cluster.WaspClient(0).CallView(
		e.Chain.ChainID, blob.Contract.Hname(), blob.ViewGetBlobField.Name,
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

// executed in cluster_test.go
func testBlobStoreSmallBlob(t *testing.T, e *ChainEnv) {
	ret := e.getBlobInfo(hashing.NilHash)
	require.Len(t, ret, 0)

	description := "testing the blob"
	fv := codec.MakeDict(map[string]interface{}{
		blob.VarFieldProgramDescription: []byte(description),
	})
	expectedHash := blob.MustGetBlobHash(fv)
	t.Logf("expected hash: %s", expectedHash.String())

	myWallet, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, myWallet)
	reqTx, err := chClient.Post1Request(
		blob.Contract.Hname(),
		blob.FuncStoreBlob.Hname(),
		chainclient.PostRequestParams{
			Args: fv,
		},
	)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	sizes := e.getBlobInfo(expectedHash)
	require.NotZero(t, len(sizes))

	require.EqualValues(t, len(description), sizes[blob.VarFieldProgramDescription])

	retBin := e.getBlobFieldValue(expectedHash, blob.VarFieldProgramDescription)
	require.NotNil(t, retBin)
	require.EqualValues(t, []byte(description), retBin)
}

// executed in cluster_test.go
func testBlobStoreManyBlobsNoEncoding(t *testing.T, e *ChainEnv) {
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
	myWallet, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(0), e.Chain.ChainID, myWallet)

	reqTx, err := chClient.DepositFunds(100)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	expectedHash, _, receipt, err := chClient.UploadBlob(fv)
	require.NoError(t, err)
	require.Empty(t, receipt.ResolvedError)
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
