package chainclient

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

// CheckRequestResult fetches the receipt for the given request ID, and returns
// an error indicating whether the request was processed successfully.
func (c *Client) CheckRequestResult(reqID isc.RequestID) error {
	ret, err := c.CallView(blocklog.Contract.Hname(), blocklog.ViewGetRequestReceipt.Name, dict.Dict{
		blocklog.ParamRequestID: codec.EncodeRequestID(reqID),
	})
	if err != nil {
		return fmt.Errorf("could not fetch receipt for request: %w", err)
	}
	if !ret.MustHas(blocklog.ParamRequestRecord) {
		return errors.New("could not fetch receipt for request: not found in blocklog")
	}
	req, err := blocklog.RequestReceiptFromBytes(ret.MustGet(blocklog.ParamRequestRecord))
	if err != nil {
		return fmt.Errorf("could not decode receipt for request: %w", err)
	}
	if req.Error != nil {
		return fmt.Errorf("the request was rejected: %v", req.Error)
	}
	return nil
}
