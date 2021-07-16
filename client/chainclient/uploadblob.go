package chainclient

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
)

const optimalSize = 32

// UploadBlob implements an optimized blob upload protocol to the chain.
// It allows to avoid placing big data chunks into the request transaction
// - creates optimized RequestArgs, which contain hash references instead of too big binary parameters
// - uploads big binary data chunks to blob caches of at least `quorum` of `waspHosts` directly
// - posts a 'storeBlob' request to the 'blob' contract with optimized parameters
// - the chain reconstructs original parameters upn settlement of the request
func (c *Client) UploadBlob(fields dict.Dict, waspHosts []string, quorum int, optSize ...int) (hashing.HashValue, *ledgerstate.Transaction, error) {
	var osize int
	if len(optSize) > 0 {
		osize = optSize[0]
	}
	if osize < optimalSize {
		osize = optimalSize
	}
	argsEncoded, optimizedBlobs := requestargs.NewOptimizedRequestArgs(fields, osize)
	fieldValues := make([][]byte, 0, len(fields))
	for _, v := range optimizedBlobs {
		fieldValues = append(fieldValues, v)
	}
	nodesMultiAPI := multiclient.New(waspHosts)
	if err := nodesMultiAPI.UploadData(fieldValues, quorum); err != nil {
		return hashing.NilHash, nil, err
	}
	blobHash := blob.MustGetBlobHash(fields)

	reqTx, err := c.Post1Request(
		blob.Interface.Hname(),
		blob.FuncStoreBlob.Hname(),
		PostRequestParams{
			Args: argsEncoded,
		},
	)
	return blobHash, reqTx, err
}
