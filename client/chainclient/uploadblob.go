package chainclient

import (
	"time"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
)

// UploadBlob sends an off-ledger request to call 'store' in the blob contract.
func (c *Client) UploadBlob(fields dict.Dict) (hashing.HashValue, *request.OffLedger, error) {
	blobHash := blob.MustGetBlobHash(fields)

	req, err := c.PostOffLedgerRequest(
		blob.Contract.Hname(),
		blob.FuncStoreBlob.Hname(),
		PostRequestParams{
			Args: requestargs.New().AddEncodeSimpleMany(fields),
		},
	)
	if err != nil {
		return hashing.NilHash, nil, err
	}

	err = c.WaspClient.WaitUntilRequestProcessed(c.ChainID, req.ID(), 2*time.Minute)
	return blobHash, req, err
}
