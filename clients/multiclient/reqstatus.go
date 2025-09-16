package multiclient

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

// WaitUntilRequestProcessed blocks until the request has been processed by all nodes
func (m *MultiClient) WaitUntilRequestProcessed(ctx context.Context, chainID isc.ChainID, reqID isc.RequestID, waitForL1Confirmation bool, timeout time.Duration) (*apiclient.ReceiptResponse, error) {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout

	receipts := make([]*apiclient.ReceiptResponse, len(m.nodes))
	err := m.Do(func(i int, w *apiclient.APIClient) error {
		receipt, _, err := w.ChainsAPI.WaitForRequest(ctx, reqID.String()).
			WaitForL1Confirmation(waitForL1Confirmation).
			Execute()
		receipts[i] = receipt
		return err
	})
	if err != nil {
		// Add some context info to the error.
		err = fmt.Errorf("failed WaitUntilRequestProcessed(reqID=%v): %w", reqID, err)
	}
	return receipts[0], err
}

// WaitUntilRequestProcessedSuccessfully is similar to WaitUntilRequestProcessed,
// but also checks the receipt and return an error if the request was processed with an error
func (m *MultiClient) WaitUntilRequestProcessedSuccessfully(ctx context.Context, chainID isc.ChainID, reqID isc.RequestID, waitForL1Confirmation bool, timeout time.Duration) (*apiclient.ReceiptResponse, error) {
	receipt, err := m.WaitUntilRequestProcessed(ctx, chainID, reqID, waitForL1Confirmation, timeout)
	if err != nil {
		return receipt, err
	}
	if receipt.ErrorMessage != nil {
		return receipt, fmt.Errorf("request processed with an error: %s", *receipt.ErrorMessage)
	}
	return receipt, nil
}

// WaitUntilEVMRequestProcessedSuccessfully is similar to WaitUntilRequestProcessed,
// but also checks the receipt and return an error if the request was processed with an error
func (m *MultiClient) WaitUntilEVMRequestProcessedSuccessfully(ctx context.Context, chainID isc.ChainID, txHash common.Hash, waitForL1Confirmation bool, timeout time.Duration) (*apiclient.ReceiptResponse, error) {
	requestID := isc.RequestIDFromEVMTxHash(txHash)
	receipt, err := m.WaitUntilRequestProcessed(ctx, chainID, requestID, waitForL1Confirmation, timeout)
	if err != nil {
		return receipt, err
	}
	if receipt.ErrorMessage != nil {
		return receipt, fmt.Errorf("request processed with an error: %s", *receipt.ErrorMessage)
	}
	return receipt, nil
}

// WaitUntilAllRequestsProcessed blocks until all requests in the given transaction have been processed
// by all nodes
func (m *MultiClient) WaitUntilAllRequestsProcessed(ctx context.Context, chainID isc.ChainID, tx *iotajsonrpc.IotaTransactionBlockResponse, waitForL1Confirmation bool, timeout time.Duration) ([]*apiclient.ReceiptResponse, error) {
	oldTimeout := m.Timeout
	defer func() { m.Timeout = oldTimeout }()

	m.Timeout = timeout

	receipts := make([][]*apiclient.ReceiptResponse, len(m.nodes))
	err := m.Do(func(i int, w *apiclient.APIClient) error {
		r, err := apiextensions.APIWaitUntilAllRequestsProcessed(ctx, w, tx, waitForL1Confirmation, timeout)
		receipts[i] = r
		return err
	})

	return receipts[0], err
}

// WaitUntilAllRequestsProcessedSuccessfully is similar to WaitUntilAllRequestsProcessed
// but also checks the receipts and return an error if any of the requests was processed with an error
func (m *MultiClient) WaitUntilAllRequestsProcessedSuccessfully(ctx context.Context, chainID isc.ChainID, tx *iotajsonrpc.IotaTransactionBlockResponse, waitForL1Confirmation bool, timeout time.Duration) ([]*apiclient.ReceiptResponse, error) {
	receipts, err := m.WaitUntilAllRequestsProcessed(ctx, chainID, tx, waitForL1Confirmation, timeout)
	if err != nil {
		return receipts, err
	}
	for i, receipt := range receipts {
		if receipt.ErrorMessage != nil {
			return receipts, fmt.Errorf("error found on receipt #%d: %s", i, *receipt.ErrorMessage)
		}
	}
	return receipts, nil
}
