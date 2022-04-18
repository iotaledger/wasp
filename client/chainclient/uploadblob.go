package chainclient

import (
	"math"
	"time"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
)

// UploadBlob sends an off-ledger request to call 'store' in the blob contract.
func (c *Client) UploadBlob(fields dict.Dict) (hashing.HashValue, *iscp.OffLedgerRequestData, error) {
	blobHash := blob.MustGetBlobHash(fields)

	req, err := c.PostOffLedgerRequest(
		blob.Contract.Hname(),
		blob.FuncStoreBlob.Hname(),
		PostRequestParams{
			Args:      fields,
			GasBudget: math.MaxUint64,
		},
	)
	if err != nil {
		return hashing.NilHash, nil, err
	}

	err = c.WaspClient.WaitUntilRequestProcessed(c.ChainID, req.ID(), 2*time.Minute)
	return blobHash, req, err
}
