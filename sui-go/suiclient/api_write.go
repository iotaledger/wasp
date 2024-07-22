package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

type DevInspectTransactionBlockRequest struct {
	SenderAddress *sui.Address
	TxKindBytes   sui.Base64Data
	GasPrice      *suijsonrpc.BigInt // optional
	Epoch         *uint64            // optional
	// additional_args // optional // FIXME
}

// The txKindBytes is `TransactionKind` in base64.
// When a `TransactionData` is given, error `Deserialization error: malformed utf8` will be returned.
// which is different from `DryRunTransaction` and `ExecuteTransactionBlock`
// `DryRunTransaction` and `ExecuteTransactionBlock` takes `TransactionData` in base64
func (c *Client) DevInspectTransactionBlock(
	ctx context.Context,
	req DevInspectTransactionBlockRequest,
) (*suijsonrpc.DevInspectResults, error) {
	var resp suijsonrpc.DevInspectResults
	return &resp, c.transport.Call(ctx, &resp, devInspectTransactionBlock, req.SenderAddress, req.TxKindBytes, req.GasPrice, req.Epoch)
}

func (c *Client) DryRunTransaction(
	ctx context.Context,
	txDataBytes sui.Base64Data,
) (*suijsonrpc.DryRunTransactionBlockResponse, error) {
	var resp suijsonrpc.DryRunTransactionBlockResponse
	return &resp, c.transport.Call(ctx, &resp, dryRunTransactionBlock, txDataBytes)
}

type ExecuteTransactionBlockRequest struct {
	TxDataBytes sui.Base64Data
	Signatures  []*suisigner.Signature
	Options     *suijsonrpc.SuiTransactionBlockResponseOptions // optional
	RequestType suijsonrpc.ExecuteTransactionRequestType       // optional
}

func (c *Client) ExecuteTransactionBlock(
	ctx context.Context,
	req ExecuteTransactionBlockRequest,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	resp := suijsonrpc.SuiTransactionBlockResponse{}
	return &resp, c.transport.Call(ctx, &resp, executeTransactionBlock, req.TxDataBytes, req.Signatures, req.Options, req.RequestType)
}
