package iotaclient

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
)

type DevInspectTransactionBlockRequest struct {
	SenderAddress *iotago.Address
	TxKindBytes   iotago.Base64Data
	GasPrice      *iotajsonrpc.BigInt // optional
	Epoch         *uint64             // optional
	// additional_args // optional // FIXME
}

// The txKindBytes is `TransactionKind` in base64.
// When a `TransactionData` is given, error `Deserialization error: malformed utf8` will be returned.
// which is different from `DryRunTransaction` and `ExecuteTransactionBlock`
// `DryRunTransaction` and `ExecuteTransactionBlock` takes `TransactionData` in base64
func (c *Client) DevInspectTransactionBlock(
	ctx context.Context,
	req DevInspectTransactionBlockRequest,
) (*iotajsonrpc.DevInspectResults, error) {
	var resp iotajsonrpc.DevInspectResults
	return &resp, c.transport.Call(
		ctx,
		&resp,
		devInspectTransactionBlock,
		req.SenderAddress,
		req.TxKindBytes,
		req.GasPrice,
		req.Epoch,
	)
}

func (c *Client) DryRunTransaction(
	ctx context.Context,
	txDataBytes iotago.Base64Data,
) (*iotajsonrpc.DryRunTransactionBlockResponse, error) {
	var resp iotajsonrpc.DryRunTransactionBlockResponse
	return &resp, c.transport.Call(ctx, &resp, dryRunTransactionBlock, txDataBytes)
}

type ExecuteTransactionBlockRequest struct {
	TxDataBytes iotago.Base64Data
	Signatures  []*iotasigner.Signature
	Options     *iotajsonrpc.IotaTransactionBlockResponseOptions // optional
	RequestType iotajsonrpc.ExecuteTransactionRequestType        // optional
}

func (c *Client) ExecuteTransactionBlock(
	ctx context.Context,
	req ExecuteTransactionBlockRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	resp := &iotajsonrpc.IotaTransactionBlockResponse{}
	err := c.transport.Call(
		ctx,
		resp,
		executeTransactionBlock,
		req.TxDataBytes,
		req.Signatures,
		req.Options,
		req.RequestType,
	)
	if err != nil {
		return nil, err
	}

	if !isResponseComplete(resp, req.Options) {
		if c.WaitUntilEffectsVisible == nil && req.RequestType == iotajsonrpc.TxnRequestTypeWaitForLocalExecution {
			return resp, fmt.Errorf("failed to execute transaction: %s", resp.Digest)
		}

		resp, err = c.GetTransactionBlock(ctx, GetTransactionBlockRequest{
			Digest:  &resp.Digest,
			Options: req.Options,
		})
		if err != nil {
			return nil, fmt.Errorf("GetTransactionBlock failed: %w", err)
		}
	}

	return resp, nil
}
