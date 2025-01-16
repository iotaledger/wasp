package iotaclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
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
	resp := iotajsonrpc.IotaTransactionBlockResponse{}
	err := c.transport.Call(
		ctx,
		&resp,
		executeTransactionBlock,
		req.TxDataBytes,
		req.Signatures,
		req.Options,
		req.RequestType,
	)
	if err != nil {
		return nil, err
	}
	if req.RequestType == iotajsonrpc.TxnRequestTypeWaitForLocalExecution && c.WaitUntilEffectsVisible != nil {
		return c.waitUntilEffectsVisible(ctx, &resp, c.WaitUntilEffectsVisible)
	}
	return &resp, nil
}

func (c *Client) waitUntilEffectsVisible(
	ctx context.Context,
	resp *iotajsonrpc.IotaTransactionBlockResponse,
	waitParams *WaitParams,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if resp.ConfirmedLocalExecution == nil {
		panic("expected ConfirmedLocalExecution != nil")
	}
	if *resp.ConfirmedLocalExecution {
		// effects are already visible
		return resp, nil
	}
	attempts := waitParams.Attempts
	for {
		time.Sleep(waitParams.DelayBetweenAttempts)
		_, err := c.GetTransactionBlock(ctx, GetTransactionBlockRequest{Digest: &resp.Digest})
		if err == nil {
			return resp, nil
		}
		if !strings.Contains(err.Error(), "Could not find the referenced transaction") {
			return nil, err
		}
		attempts--
		if attempts == 0 {
			return nil, fmt.Errorf("timed out waiting for the transaction effects to be visible")
		}
	}
}
