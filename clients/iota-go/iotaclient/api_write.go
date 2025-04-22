package iotaclient

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
)

type DevInspectTransactionBlockRequest struct {
	SenderAddress *iotago.Address
	TxKindBytes   iotago.Base64Data
	GasPrice      *iotajsonrpc.BigInt // optional
	// The epoch to perform the call. Will be set from the system state object if not provided
	Epoch *uint64 // optional
	// Additional arguments including gas_budget, gas_objects, gas_sponsor and skip_checks
	AdditionalArgs *DevInspectArgs // optional
}

type DevInspectArgs struct {
	// The sponsor of the gas for the transaction, might be different from the sender.
	GasSponsor *iotago.Address `json:"gasSponsor,omitempty"`
	// The gas budget for the transaction.
	GasBudget *iotajsonrpc.BigInt `json:"gasBudget,omitempty"`
	// The gas objects used to pay for the transaction.
	GasObjects *[]*iotago.ObjectRef `json:"gasObjects,omitempty"`
	// Whether to skip transaction checks for the transaction.
	SkipChecks *bool `json:"skipChecks,omitempty"`
	// Whether to return the raw transaction data and effects.
	ShowRawTxnDataAndEffects *bool `json:"showRawTxnDataAndEffects,omitempty"`
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
		req.AdditionalArgs,
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
