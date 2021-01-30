package chainclient

import (
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/requestargs"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"time"
)

const optimalSize = 32

// UploadBlob sends the data to the 'blob' core contract on the chain
// The sending optimized: the request transaction contains only hashes of big data chunks,
// the data chunks itself are send directly to Wasp nodes.
// This we we avoid huge value transactions
func (c *Client) UploadBlob(fields dict.Dict, waspHosts []string, quorum int, optSize ...int) (hashing.HashValue, error) {
	var osize int
	if len(optSize) > 0 {
		osize = optSize[0]
	}
	if osize < optimalSize {
		osize = optimalSize
	}
	argsEncoded, optimizedBlobs := requestargs.NewOptimizedRequestArgs(fields)
	fieldValues := make([][]byte, 0, len(fields))
	for _, v := range optimizedBlobs {
		fieldValues = append(fieldValues, v)
	}
	nodesMultiApi := multiclient.New(waspHosts)
	if err := nodesMultiApi.UploadData(fieldValues, quorum); err != nil {
		return hashing.HashValue{}, err
	}
	blobHash := blob.MustGetBlobHash(fields)

	reqTx, err := c.PostRequest(
		blob.Interface.Hname(),
		coretypes.Hn(blob.FuncStoreBlob),
		PostRequestParams{
			Args: argsEncoded,
		},
	)
	if err != nil {
		return hashing.HashValue{}, err
	}
	err = c.WaspClient.WaitUntilAllRequestsProcessed(reqTx, 30*time.Second)
	if err != nil {
		return hashing.HashValue{}, err
	}
	return blobHash, nil
}
