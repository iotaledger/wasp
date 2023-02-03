package chainclient

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
)

// UploadBlob sends an off-ledger request to call 'store' in the blob contract.
func (c *Client) UploadBlob(ctx context.Context, fields dict.Dict) (hashing.HashValue, isc.OffLedgerRequest, *apiclient.ReceiptResponse, error) {
	blobHash := blob.MustGetBlobHash(fields)

	req, err := c.PostOffLedgerRequest(
		blob.Contract.Hname(),
		blob.FuncStoreBlob.Hname(),
		PostRequestParams{
			Args: fields,
		},
	)
	if err != nil {
		return hashing.NilHash, nil, nil, err
	}

	receipt, _, err := c.WaspClient.RequestsApi.WaitForRequest(ctx, c.ChainID.String(), req.ID().String()).Execute()

	return blobHash, req, receipt, err
}
