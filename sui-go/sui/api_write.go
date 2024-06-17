package sui

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

// The txKindBytes is `TransactionKind` in base64.
// When a `TransactionData` is given, error `Deserialization error: malformed utf8` will be returned.
// which is different from `DryRunTransaction` and `ExecuteTransactionBlock`
// `DryRunTransaction` and `ExecuteTransactionBlock` takes `TransactionData` in base64
func (s *ImplSuiAPI) DevInspectTransactionBlock(
	ctx context.Context,
	senderAddress *sui_types.SuiAddress,
	txKindBytes sui_types.Base64Data,
	gasPrice *models.SafeSuiBigInt[uint64], // optional
	epoch *uint64, // optional
) (*models.DevInspectResults, error) {
	var resp models.DevInspectResults
	return &resp, s.http.CallContext(ctx, &resp, devInspectTransactionBlock, senderAddress, txKindBytes, gasPrice, epoch)
}

func (s *ImplSuiAPI) DryRunTransaction(
	ctx context.Context,
	txDataBytes sui_types.Base64Data,
) (*models.DryRunTransactionBlockResponse, error) {
	var resp models.DryRunTransactionBlockResponse
	return &resp, s.http.CallContext(ctx, &resp, dryRunTransactionBlock, txDataBytes)
}

func (s *ImplSuiAPI) ExecuteTransactionBlock(
	ctx context.Context,
	txDataBytes sui_types.Base64Data,
	signatures []*sui_signer.Signature,
	options *models.SuiTransactionBlockResponseOptions,
	requestType models.ExecuteTransactionRequestType,
) (*models.SuiTransactionBlockResponse, error) {
	resp := models.SuiTransactionBlockResponse{}
	return &resp, s.http.CallContext(ctx, &resp, executeTransactionBlock, txDataBytes, signatures, options, requestType)
}
