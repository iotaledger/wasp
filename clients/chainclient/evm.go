package chainclient

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/isc"
)

func (c *Client) RequestIDByEVMTransactionHash(ctx context.Context, txHash common.Hash) (isc.RequestID, error) {
	requestIDStr, _, err := c.WaspClient.ChainsApi.GetRequestIDFromEVMTransactionID(ctx, c.ChainID.String(), txHash.String()).Execute()

	if err != nil {
		return isc.RequestID{}, err
	}

	requestID, err := isc.RequestIDFromString(requestIDStr)

	return requestID, err
}
