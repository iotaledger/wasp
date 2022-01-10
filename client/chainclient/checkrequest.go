package chainclient

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"golang.org/x/xerrors"
)

// CheckRequestResult fetches the receipt for the given request ID, and returns
// an error indicating whether the request was processed successfully.
func (c *Client) CheckRequestResult(reqID iscp.RequestID) error {
	ret, err := c.CallView(blocklog.Contract.Hname(), blocklog.FuncGetRequestReceipt.Name, dict.Dict{
		blocklog.ParamRequestID: codec.EncodeRequestID(reqID),
	})
	if err != nil {
		return xerrors.Errorf("Could not fetch receipt for request: %w", err)
	}
	if !ret.MustHas(blocklog.ParamRequestRecord) {
		return xerrors.Errorf("Could not fetch receipt for request: not found in blocklog")
	}
	req, err := blocklog.RequestReceiptFromBytes(ret.MustGet(blocklog.ParamRequestRecord))
	if err != nil {
		return xerrors.Errorf("Could not decode receipt for request: %w", err)
	}
	if req.Error() != nil {
		return xerrors.Errorf("The request was rejected: %v", req.Error())
	}
	return nil
}
