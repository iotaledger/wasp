package chainclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

// CheckRequestResult fetches the receipt for the given request ID, and returns
// an error indicating whether the request was processed successfully.
func (c *Client) CheckRequestResult(ctx context.Context, reqID isc.RequestID) error {
	receipt, _, err := c.WaspClient.CorecontractsAPI.BlocklogGetRequestReceipt(ctx, reqID.String()).Execute()
	if err != nil {
		return errors.New("could not fetch receipt for request: not found in blocklog")
	}
	if receipt.ErrorMessage != nil {
		return fmt.Errorf("the request was rejected: %v", receipt.ErrorMessage)
	}
	return nil
}
