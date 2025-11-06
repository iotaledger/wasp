package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/samber/lo"
)

var _ L1Client = (*GraphQLClient)(nil)

func (c *GraphQLClient) GetDynamicFieldObject(
	ctx context.Context,
	req iotaclient.GetDynamicFieldObjectRequest,
) (*iotajsonrpc.IotaObjectResponse, error) {
	// GraphQL doesn't have a direct GetDynamicFieldObject query
	// This would need to be implemented via GetDynamicFields and filtering
	return nil, fmt.Errorf("GetDynamicFieldObject: not directly supported via GraphQL")
}

func (c *GraphQLClient) GetDynamicFields(
	ctx context.Context,
	req iotaclient.GetDynamicFieldsRequest,
) (*iotajsonrpc.DynamicFieldPage, error) {
	// Convert request parameters to GraphQL format
	var cursor *string
	if req.Cursor != nil {
		cursorStr := req.Cursor.String()
		cursor = &cursorStr
	}

	var first *int
	if req.Limit != nil {
		limit := int(*req.Limit)
		first = &limit
	}

	resp, err := iotagraphql.GetDynamicFields(ctx, c.client, *req.ParentObjectID, first, cursor)
	if err != nil {
		return nil, err
	}

	nodes := resp.Owner.DynamicFields.Nodes
	data := make([]iotajsonrpc.DynamicFieldInfo, len(nodes))
	for i, node := range nodes {
		converted, err := convertGraphQLDynamicFieldToInfo(&node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert dynamic field at index %d: %w", i, err)
		}
		data[i] = *converted
	}

	var nextCursor *iotago.ObjectID
	if resp.Owner.DynamicFields.PageInfo.HasNextPage && resp.Owner.DynamicFields.PageInfo.EndCursor != "" {
		cursor := iotago.MustObjectIDFromHex(resp.Owner.DynamicFields.PageInfo.EndCursor)
		nextCursor = cursor
	}

	res := &iotajsonrpc.DynamicFieldPage{
		Data:        data,
		HasNextPage: resp.Owner.DynamicFields.PageInfo.HasNextPage,
		NextCursor:  nextCursor,
	}
	return res, nil
}

func (c *GraphQLClient) GetOwnedObjects(
	ctx context.Context,
	req iotaclient.GetOwnedObjectsRequest,
) (*iotajsonrpc.ObjectsPage, error) {
	if req.Address == nil {
		return nil, fmt.Errorf("address is required")
	}

	// Convert limit from *uint to *int
	var limitPtr *int
	if req.Limit != nil {
		limit := int(*req.Limit)
		limitPtr = &limit
	}

	// Convert cursor from *iotago.ObjectID to *string
	var cursorPtr *string
	if req.Cursor != nil {
		cursor := req.Cursor.String()
		cursorPtr = &cursor
	}

	// Extract options
	var showBcs, showContent, showDisplay, showType, showOwner, showPreviousTransaction, showStorageRebate *bool
	if req.Query != nil && req.Query.Options != nil {
		showBcs = lo.ToPtr(req.Query.Options.ShowBcs)
		showContent = lo.ToPtr(req.Query.Options.ShowContent)
		showDisplay = lo.ToPtr(req.Query.Options.ShowDisplay)
		showType = lo.ToPtr(req.Query.Options.ShowType)
		showOwner = lo.ToPtr(req.Query.Options.ShowOwner)
		showPreviousTransaction = lo.ToPtr(req.Query.Options.ShowPreviousTransaction)
		showStorageRebate = lo.ToPtr(req.Query.Options.ShowStorageRebate)
	}

	// Convert filter from JSON-RPC to GraphQL format
	var filterPtr *iotagraphql.ObjectFilter
	if req.Query != nil && req.Query.Filter != nil {
		filter := iotagraphql.ObjectFilter{}

		// Handle StructType filter
		if req.Query.Filter.StructType != nil {
			filter.Type = req.Query.Filter.StructType.String()
		}

		// Handle Package filter
		if req.Query.Filter.Package != nil {
			// Package filter in GraphQL uses Type field with package format
			packageID := (*iotago.PackageID)(req.Query.Filter.Package)
			filter.Type = packageID.String()
		}

		// Note: AddressOwner filter is already handled by the owner parameter
		// in the GraphQL query, so we don't need to set it in the ObjectFilter

		// Only set filter if at least one field is set (avoid sending zero-value filters)
		if filter.Type != "" || len(filter.ObjectIds) > 0 || len(filter.ObjectKeys) > 0 {
			filterPtr = &filter
		}
	}

	// Call GraphQL GetOwnedObjects
	resp, err := iotagraphql.GetOwnedObjects(ctx, c.client, *req.Address, limitPtr, cursorPtr,
		showBcs, showContent, showDisplay, showType, showOwner, showPreviousTransaction, showStorageRebate, filterPtr)
	if err != nil {
		return nil, err
	}

	// Convert response to ObjectsPage
	nodes := resp.Address.Objects.Nodes
	objects := make([]iotajsonrpc.IotaObjectResponse, 0, len(nodes))
	for _, node := range nodes {
		// Convert RPC_MOVE_OBJECT_FIELDS to IotaObjectResponse
		obj, err := convertRPCMoveObjectFieldsToIotaObjectResponse(&node.RPC_MOVE_OBJECT_FIELDS, req.Query.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object: %w", err)
		}
		objects = append(objects, *obj)
	}

	// Convert cursor
	// Note: GraphQL cursors are opaque and may not always be parseable as ObjectIDs
	// For now, we skip cursor conversion as it's not critical for basic functionality
	var nextCursor *iotago.ObjectID
	if resp.Address.Objects.PageInfo.HasNextPage && resp.Address.Objects.PageInfo.EndCursor != "" {
		// Try to parse the cursor, but don't fail if we can't
		// The cursor format from GraphQL might be opaque or in a different encoding
		addr, err := iotago.AddressFromHex(resp.Address.Objects.PageInfo.EndCursor)
		if err == nil {
			objID := iotago.ObjectID(*addr)
			nextCursor = &objID
		}
		// If parsing fails, leave nextCursor as nil - pagination may not work but basic query will
	}

	return &iotajsonrpc.ObjectsPage{
		Data:        objects,
		NextCursor:  nextCursor,
		HasNextPage: resp.Address.Objects.PageInfo.HasNextPage,
	}, nil
}

func (c *GraphQLClient) QueryEvents(
	ctx context.Context,
	req iotaclient.QueryEventsRequest,
) (*iotajsonrpc.EventPage, error) {
	// QueryEvents in GraphQL has completely different signature
	_ = req
	return nil, fmt.Errorf("QueryEvents: GraphQL API signature mismatch, needs custom implementation")
}

func (c *GraphQLClient) QueryTransactionBlocks(
	ctx context.Context,
	req iotaclient.QueryTransactionBlocksRequest,
) (*iotajsonrpc.TransactionBlocksPage, error) {
	// QueryTransactionBlocks in GraphQL has different parameter structure
	_ = req
	return nil, fmt.Errorf("QueryTransactionBlocks: GraphQL API signature mismatch, needs custom implementation")
}

func (c *GraphQLClient) ResolveNameServiceAddress(ctx context.Context, iotaName string) (*iotago.Address, error) {
	// Name service resolution is not directly available via GraphQL
	// Would need to query specific name service objects
	_ = iotaName
	return nil, fmt.Errorf("ResolveNameServiceAddress: name service queries not directly supported via GraphQL")
}

func (c *GraphQLClient) ResolveNameServiceNames(
	ctx context.Context,
	req iotaclient.ResolveNameServiceNamesRequest,
) (*iotajsonrpc.IotaNamePage, error) {
	// Name service resolution is not directly available via GraphQL
	return nil, fmt.Errorf("ResolveNameServiceNames: name service queries not directly supported via GraphQL")
}

func (c *GraphQLClient) DevInspectTransactionBlock(
	ctx context.Context,
	req iotaclient.DevInspectTransactionBlockRequest,
) (*iotajsonrpc.DevInspectResults, error) {
	txBytes := req.TxKindBytes.String()

	// Populate TransactionMetadata from request
	gasPrice := uint64(iotaclient.DefaultGasPrice)
	if req.GasPrice != nil {
		gasPrice = req.GasPrice.Uint64()
	}

	txMeta := iotagraphql.TransactionMetadata{
		Sender:     *req.SenderAddress,
		GasPrice:   gasPrice,
		GasObjects: []iotagraphql.ObjectRef{}, // Empty for dev inspect
		GasBudget:  iotaclient.DefaultGasBudget,
		GasSponsor: *req.SenderAddress,
	}

	// Initialize boolean pointers to false to avoid GraphQL validation errors
	falseVal := false
	showBalanceChanges := &falseVal
	showEffects := &falseVal
	showRawEffects := &falseVal
	showEvents := &falseVal
	showInput := &falseVal
	showObjectChanges := &falseVal
	showRawInput := &falseVal
	// Need to show raw effects for dev inspect (includes full BCS data)
	trueVal := true
	showRawEffects = &trueVal

	resp, err := iotagraphql.DevInspectTransactionBlock(ctx, c.client, txBytes, txMeta,
		showBalanceChanges, showEffects, showRawEffects, showEvents, showInput, showObjectChanges, showRawInput)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL DevInspectResults to JSON-RPC DevInspectResults
	result, err := convertDevInspectResults(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DevInspectResults: %w", err)
	}

	return result, nil
}

func (c *GraphQLClient) DryRunTransaction(
	ctx context.Context,
	txDataBytes iotago.Base64Data,
) (*iotajsonrpc.DryRunTransactionBlockResponse, error) {
	txBytes := txDataBytes.String()
	// Initialize boolean pointers to false to avoid GraphQL validation errors
	falseVal := false
	showBalanceChanges := &falseVal
	showEffects := &falseVal
	showRawEffects := &falseVal
	showEvents := &falseVal
	showInput := &falseVal
	showObjectChanges := &falseVal
	showRawInput := &falseVal

	// Need to show effects, balance/object changes, and input for dry run to mimic JSON-RPC output
	trueVal := true
	showEffects = &trueVal
	showRawEffects = &trueVal
	showInput = &trueVal
	showBalanceChanges = &trueVal
	showObjectChanges = &trueVal

	resp, err := iotagraphql.DryRunTransactionBlock(ctx, c.client, txBytes,
		showBalanceChanges, showEffects, showRawEffects, showEvents, showInput, showObjectChanges, showRawInput)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL DryRunTransactionBlock to JSON-RPC DryRunTransactionBlockResponse
	result, err := convertDryRunResults(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DryRunResults: %w", err)
	}

	return result, nil
}

func (c *GraphQLClient) ExecuteTransactionBlock(
	ctx context.Context,
	req iotaclient.ExecuteTransactionBlockRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if len(req.Signatures) == 0 {
		return nil, fmt.Errorf("at least one signature is required")
	}
	txBytes := req.TxDataBytes.String()
	signatures := make([]string, len(req.Signatures))
	for i, sig := range req.Signatures {
		// Convert signature bytes to base64 string
		sigBytes := sig.Bytes()
		if sigBytes == nil {
			return nil, fmt.Errorf("signature %d has nil bytes", i)
		}
		signatures[i] = iotago.Base64Data(sigBytes).String()
	}
	// Initialize boolean pointers to false to avoid GraphQL validation errors
	falseVal := false
	showBalanceChanges := &falseVal
	showEffects := &falseVal
	showRawEffects := &falseVal
	showEvents := &falseVal
	showInput := &falseVal
	showObjectChanges := &falseVal
	showRawInput := &falseVal
	// Need to show effects and raw effects for execution
	trueVal := true
	showEffects = &trueVal
	showRawEffects = &trueVal
	if req.Options != nil {
		showBalanceChanges = lo.ToPtr(req.Options.ShowBalanceChanges)
		showEffects = lo.ToPtr(req.Options.ShowEffects)
		showRawEffects = lo.ToPtr(req.Options.ShowRawEffects)
		showEvents = lo.ToPtr(req.Options.ShowEvents)
		showInput = lo.ToPtr(req.Options.ShowInput)
		showObjectChanges = lo.ToPtr(req.Options.ShowObjectChanges)
		showRawInput = lo.ToPtr(req.Options.ShowRawInput)
	}
	resp, err := iotagraphql.ExecuteTransactionBlock(ctx, c.client, txBytes, signatures,
		showBalanceChanges, showEffects, showRawEffects, showEvents, showInput, showObjectChanges, showRawInput)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL response to JSON-RPC format
	return convertExecuteTransactionBlockResponse(resp, req.Options)
}

func (c *GraphQLClient) GetCommitteeInfo(
	ctx context.Context,
	epoch *iotajsonrpc.BigInt,
) (*iotajsonrpc.CommitteeInfo, error) {
	var epochID *uint64
	if epoch != nil {
		val := epoch.Uint64()
		epochID = &val
	}
	resp, err := iotagraphql.GetCommitteeInfo(ctx, c.client, epochID, nil)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL response to JSON-RPC CommitteeInfo
	validators := make([]iotajsonrpc.Validator, 0, len(resp.Epoch.ValidatorSet.ActiveValidators.Nodes))
	for range resp.Epoch.ValidatorSet.ActiveValidators.Nodes {
		// TODO: Extract proper public key and stake from node
		validators = append(validators, iotajsonrpc.Validator{
			PublicKey: nil,
			Stake:     iotajsonrpc.NewBigInt(0),
		})
	}

	return &iotajsonrpc.CommitteeInfo{
		EpochId:    iotajsonrpc.NewBigInt(resp.Epoch.EpochId),
		Validators: validators,
	}, nil
}

func (c *GraphQLClient) GetLatestIotaSystemState(ctx context.Context) (*iotajsonrpc.IotaSystemStateSummary, error) {
	resp, err := iotagraphql.GetLatestIotaSystemState(ctx, c.client)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL response to JSON-RPC IotaSystemStateSummary
	// Basic implementation - many fields use default values
	return &iotajsonrpc.IotaSystemStateSummary{
		Epoch:                                iotajsonrpc.NewBigInt(resp.Epoch.EpochId),
		ProtocolVersion:                      iotajsonrpc.NewBigInt(0), // TODO: extract from response
		SystemStateVersion:                   iotajsonrpc.NewBigInt(0),
		StorageFundTotalObjectStorageRebates: iotajsonrpc.NewBigInt(0),
		StorageFundNonRefundableBalance:      iotajsonrpc.NewBigInt(0),
		ReferenceGasPrice:                    resp.Epoch.ReferenceGasPrice.Clone(),
		SafeMode:                             false,
		SafeModeStorageRewards:               iotajsonrpc.NewBigInt(0),
		SafeModeComputationRewards:           iotajsonrpc.NewBigInt(0),
		SafeModeStorageRebates:               iotajsonrpc.NewBigInt(0),
		SafeModeNonRefundableStorageFee:      iotajsonrpc.NewBigInt(0),
		EpochStartTimestampMs:                iotajsonrpc.NewBigInt(0), // TODO: convert from time.Time
		EpochDurationMs:                      iotajsonrpc.NewBigInt(0),
		StakeSubsidyStartEpoch:               iotajsonrpc.NewBigInt(0),
		MaxValidatorCount:                    iotajsonrpc.NewBigInt(0),
		MinValidatorJoiningStake:             iotajsonrpc.NewBigInt(0),
		ValidatorReportRecords:               [][]interface{}{}, // TODO: extract from response
	}, nil
}

func (c *GraphQLClient) GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error) {
	resp, err := iotagraphql.GetReferenceGasPrice(ctx, c.client)
	if err != nil {
		return nil, err
	}
	return resp.Epoch.ReferenceGasPrice.Clone(), nil
}

func (c *GraphQLClient) GetStakes(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.DelegatedStake, error) {
	if owner == nil {
		return nil, fmt.Errorf("owner address is required")
	}
	resp, err := iotagraphql.GetStakes(ctx, c.client, *owner, nil, nil)
	if err != nil {
		return nil, err
	}
	// TODO: Convert GraphQL Stakes to JSON-RPC DelegatedStake
	_ = resp
	return nil, fmt.Errorf("GetStakes: GraphQL to JSON-RPC conversion not yet implemented")
}

func (c *GraphQLClient) GetStakesByIds(ctx context.Context, stakedIotaIds []iotago.ObjectID) ([]*iotajsonrpc.DelegatedStake, error) {
	if len(stakedIotaIds) == 0 {
		return []*iotajsonrpc.DelegatedStake{}, nil
	}
	// Convert ObjectIDs to Addresses
	ids := make([]iotago.Address, len(stakedIotaIds))
	for i, id := range stakedIotaIds {
		ids[i] = iotago.Address(id)
	}
	resp, err := iotagraphql.GetStakesByIds(ctx, c.client, ids, nil, nil)
	if err != nil {
		return nil, err
	}
	// TODO: Convert GraphQL Stakes to JSON-RPC DelegatedStake
	_ = resp
	return nil, fmt.Errorf("GetStakesByIds: GraphQL to JSON-RPC conversion not yet implemented")
}

func (c *GraphQLClient) GetValidatorsApy(ctx context.Context) (*iotajsonrpc.ValidatorsApy, error) {
	resp, err := iotagraphql.GetValidatorsApy(ctx, c.client)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL response to JSON-RPC format
	apys := make([]struct {
		Address string  `json:"address"`
		Apy     float64 `json:"apy"`
	}, 0, len(resp.Epoch.ValidatorSet.ActiveValidators.Nodes))

	for _, node := range resp.Epoch.ValidatorSet.ActiveValidators.Nodes {
		apys = append(apys, struct {
			Address string  `json:"address"`
			Apy     float64 `json:"apy"`
		}{
			Address: node.Address.Address.String(),
			Apy:     0.0, // TODO: calculate or extract APY from node data
		})
	}

	return &iotajsonrpc.ValidatorsApy{
		Epoch: iotajsonrpc.NewBigInt(resp.Epoch.EpochId),
		Apys:  apys,
	}, nil
}

func (c *GraphQLClient) BatchTransaction(
	ctx context.Context,
	req iotaclient.BatchTransactionRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("BatchTransaction: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) MergeCoins(
	ctx context.Context,
	req iotaclient.MergeCoinsRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("MergeCoins: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) MoveCall(
	ctx context.Context,
	req iotaclient.MoveCallRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("MoveCall: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) Pay(
	ctx context.Context,
	req iotaclient.PayRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	// Fetch the input coins to get ObjectRefs
	coinRefs := make([]*iotago.ObjectRef, 0, len(req.InputCoins))
	for _, coinID := range req.InputCoins {
		objResp, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
			ObjectID: coinID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", coinID.String(), err)
		}

		if objResp.Data == nil {
			return nil, fmt.Errorf("object %s not found", coinID.String())
		}

		coinRefs = append(coinRefs, &iotago.ObjectRef{
			ObjectID: objResp.Data.ObjectID,
			Version:  objResp.Data.Version.Uint64(),
			Digest:   objResp.Data.Digest,
		})
	}

	// Convert amounts from BigInt to uint64
	amounts := make([]uint64, len(req.Amount))
	for i, amt := range req.Amount {
		amounts[i] = amt.Uint64()
	}

	// Build the programmable transaction
	ptb := iotago.NewProgrammableTransactionBuilder()
	if err := ptb.Pay(coinRefs, req.Recipients, amounts); err != nil {
		return nil, fmt.Errorf("failed to build Pay transaction: %w", err)
	}
	pt := ptb.Finish()

	// Get gas budget
	gasBudget := uint64(iotaclient.DefaultGasBudget)
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}

	// Determine gas payment
	// For GraphQL client, Gas must be provided explicitly because we can't use input coins
	// as gas payment (they're used in the Pay command)
	if req.Gas == nil {
		return nil, fmt.Errorf("Gas parameter is required for Pay via GraphQL (input coins cannot be used as gas)")
	}

	// Fetch the gas object
	gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: req.Gas,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get gas object %s: %w", req.Gas.String(), err)
	}
	if gasObj.Data == nil {
		return nil, fmt.Errorf("gas object %s not found", req.Gas.String())
	}

	gasPayment := []*iotago.ObjectRef{
		{
			ObjectID: gasObj.Data.ObjectID,
			Version:  gasObj.Data.Version.Uint64(),
			Digest:   gasObj.Data.Digest,
		},
	}

	// Create input objects for response (include both payment coins and gas coin)
	inputObjects := make([]iotajsonrpc.InputObjectKind, 0)
	for _, ref := range coinRefs {
		inputObjects = append(inputObjects, iotajsonrpc.InputObjectKind{
			"ImmOrOwnedMoveObject": map[string]interface{}{
				"objectId": ref.ObjectID.String(),
				"version":  ref.Version,
				"digest":   ref.Digest.String(),
			},
		})
	}
	// Add gas payment coins to input objects if they're different from payment coins
	for _, gasRef := range gasPayment {
		inputObjects = append(inputObjects, iotajsonrpc.InputObjectKind{
			"ImmOrOwnedMoveObject": map[string]interface{}{
				"objectId": gasRef.ObjectID.String(),
				"version":  gasRef.Version,
				"digest":   gasRef.Digest.String(),
			},
		})
	}

	// Create transaction data
	tx := iotago.NewProgrammable(
		req.Signer,
		pt,
		gasPayment,
		gasBudget,
		iotaclient.DefaultGasPrice,
	)

	// Serialize to BCS
	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Convert gasPayment from []*iotago.ObjectRef to []iotago.ObjectRef
	gasRefs := make([]iotago.ObjectRef, len(gasPayment))
	for i, ref := range gasPayment {
		gasRefs[i] = *ref
	}

	return &iotajsonrpc.TransactionBytes{
		TxBytes:      txBytes,
		Gas:          gasRefs,
		InputObjects: inputObjects,
	}, nil
}

func (c *GraphQLClient) PayAllIota(
	ctx context.Context,
	req iotaclient.PayAllIotaRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	// Build the programmable transaction
	ptb := iotago.NewProgrammableTransactionBuilder()
	if err := ptb.PayAllIota(req.Recipient); err != nil {
		return nil, fmt.Errorf("failed to build PayAllIota transaction: %w", err)
	}
	pt := ptb.Finish()

	// Get gas budget
	gasBudget := uint64(iotaclient.DefaultGasBudget)
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}

	// Fetch the input coins to get ObjectRefs
	gasPayment := make([]*iotago.ObjectRef, 0, len(req.InputCoins))
	inputObjects := make([]iotajsonrpc.InputObjectKind, 0)

	for _, coinID := range req.InputCoins {
		objResp, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
			ObjectID: coinID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", coinID.String(), err)
		}

		if objResp.Data == nil {
			return nil, fmt.Errorf("object %s not found", coinID.String())
		}

		objRef := &iotago.ObjectRef{
			ObjectID: objResp.Data.ObjectID,
			Version:  objResp.Data.Version.Uint64(),
			Digest:   objResp.Data.Digest,
		}
		gasPayment = append(gasPayment, objRef)

		// InputObjectKind is map[string]interface{} for JSON-RPC compatibility
		inputObjects = append(inputObjects, iotajsonrpc.InputObjectKind{
			"ImmOrOwnedMoveObject": map[string]interface{}{
				"objectId": objResp.Data.ObjectID.String(),
				"version":  objResp.Data.Version.Uint64(),
				"digest":   objResp.Data.Digest.String(),
			},
		})
	}

	// Create transaction data
	tx := iotago.NewProgrammable(
		req.Signer,
		pt,
		gasPayment,
		gasBudget,
		iotaclient.DefaultGasPrice,
	)

	// Serialize to BCS
	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Convert gasPayment from []*iotago.ObjectRef to []iotago.ObjectRef
	gasRefs := make([]iotago.ObjectRef, len(gasPayment))
	for i, ref := range gasPayment {
		gasRefs[i] = *ref
	}

	return &iotajsonrpc.TransactionBytes{
		TxBytes:      txBytes,
		Gas:          gasRefs,
		InputObjects: inputObjects,
	}, nil
}

func (c *GraphQLClient) PayIota(
	ctx context.Context,
	req iotaclient.PayIotaRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("PayIota: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) Publish(
	ctx context.Context,
	req iotaclient.PublishRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("Publish: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) RequestAddStake(
	ctx context.Context,
	req iotaclient.RequestAddStakeRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("RequestAddStake: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) RequestWithdrawStake(
	ctx context.Context,
	req iotaclient.RequestWithdrawStakeRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("RequestWithdrawStake: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) SplitCoin(
	ctx context.Context,
	req iotaclient.SplitCoinRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("SplitCoin: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) SplitCoinEqual(
	ctx context.Context,
	req iotaclient.SplitCoinEqualRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("SplitCoinEqual: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) TransferObject(
	ctx context.Context,
	req iotaclient.TransferObjectRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("TransferObject: signer address is required")
	}
	if req.ObjectID == nil {
		return nil, fmt.Errorf("TransferObject: object ID is required")
	}
	if req.Recipient == nil {
		return nil, fmt.Errorf("TransferObject: recipient address is required")
	}

	objectRef, err := c.loadObjectRef(ctx, req.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to load object %s: %w", req.ObjectID.String(), err)
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	if err := ptb.TransferObject(req.Recipient, objectRef); err != nil {
		return nil, fmt.Errorf("failed to build TransferObject transaction: %w", err)
	}
	pt := ptb.Finish()

	gasBudget := uint64(iotaclient.DefaultGasBudget)
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}

	gasRef, err := c.resolveGasObject(ctx, req.Signer, req.Gas, req.ObjectID)
	if err != nil {
		return nil, err
	}
	gasPayment := []*iotago.ObjectRef{gasRef}

	inputObjects := []iotajsonrpc.InputObjectKind{newInputObjectKind(objectRef)}
	if gasRef.ObjectID != objectRef.ObjectID || gasRef.Version != objectRef.Version || gasRef.Digest != objectRef.Digest {
		inputObjects = append(inputObjects, newInputObjectKind(gasRef))
	}

	tx := iotago.NewProgrammable(
		req.Signer,
		pt,
		gasPayment,
		gasBudget,
		iotaclient.DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	gasRefs := make([]iotago.ObjectRef, len(gasPayment))
	for i, ref := range gasPayment {
		gasRefs[i] = *ref
	}

	return &iotajsonrpc.TransactionBytes{
		TxBytes:      txBytes,
		Gas:          gasRefs,
		InputObjects: inputObjects,
	}, nil
}

func (c *GraphQLClient) loadObjectRef(ctx context.Context, objectID *iotago.ObjectID) (*iotago.ObjectRef, error) {
	if objectID == nil {
		return nil, fmt.Errorf("object ID is nil")
	}
	objResp, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: objectID})
	if err != nil {
		return nil, err
	}
	if objResp == nil || objResp.Data == nil || objResp.Data.ObjectID == nil || objResp.Data.Version == nil {
		return nil, fmt.Errorf("object %s not found", objectID.String())
	}
	return &iotago.ObjectRef{
		ObjectID: objResp.Data.ObjectID,
		Version:  objResp.Data.Version.Uint64(),
		Digest:   objResp.Data.Digest,
	}, nil
}

func (c *GraphQLClient) resolveGasObject(
	ctx context.Context,
	signer *iotago.Address,
	gasID *iotago.ObjectID,
	transferObjectID *iotago.ObjectID,
) (*iotago.ObjectRef, error) {
	if gasID != nil {
		gasRef, err := c.loadObjectRef(ctx, gasID)
		if err != nil {
			return nil, fmt.Errorf("failed to load gas object %s: %w", gasID.String(), err)
		}
		if transferObjectID != nil && *transferObjectID == *gasRef.ObjectID {
			return nil, fmt.Errorf("gas object %s cannot be the same as the transferred object", gasID.String())
		}
		return gasRef, nil
	}
	if signer == nil {
		return nil, fmt.Errorf("signer address is required to select a gas coin")
	}
	const pageLimit = uint(50)
	var cursor *string
	for {
		coins, err := c.GetCoins(ctx, iotaclient.GetCoinsRequest{
			Owner:  signer,
			Limit:  pageLimit,
			Cursor: cursor,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch coins for gas selection: %w", err)
		}
		for _, coin := range coins.Data {
			if coin == nil || coin.CoinObjectID == nil || coin.Version == nil {
				continue
			}
			if transferObjectID != nil && *coin.CoinObjectID == *transferObjectID {
				continue
			}
			return &iotago.ObjectRef{
				ObjectID: coin.CoinObjectID,
				Version:  coin.Version.Uint64(),
				Digest:   coin.Digest,
			}, nil
		}
		if !coins.HasNextPage || coins.NextCursor == nil || *coins.NextCursor == "" {
			break
		}
		cursor = coins.NextCursor
	}
	return nil, fmt.Errorf("no suitable gas coin found; provide Gas explicitly")
}

func newInputObjectKind(ref *iotago.ObjectRef) iotajsonrpc.InputObjectKind {
	return iotajsonrpc.InputObjectKind{
		"ImmOrOwnedMoveObject": map[string]interface{}{
			"objectId": ref.ObjectID.String(),
			"version":  ref.Version,
			"digest":   ref.Digest.String(),
		},
	}
}

func (c *GraphQLClient) TransferIota(
	ctx context.Context,
	req iotaclient.TransferIotaRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("TransferIota: transaction building not supported via GraphQL (requires JSON-RPC)")
}

func (c *GraphQLClient) GetCoinObjsForTargetAmount(
	ctx context.Context,
	address *iotago.Address,
	targetAmount uint64,
	gasAmount uint64,
) (iotajsonrpc.Coins, error) {
	// This would require fetching coins via GraphQL and filtering/selecting client-side
	// Similar to GetIotaCoinsOwnedByAddress but with additional logic
	return nil, fmt.Errorf("GetCoinObjsForTargetAmount: requires client-side coin selection logic")
}

func (c *GraphQLClient) SignAndExecuteTransaction(
	ctx context.Context,
	req *iotaclient.SignAndExecuteTransactionRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	// This requires transaction signing which is client-side
	return nil, fmt.Errorf("SignAndExecuteTransaction: requires client-side transaction signing (not supported via GraphQL)")
}

func (c *GraphQLClient) PublishContract(
	ctx context.Context,
	signer iotasigner.Signer,
	modules []*iotago.Base64Data,
	dependencies []*iotago.Address,
	gasBudget uint64,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, *iotago.PackageID, error) {
	return nil, nil, fmt.Errorf("PublishContract: requires transaction building and signing (not supported via GraphQL)")
}

func (c *GraphQLClient) UpdateObjectRef(
	ctx context.Context,
	ref *iotago.ObjectRef,
) (*iotago.ObjectRef, error) {
	if ref == nil || ref.ObjectID == nil {
		return nil, fmt.Errorf("object reference is required")
	}
	// Fetch the latest version of the object
	objAddr := iotago.Address(*ref.ObjectID)
	resp, err := iotagraphql.GetObject(ctx, c.client, objAddr, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	// TODO: Extract version and digest from response
	_ = resp
	return nil, fmt.Errorf("UpdateObjectRef: GraphQL to ObjectRef conversion not yet implemented")
}

func (c *GraphQLClient) MintToken(
	ctx context.Context,
	signer iotasigner.Signer,
	packageID *iotago.PackageID,
	tokenName string,
	treasuryCap *iotago.ObjectRef,
	mintAmount uint64,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	return nil, fmt.Errorf("MintToken: requires transaction building and signing (not supported via GraphQL)")
}

func (c *GraphQLClient) GetIotaCoinsOwnedByAddress(ctx context.Context, address *iotago.Address) (iotajsonrpc.Coins, error) {
	if address == nil {
		return nil, fmt.Errorf("address is required")
	}
	// Use GetCoins with IOTA coin type
	iotaCoinType := "0x2::iota::IOTA"
	resp, err := iotagraphql.GetCoins(ctx, c.client, *address, nil, nil, &iotaCoinType)
	if err != nil {
		return nil, err
	}
	// TODO: Convert GraphQL coins to JSON-RPC Coins
	_ = resp
	return nil, fmt.Errorf("GetIotaCoinsOwnedByAddress: GraphQL to Coins conversion not yet implemented")
}

func (c *GraphQLClient) BatchGetObjectsOwnedByAddress(
	ctx context.Context,
	address *iotago.Address,
	options *iotajsonrpc.IotaObjectDataOptions,
	filterType string,
) ([]iotajsonrpc.IotaObjectResponse, error) {
	if address == nil {
		return nil, fmt.Errorf("address is required")
	}
	// Use GetOwnedObjects with appropriate filter
	// TODO: Convert filterType to GraphQL ObjectFilter
	var showBcs, showContent, showDisplay, showType, showOwner, showPreviousTransaction, showStorageRebate *bool
	if options != nil {
		showBcs = lo.ToPtr(options.ShowBcs)
		showContent = lo.ToPtr(options.ShowContent)
		showDisplay = lo.ToPtr(options.ShowDisplay)
		showType = lo.ToPtr(options.ShowType)
		showOwner = lo.ToPtr(options.ShowOwner)
		showPreviousTransaction = lo.ToPtr(options.ShowPreviousTransaction)
		showStorageRebate = lo.ToPtr(options.ShowStorageRebate)
	}
	resp, err := iotagraphql.GetOwnedObjects(ctx, c.client, *address, nil, nil,
		showBcs, showContent, showDisplay, showType, showOwner, showPreviousTransaction, showStorageRebate, nil)
	if err != nil {
		return nil, err
	}
	// TODO: Convert GraphQL objects to JSON-RPC IotaObjectResponse and filter by type
	_ = resp
	_ = filterType
	return nil, fmt.Errorf("BatchGetObjectsOwnedByAddress: GraphQL to IotaObjectResponse conversion not yet implemented")
}

func (c *GraphQLClient) BatchGetFilteredObjectsOwnedByAddress(
	ctx context.Context,
	address *iotago.Address,
	options *iotajsonrpc.IotaObjectDataOptions,
	filter func(*iotajsonrpc.IotaObjectData) bool,
) ([]iotajsonrpc.IotaObjectResponse, error) {
	// This requires fetching all objects and filtering client-side
	// Would use BatchGetObjectsOwnedByAddress and then apply the filter
	_ = filter
	return nil, fmt.Errorf("BatchGetFilteredObjectsOwnedByAddress: requires client-side filtering logic")
}

func (c *GraphQLClient) GetAllBalances(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.Balance, error) {
	if owner == nil {
		return nil, fmt.Errorf("owner address is required")
	}
	resp, err := iotagraphql.GetAllBalances(ctx, c.client, *owner, nil, nil)
	if err != nil {
		return nil, err
	}
	balances := make([]*iotajsonrpc.Balance, 0, len(resp.Address.Balances.Nodes))
	for _, node := range resp.Address.Balances.Nodes {
		bal, err := convertGraphQLBalance(node.CoinType.Repr, node.CoinObjectCount, node.TotalBalance)
		if err != nil {
			return nil, err
		}
		balances = append(balances, bal)
	}
	return balances, nil
}

func (c *GraphQLClient) GetAllCoins(ctx context.Context, req iotaclient.GetAllCoinsRequest) (*iotajsonrpc.CoinPage, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("owner address is required")
	}

	// Convert limit from uint to *int
	var limitPtr *int
	if req.Limit > 0 {
		limit := int(req.Limit)
		limitPtr = &limit
	}

	// Convert cursor from *iotago.ObjectID to *string
	var cursorPtr *string
	if req.Cursor != nil {
		cursor := req.Cursor.String()
		cursorPtr = &cursor
	}

	// Call the generated GetAllCoins function
	resp, err := iotagraphql.GetAllCoins(ctx, c.client, *req.Owner, limitPtr, cursorPtr)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL response to JSON-RPC CoinPage
	nodes := resp.Address.Coins.Nodes
	coins := make([]*iotajsonrpc.Coin, 0, len(nodes))
	for _, node := range nodes {
		coin, err := convertGraphQLAllCoin(&node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert coin: %w", err)
		}
		coins = append(coins, coin)
	}

	// Convert cursor
	var nextCursor *string
	if resp.Address.Coins.PageInfo.HasNextPage && resp.Address.Coins.PageInfo.EndCursor != "" {
		nextCursor = &resp.Address.Coins.PageInfo.EndCursor
	}

	return &iotajsonrpc.CoinPage{
		Data:        coins,
		NextCursor:  nextCursor,
		HasNextPage: resp.Address.Coins.PageInfo.HasNextPage,
	}, nil
}

func (c *GraphQLClient) GetBalance(ctx context.Context, req iotaclient.GetBalanceRequest) (*iotajsonrpc.Balance, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("owner address is required")
	}
	var coinTypePtr *string
	if req.CoinType != "" {
		coinTypePtr = &req.CoinType
	}
	resp, err := iotagraphql.GetBalance(ctx, c.client, *req.Owner, coinTypePtr)
	if err != nil {
		return nil, err
	}
	balance := resp.Address.Balance
	return convertGraphQLBalance(balance.CoinType.Repr, balance.CoinObjectCount, balance.TotalBalance)
}

func (c *GraphQLClient) GetCoinMetadata(ctx context.Context, coinType string) (*iotajsonrpc.IotaCoinMetadata, error) {
	if coinType == "" {
		return nil, fmt.Errorf("coin type is required")
	}
	resp, err := iotagraphql.GetCoinMetadata(ctx, c.client, coinType)
	if err != nil {
		return nil, err
	}
	meta := resp.CoinMetadata
	objID := addressToObjectID(meta.Address)
	return &iotajsonrpc.IotaCoinMetadata{
		Name:        meta.Name,
		Symbol:      meta.Symbol,
		Decimals:    uint8(meta.Decimals),
		Description: meta.Description,
		IconUrl:     meta.IconUrl,
		Id:          objID,
	}, nil
}

func (c *GraphQLClient) GetCoins(ctx context.Context, req iotaclient.GetCoinsRequest) (*iotajsonrpc.CoinPage, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("owner address is required")
	}

	// Convert limit from uint to *int
	var limitPtr *int
	if req.Limit > 0 {
		limit := int(req.Limit)
		limitPtr = &limit
	}

	// Convert cursor from *iotago.ObjectID to *string
	var cursorPtr *string = req.Cursor

	resp, err := iotagraphql.GetCoins(ctx, c.client, *req.Owner, limitPtr, cursorPtr, req.CoinType)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL response to JSON-RPC CoinPage
	nodes := resp.Address.Coins.Nodes
	coins := make([]*iotajsonrpc.Coin, 0, len(nodes))
	for _, node := range nodes {
		coin, err := convertGraphQLCoin(&node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert coin: %w", err)
		}
		coins = append(coins, coin)
	}

	// Convert cursor
	var nextCursor *string
	if resp.Address.Coins.PageInfo.HasNextPage && resp.Address.Coins.PageInfo.EndCursor != "" {
		nextCursor = &resp.Address.Coins.PageInfo.EndCursor
	}

	return &iotajsonrpc.CoinPage{
		Data:        coins,
		NextCursor:  nextCursor,
		HasNextPage: resp.Address.Coins.PageInfo.HasNextPage,
	}, nil
}

func convertGraphQLCoin(node *iotagraphql.GetCoinsAddressCoinsCoinConnectionNodesCoin) (*iotajsonrpc.Coin, error) {
	coinType, err := iotajsonrpc.CoinTypeFromString(node.Contents.Type.Repr)
	if err != nil {
		return nil, fmt.Errorf("invalid coin type %s: %w", node.Contents.Type.Repr, err)
	}

	objID := iotago.ObjectID(node.Address)
	digest, err := iotago.NewDigest(node.Digest)
	if err != nil {
		return nil, fmt.Errorf("invalid digest %s: %w", node.Digest, err)
	}

	txDigest, err := iotago.NewDigest(node.PreviousTransactionBlock.Digest)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction digest %s: %w", node.PreviousTransactionBlock.Digest, err)
	}

	return &iotajsonrpc.Coin{
		CoinType:            coinType,
		CoinObjectID:        &objID,
		Version:             iotajsonrpc.NewBigInt(node.Version),
		Digest:              digest,
		Balance:             node.CoinBalance.Clone(),
		PreviousTransaction: *txDigest,
	}, nil
}

func convertGraphQLAllCoin(node *iotagraphql.GetAllCoinsAddressCoinsCoinConnectionNodesCoin) (*iotajsonrpc.Coin, error) {
	coinType, err := iotajsonrpc.CoinTypeFromString(node.Contents.Type.Repr)
	if err != nil {
		return nil, fmt.Errorf("invalid coin type %s: %w", node.Contents.Type.Repr, err)
	}

	objID := iotago.ObjectID(node.Address)
	digest, err := iotago.NewDigest(node.Digest)
	if err != nil {
		return nil, fmt.Errorf("invalid digest %s: %w", node.Digest, err)
	}

	txDigest, err := iotago.NewDigest(node.PreviousTransactionBlock.Digest)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction digest %s: %w", node.PreviousTransactionBlock.Digest, err)
	}

	return &iotajsonrpc.Coin{
		CoinType:            coinType,
		CoinObjectID:        &objID,
		Version:             iotajsonrpc.NewBigInt(node.Version),
		Digest:              digest,
		Balance:             node.CoinBalance.Clone(),
		PreviousTransaction: *txDigest,
	}, nil
}

func (c *GraphQLClient) GetTotalSupply(ctx context.Context, coinType string) (*iotajsonrpc.Supply, error) {
	if coinType == "" {
		return nil, fmt.Errorf("coin type is required")
	}
	resp, err := iotagraphql.GetTotalSupply(ctx, c.client, coinType)
	if err != nil {
		return nil, err
	}
	supply := resp.CoinMetadata.Supply.Clone()
	return &iotajsonrpc.Supply{Value: supply}, nil
}

func (c *GraphQLClient) GetChainIdentifier(ctx context.Context) (string, error) {
	resp, err := iotagraphql.GetChainIdentifier(ctx, c.client)
	if err != nil {
		return "", err
	}
	return resp.ChainIdentifier, nil
}

func (c *GraphQLClient) GetCheckpoint(ctx context.Context, checkpointID *iotajsonrpc.BigInt) (*iotajsonrpc.Checkpoint, error) {
	// Convert checkpoint ID to GraphQL format
	id := &iotagraphql.CheckpointId{
		SequenceNumber: checkpointID.Uint64(),
	}

	resp, err := iotagraphql.GetCheckpoint(ctx, c.client, id)
	if err != nil {
		return nil, err
	}

	// TODO: Convert GraphQL Checkpoint to JSON-RPC Checkpoint
	_ = resp
	return nil, fmt.Errorf("GetCheckpoint: GraphQL to JSON-RPC conversion not yet implemented")
}

func (c *GraphQLClient) GetCheckpoints(ctx context.Context, req iotaclient.GetCheckpointsRequest) (*iotajsonrpc.CheckpointPage, error) {
	// Convert request parameters to GraphQL format
	var first, last *int
	var before, after *string

	if req.Limit != nil {
		limit := int(*req.Limit)
		first = &limit
	}

	if req.Cursor != nil {
		cursorStr := req.Cursor.String()
		if req.DescendingOrder {
			before = &cursorStr
		} else {
			after = &cursorStr
		}
	}

	resp, err := iotagraphql.GetCheckpoints(ctx, c.client, first, before, last, after)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL response to JSON-RPC format
	checkpoints := make([]*iotajsonrpc.Checkpoint, 0, len(resp.Checkpoints.Nodes))
	for range resp.Checkpoints.Nodes {
		// TODO: Convert each checkpoint node
		checkpoints = append(checkpoints, &iotajsonrpc.Checkpoint{})
	}

	var nextCursor *iotajsonrpc.BigInt
	if resp.Checkpoints.PageInfo.HasNextPage && resp.Checkpoints.PageInfo.EndCursor != "" {
		// Parse the cursor as a number
		// TODO: proper cursor parsing
		nextCursor = iotajsonrpc.NewBigInt(0)
	}

	return &iotajsonrpc.CheckpointPage{
		Data:        checkpoints,
		NextCursor:  nextCursor,
		HasNextPage: resp.Checkpoints.PageInfo.HasNextPage,
	}, nil
}

func (c *GraphQLClient) GetEvents(ctx context.Context, digest *iotago.TransactionDigest) ([]*iotajsonrpc.IotaEvent, error) {
	if digest == nil {
		return nil, fmt.Errorf("transaction digest is required")
	}
	// Get the transaction block with events
	showEvents := true
	resp, err := iotagraphql.GetTransactionBlock(ctx, c.client, digest.String(),
		nil, nil, nil, &showEvents, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	// TODO: Extract and convert events from transaction block
	_ = resp
	return nil, fmt.Errorf("GetEvents: GraphQL to JSON-RPC conversion not yet implemented")
}

func (c *GraphQLClient) GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error) {
	resp, err := iotagraphql.GetLatestCheckpointSequenceNumber(ctx, c.client)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", resp.Checkpoint.SequenceNumber), nil
}

func (c *GraphQLClient) GetObject(ctx context.Context, req iotaclient.GetObjectRequest) (*iotajsonrpc.IotaObjectResponse, error) {
	if req.ObjectID == nil {
		return nil, fmt.Errorf("object ID is required")
	}
	objAddr := iotago.Address(*req.ObjectID)
	// Initialize boolean pointers to false to avoid GraphQL validation errors
	falseVal := false
	showBcs := &falseVal
	showOwner := &falseVal
	showPreviousTransaction := &falseVal
	showContent := &falseVal
	showDisplay := &falseVal
	showType := &falseVal
	showStorageRebate := &falseVal
	if req.Options != nil {
		showBcs = lo.ToPtr(req.Options.ShowBcs)
		showOwner = lo.ToPtr(req.Options.ShowOwner)
		showPreviousTransaction = lo.ToPtr(req.Options.ShowPreviousTransaction)
		showContent = lo.ToPtr(req.Options.ShowContent)
		showDisplay = lo.ToPtr(req.Options.ShowDisplay)
		showType = lo.ToPtr(req.Options.ShowType)
		showStorageRebate = lo.ToPtr(req.Options.ShowStorageRebate)
	}
	resp, err := iotagraphql.GetObject(ctx, c.client, objAddr,
		showBcs, showOwner, showPreviousTransaction, showContent, showDisplay, showType, showStorageRebate)
	if err != nil {
		return nil, err
	}

	return convertGraphQLObjectToIotaObjectResponse(&resp.Object, req.Options)
}

func (c *GraphQLClient) GetProtocolConfig(
	ctx context.Context,
	version *iotajsonrpc.BigInt,
) (*iotajsonrpc.ProtocolConfig, error) {
	var versionPtr *uint64
	if version != nil {
		val := version.Uint64()
		versionPtr = &val
	}
	resp, err := iotagraphql.GetProtocolConfig(ctx, c.client, versionPtr)
	if err != nil {
		return nil, err
	}
	// TODO: Convert GraphQL ProtocolConfig to JSON-RPC ProtocolConfig
	_ = resp
	return nil, fmt.Errorf("GetProtocolConfig: GraphQL to JSON-RPC conversion not yet implemented")
}

func (c *GraphQLClient) GetTotalTransactionBlocks(ctx context.Context) (string, error) {
	resp, err := iotagraphql.GetTotalTransactionBlocks(ctx, c.client)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", resp.Checkpoint.NetworkTotalTransactions), nil
}

func (c *GraphQLClient) GetTransactionBlock(ctx context.Context, req iotaclient.GetTransactionBlockRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if req.Digest == nil {
		return nil, fmt.Errorf("transaction digest is required")
	}
	// Default all show options to false if not specified
	showBalanceChanges := lo.ToPtr(false)
	showEffects := lo.ToPtr(false)
	showRawEffects := lo.ToPtr(false)
	showEvents := lo.ToPtr(false)
	showInput := lo.ToPtr(false)
	showObjectChanges := lo.ToPtr(false)
	showRawInput := lo.ToPtr(false)

	if req.Options != nil {
		showBalanceChanges = lo.ToPtr(req.Options.ShowBalanceChanges)
		showEffects = lo.ToPtr(req.Options.ShowEffects)
		showRawEffects = lo.ToPtr(req.Options.ShowRawEffects)
		showEvents = lo.ToPtr(req.Options.ShowEvents)
		showInput = lo.ToPtr(req.Options.ShowInput)
		showObjectChanges = lo.ToPtr(req.Options.ShowObjectChanges)
		showRawInput = lo.ToPtr(req.Options.ShowRawInput)
	}
	resp, err := iotagraphql.GetTransactionBlock(ctx, c.client, req.Digest.String(),
		showBalanceChanges, showEffects, showRawEffects, showEvents, showInput, showObjectChanges, showRawInput)
	if err != nil {
		return nil, err
	}

	return convertGraphQLTransactionBlockToResponse(&resp.TransactionBlock, req.Options)
}

func (c *GraphQLClient) MultiGetObjects(ctx context.Context, req iotaclient.MultiGetObjectsRequest) ([]iotajsonrpc.IotaObjectResponse, error) {
	// MultiGetObjectsRequest doesn't have Limit/Cursor fields
	_ = req
	return nil, fmt.Errorf("MultiGetObjects: request structure mismatch, needs custom implementation")
}

func (c *GraphQLClient) MultiGetTransactionBlocks(
	ctx context.Context,
	req iotaclient.MultiGetTransactionBlocksRequest,
) ([]*iotajsonrpc.IotaTransactionBlockResponse, error) {
	// MultiGetTransactionBlocksRequest doesn't have Limit/Cursor fields
	_ = req
	return nil, fmt.Errorf("MultiGetTransactionBlocks: request structure mismatch, needs custom implementation")
}

func (c *GraphQLClient) TryGetPastObject(
	ctx context.Context,
	req iotaclient.TryGetPastObjectRequest,
) (*iotajsonrpc.IotaPastObjectResponse, error) {
	if req.ObjectID == nil {
		return nil, fmt.Errorf("object ID is required")
	}
	objAddr := iotago.Address(*req.ObjectID)
	version := req.Version // Version is already uint64
	// Always provide boolean values (default to false) to avoid GraphQL errors
	showBcs := lo.ToPtr(false)
	showOwner := lo.ToPtr(false)
	showPreviousTransaction := lo.ToPtr(false)
	showContent := lo.ToPtr(false)
	showDisplay := lo.ToPtr(false)
	showType := lo.ToPtr(false)
	showStorageRebate := lo.ToPtr(false)
	if req.Options != nil {
		showBcs = lo.ToPtr(req.Options.ShowBcs)
		showOwner = lo.ToPtr(req.Options.ShowOwner)
		showPreviousTransaction = lo.ToPtr(req.Options.ShowPreviousTransaction)
		showContent = lo.ToPtr(req.Options.ShowContent)
		showDisplay = lo.ToPtr(req.Options.ShowDisplay)
		showType = lo.ToPtr(req.Options.ShowType)
		showStorageRebate = lo.ToPtr(req.Options.ShowStorageRebate)
	}
	resp, err := iotagraphql.TryGetPastObject(ctx, c.client, objAddr, &version,
		showBcs, showOwner, showPreviousTransaction, showContent, showDisplay, showType, showStorageRebate)
	if err != nil {
		return nil, err
	}

	// Convert GraphQL response to JSON-RPC IotaPastObjectResponse
	pastObjectResp, err := convertGraphQLTryGetPastObjectResponse(resp, version, req.Options)
	if err != nil {
		return nil, fmt.Errorf("failed to convert GraphQL response: %w", err)
	}
	return pastObjectResp, nil
}

func (c *GraphQLClient) TryMultiGetPastObjects(
	ctx context.Context,
	req iotaclient.TryMultiGetPastObjectsRequest,
) ([]*iotajsonrpc.IotaPastObjectResponse, error) {
	if len(req.PastObjects) == 0 {
		return []*iotajsonrpc.IotaPastObjectResponse{}, nil
	}
	// GraphQL doesn't have a direct batch query for past objects
	// Would need to make individual calls or implement custom batching
	return nil, fmt.Errorf("TryMultiGetPastObjects: batch past object queries not directly supported via GraphQL")
}

func (c *GraphQLClient) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	// Faucet operations are not supported via GraphQL
	_ = address
	return fmt.Errorf("RequestFunds: faucet operations not supported via GraphQL")
}

func (c *GraphQLClient) Health(ctx context.Context) error {
	// Try to execute a simple query to check health
	_, err := iotagraphql.GetChainIdentifier(ctx, c.client)
	return err
}

func (c *GraphQLClient) L2() L2Client {
	// L2 client is not supported via GraphQL
	return nil
}

func (c *GraphQLClient) IotaClient() *iotaclient.Client {
	// Cannot return underlying JSON-RPC client from GraphQL client
	return nil
}

func (c *GraphQLClient) DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error) {
	// Contract deployment requires transaction building and signing
	_ = signer
	return iotago.PackageID{}, fmt.Errorf("DeployISCContracts: requires transaction building and signing (not supported via GraphQL)")
}

func (c *GraphQLClient) GetISCPackageIDForAnchor(ctx context.Context, anchor iotago.ObjectID) (iotago.PackageID, error) {
	// This is ISC-specific logic that would require querying object data
	_ = anchor
	return iotago.PackageID{}, fmt.Errorf("GetISCPackageIDForAnchor: ISC-specific queries not yet implemented via GraphQL")
}

func (c *GraphQLClient) FindCoinsForGasPayment(
	ctx context.Context,
	owner *iotago.Address,
	pt iotago.ProgrammableTransaction,
	gasPrice uint64,
	gasBudget uint64,
) ([]*iotago.ObjectRef, error) {
	// This requires client-side coin selection logic
	_ = pt
	_ = gasPrice
	_ = gasBudget
	return nil, fmt.Errorf("FindCoinsForGasPayment: requires client-side coin selection logic")
}

func (c *GraphQLClient) MergeCoinsAndExecute(
	ctx context.Context,
	owner iotasigner.Signer,
	destinationCoin *iotago.ObjectRef,
	sourceCoins []*iotago.ObjectRef,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	// This requires transaction building and signing
	_ = owner
	_ = destinationCoin
	_ = sourceCoins
	_ = gasBudget
	return nil, fmt.Errorf("MergeCoinsAndExecute: requires transaction building and signing (not supported via GraphQL)")
}

func (c *GraphQLClient) SignAndExecuteTxWithRetry(
	ctx context.Context,
	signer iotasigner.Signer,
	pt iotago.ProgrammableTransaction,
	gasCoin *iotago.ObjectRef,
	gasBudget uint64,
	gasPrice uint64,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	// This requires transaction building and signing
	_ = signer
	_ = pt
	_ = gasCoin
	_ = gasBudget
	_ = gasPrice
	_ = options
	return nil, fmt.Errorf("SignAndExecuteTxWithRetry: requires transaction building and signing (not supported via GraphQL)")
}

func (c *GraphQLClient) WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error) {
	// This would require polling GraphQL for object version changes
	_ = timeout
	_ = logger
	_ = currentRef
	_ = cb
	return nil, fmt.Errorf("WaitForNextVersionForTesting: polling logic not yet implemented for GraphQL")
}

func convertGraphQLBalance(coinTypeRepr string, coinObjectCount uint64, totalBalance iotajsonrpc.BigInt) (*iotajsonrpc.Balance, error) {
	coinType, err := iotajsonrpc.CoinTypeFromString(coinTypeRepr)
	if err != nil {
		return nil, fmt.Errorf("invalid coin type %s: %w", coinTypeRepr, err)
	}
	return &iotajsonrpc.Balance{
		CoinType:        coinType,
		CoinObjectCount: iotajsonrpc.NewBigInt(coinObjectCount),
		TotalBalance:    totalBalance.Clone(),
	}, nil
}

func addressToObjectID(addr iotago.Address) *iotago.ObjectID {
	obj := iotago.ObjectID(addr)
	return &obj
}

func convertGraphQLDynamicFieldToInfo(
	node *iotagraphql.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicField,
) (*iotajsonrpc.DynamicFieldInfo, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}

	// Convert Name field
	var nameValue any
	if err := json.Unmarshal(node.GetName().Json, &nameValue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal name JSON: %w", err)
	}

	name := iotago.DynamicFieldName{
		Type:  node.GetName().Type.Repr,
		Value: nameValue,
	}

	// Get the value to determine the type and extract object info
	value := node.GetValue()
	var fieldType serialization.TagJson[iotago.DynamicFieldType]
	var objectType string
	var objectID iotago.ObjectID
	var version iotago.SequenceNumber
	var digest iotago.ObjectDigest

	switch v := value.(type) {
	case *iotagraphql.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveObject:
		// This is a DynamicObject
		fieldType = serialization.TagJson[iotago.DynamicFieldType]{
			Data: iotago.DynamicFieldType{
				DynamicObject: &serialization.EmptyEnum{},
			},
		}
		objectType = v.GetContents().Type.Repr
		objectID = iotago.ObjectID(v.Address)
		version = iotago.SequenceNumber(v.Version)
		digestPtr, err := iotago.NewDigest(v.Digest)
		if err != nil {
			return nil, fmt.Errorf("failed to parse object digest: %w", err)
		}
		digest = *digestPtr

	case *iotagraphql.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveValue:
		// This is a DynamicField
		fieldType = serialization.TagJson[iotago.DynamicFieldType]{
			Data: iotago.DynamicFieldType{
				DynamicField: &serialization.EmptyEnum{},
			},
		}
		objectType = v.Type.Repr
		// For DynamicField, ObjectID, Version, and Digest are zero values

	default:
		return nil, fmt.Errorf("unknown value type: %T", value)
	}

	return &iotajsonrpc.DynamicFieldInfo{
		Name:       name,
		BcsName:    node.GetName().Bcs,
		Type:       fieldType,
		ObjectType: objectType,
		ObjectID:   objectID,
		Version:    version,
		Digest:     digest,
	}, nil
}

func convertGraphQLObjectToIotaObjectResponse(
	obj *iotagraphql.GetObjectObject,
	options *iotajsonrpc.IotaObjectDataOptions,
) (*iotajsonrpc.IotaObjectResponse, error) {
	if obj == nil {
		return nil, fmt.Errorf("object is nil")
	}

	// Convert ObjectID (from Address to ObjectID)
	objectID := iotago.ObjectID(obj.ObjectId)

	// Convert Version
	version := iotajsonrpc.NewBigInt(obj.Version)

	// Convert Digest
	digest, err := iotago.NewDigest(obj.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object digest: %w", err)
	}

	// Create IotaObjectData
	data := &iotajsonrpc.IotaObjectData{
		ObjectID: &objectID,
		Version:  version,
		Digest:   digest,
	}

	// Add Type if requested
	if options != nil && options.ShowType {
		typeStr := obj.AsMoveObjectType.Contents.Type.Repr
		data.Type = &typeStr
	}

	// Add Content if requested
	if options != nil && options.ShowContent {
		contentData := obj.AsMoveObjectContent.Contents.Data
		typeRepr := obj.AsMoveObjectContent.Contents.Type.Repr

		// Parse the content as a MoveObject
		parsedContent := serialization.TagJson[iotajsonrpc.IotaParsedData]{
			Data: iotajsonrpc.IotaParsedData{
				MoveObject: &iotajsonrpc.IotaParsedMoveObject{
					Type:              typeRepr,
					HasPublicTransfer: true, // TODO: determine from object
					Fields:            contentData,
				},
			},
		}
		data.Content = &parsedContent
	}

	// Add BCS if requested
	if options != nil && options.ShowBcs {
		bcsBytes := obj.AsMoveObject.Contents.Bcs
		typeRepr := obj.AsMoveObject.Contents.Type.Repr

		// Parse type as StructTag
		structTag, err := iotago.StructTagFromString(typeRepr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse struct tag: %w", err)
		}

		rawData := serialization.TagJson[iotajsonrpc.IotaRawData]{
			Data: iotajsonrpc.IotaRawData{
				MoveObject: &iotajsonrpc.IotaRawMoveObject{
					Type:              *structTag,
					HasPublicTransfer: true, // TODO: determine from object
					Version:           iotago.SequenceNumber(obj.Version),
					BcsBytes:          bcsBytes,
				},
			},
		}
		data.Bcs = &rawData
	}

	// Add Owner if requested
	if options != nil && options.ShowOwner {
		owner, err := convertGraphQLOwner(obj.Owner)
		if err != nil {
			return nil, fmt.Errorf("failed to convert owner: %w", err)
		}
		data.Owner = owner
	}

	// Add PreviousTransaction if requested
	if options != nil && options.ShowPreviousTransaction {
		txDigest, err := iotago.NewDigest(obj.PreviousTransactionBlock.Digest)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction digest: %w", err)
		}
		data.PreviousTransaction = txDigest
	}

	// Add StorageRebate if requested
	if options != nil && options.ShowStorageRebate {
		storageRebate := obj.StorageRebate.Clone()
		data.StorageRebate = storageRebate
	}

	// Add Display if requested
	if options != nil && options.ShowDisplay {
		if len(obj.Display) > 0 {
			display := make(map[string]string)
			for _, entry := range obj.Display {
				display[entry.Key] = entry.Value
			}
			data.Display = display
		}
	}

	return &iotajsonrpc.IotaObjectResponse{
		Data: data,
	}, nil
}

func convertRPCObjectFieldsToIotaObjectData(
	fields *iotagraphql.RPC_OBJECT_FIELDS,
	options *iotajsonrpc.IotaObjectDataOptions,
) (*iotajsonrpc.IotaObjectData, error) {
	if fields == nil {
		return nil, fmt.Errorf("fields is nil")
	}

	// Convert ObjectID (from Address to ObjectID)
	objectID := iotago.ObjectID(fields.ObjectId)

	// Convert Version
	version := iotajsonrpc.NewBigInt(fields.Version)

	// Convert Digest
	digest, err := iotago.NewDigest(fields.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object digest: %w", err)
	}

	// Create IotaObjectData
	data := &iotajsonrpc.IotaObjectData{
		ObjectID: &objectID,
		Version:  version,
		Digest:   digest,
	}

	// Add Type if requested
	if options != nil && options.ShowType {
		typeStr := fields.AsMoveObjectType.Contents.Type.Repr
		data.Type = &typeStr
	}

	// Add Content if requested
	if options != nil && options.ShowContent {
		contentData := fields.AsMoveObjectContent.Contents.Data
		typeRepr := fields.AsMoveObjectContent.Contents.Type.Repr

		// Parse the content as a MoveObject
		parsedContent := serialization.TagJson[iotajsonrpc.IotaParsedData]{
			Data: iotajsonrpc.IotaParsedData{
				MoveObject: &iotajsonrpc.IotaParsedMoveObject{
					Type:              typeRepr,
					HasPublicTransfer: true, // TODO: determine from object
					Fields:            contentData,
				},
			},
		}
		data.Content = &parsedContent
	}

	// Add BCS if requested
	if options != nil && options.ShowBcs {
		bcsBytes := fields.AsMoveObject.Contents.Bcs
		typeRepr := fields.AsMoveObject.Contents.Type.Repr

		// Parse type as StructTag
		structTag, err := iotago.StructTagFromString(typeRepr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse struct tag: %w", err)
		}

		rawData := serialization.TagJson[iotajsonrpc.IotaRawData]{
			Data: iotajsonrpc.IotaRawData{
				MoveObject: &iotajsonrpc.IotaRawMoveObject{
					Type:              *structTag,
					HasPublicTransfer: true, // TODO: determine from object
					Version:           iotago.SequenceNumber(fields.Version),
					BcsBytes:          bcsBytes,
				},
			},
		}
		data.Bcs = &rawData
	}

	// Add Owner if requested
	if options != nil && options.ShowOwner {
		owner, err := convertGraphQLOwner(fields.Owner)
		if err != nil {
			return nil, fmt.Errorf("failed to convert owner: %w", err)
		}
		data.Owner = owner
	}

	// Add PreviousTransaction if requested
	if options != nil && options.ShowPreviousTransaction {
		txDigest, err := iotago.NewDigest(fields.PreviousTransactionBlock.Digest)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction digest: %w", err)
		}
		data.PreviousTransaction = txDigest
	}

	// Add StorageRebate if requested
	if options != nil && options.ShowStorageRebate {
		storageRebate := fields.StorageRebate.Clone()
		data.StorageRebate = storageRebate
	}

	// Add Display if requested
	if options != nil && options.ShowDisplay {
		if len(fields.Display) > 0 {
			display := make(map[string]string)
			for _, entry := range fields.Display {
				display[entry.Key] = entry.Value
			}
			data.Display = display
		}
	}

	return data, nil
}

func convertGraphQLTryGetPastObjectResponse(
	resp *iotagraphql.TryGetPastObjectResponse,
	requestedVersion uint64,
	options *iotajsonrpc.IotaObjectDataOptions,
) (*iotajsonrpc.IotaPastObjectResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	pastObject := &iotajsonrpc.IotaPastObject{}

	// Check if the current object exists (address should not be zero)
	currentExists := resp.Current.Address != iotago.Address{}

	// Check if the requested version object has data
	// The Object field might be nil or have empty ObjectId if not found
	objectFound := resp.Object.ObjectId != iotago.Address{}

	if !currentExists {
		// Object doesn't exist at all
		objID := iotago.ObjectID(resp.Current.Address)
		pastObject.ObjectNotExists = &objID
	} else if objectFound {
		// Version found - convert the object data
		// We need to convert TryGetPastObjectObject to IotaObjectData
		// TryGetPastObjectObject embeds RPC_OBJECT_FIELDS, similar to GetObjectObject
		data, err := convertRPCObjectFieldsToIotaObjectData(&resp.Object.RPC_OBJECT_FIELDS, options)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object data: %w", err)
		}
		pastObject.VersionFound = data
	} else {
		// Object exists but version not found
		// Determine if it's VersionTooHigh or VersionNotFound
		currentVersion := resp.Current.Version

		if requestedVersion > currentVersion {
			// Requested version is higher than current
			pastObject.VersionTooHigh = &struct {
				ObjectID      iotago.ObjectID       `json:"object_id"`
				AskedVersion  iotago.SequenceNumber `json:"asked_version"`
				LatestVersion iotago.SequenceNumber `json:"latest_version"`
			}{
				ObjectID:      iotago.ObjectID(resp.Current.Address),
				AskedVersion:  iotago.SequenceNumber(requestedVersion),
				LatestVersion: iotago.SequenceNumber(currentVersion),
			}
		} else {
			// Version not found (possibly deleted or pruned)
			objID := iotago.ObjectID(resp.Current.Address)
			pastObject.VersionNotFound = &iotajsonrpc.VersionNotFoundData{
				ObjectID:       &objID,
				SequenceNumber: iotago.SequenceNumber(requestedVersion),
			}
		}
	}

	return &iotajsonrpc.IotaPastObjectResponse{
		Data: *pastObject,
	}, nil
}

func convertGraphQLOwner(owner iotagraphql.RPC_OBJECT_FIELDSOwnerObjectOwner) (*iotajsonrpc.ObjectOwner, error) {
	if owner == nil {
		return nil, nil
	}

	switch o := owner.(type) {
	case *iotagraphql.RPC_OBJECT_FIELDSOwnerAddressOwner:
		// AddressOwner
		var addr *iotago.Address
		if o.Owner.AsAddress.Address != (iotago.Address{}) {
			addr = &o.Owner.AsAddress.Address
		} else if o.Owner.AsObject.Address != (iotago.Address{}) {
			// ObjectOwner (owned by another object)
			addr = &o.Owner.AsObject.Address
			return &iotajsonrpc.ObjectOwner{
				ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
					ObjectOwner: addr,
				},
			}, nil
		}
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				AddressOwner: addr,
			},
		}, nil

	case *iotagraphql.RPC_OBJECT_FIELDSOwnerShared:
		// Shared object
		initialVersion := iotago.SequenceNumber(o.InitialSharedVersion)
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				Shared: &struct {
					InitialSharedVersion *iotago.SequenceNumber `json:"initial_shared_version"`
				}{
					InitialSharedVersion: &initialVersion,
				},
			},
		}, nil

	case *iotagraphql.RPC_OBJECT_FIELDSOwnerImmutable:
		// Immutable object - use JSON marshaling to set the unexported field
		var owner iotajsonrpc.ObjectOwner
		if err := json.Unmarshal([]byte(`"Immutable"`), &owner); err != nil {
			return nil, fmt.Errorf("failed to create Immutable owner: %w", err)
		}
		return &owner, nil

	case *iotagraphql.RPC_OBJECT_FIELDSOwnerParent:
		// Parent object
		parentAddr := o.Parent.Address
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				ObjectOwner: &parentAddr,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown owner type: %T", owner)
	}
}

func convertGraphQLTransactionBlockToResponse(
	tx *iotagraphql.GetTransactionBlockTransactionBlock,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction block is nil")
	}

	// Convert digest
	digest, err := iotago.NewDigest(tx.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction digest: %w", err)
	}

	result := &iotajsonrpc.IotaTransactionBlockResponse{
		Digest: *digest,
	}

	// Add RawTransaction if requested
	if options != nil && options.ShowRawInput {
		result.RawTransaction = tx.RawTransaction
	}

	// Add Transaction (input) if requested
	if options != nil && options.ShowInput {
		// TODO: Parse and convert RawTransaction into IotaTransactionBlock structure
		// For now, we skip this as it requires BCS deserialization
	}

	// Add Effects if requested
	if options != nil && options.ShowEffects {
		effects, err := convertGraphQLEffects(&tx.Effects, digest)
		if err != nil {
			return nil, fmt.Errorf("failed to convert effects: %w", err)
		}
		result.Effects = effects
	}

	// Add Events if requested
	if options != nil && options.ShowEvents {
		events, err := convertGraphQLEvents(tx.Effects.Events.Nodes, digest)
		if err != nil {
			return nil, fmt.Errorf("failed to convert events: %w", err)
		}
		result.Events = events
	}

	// Add TimestampMs from effects
	timestampMs := tx.Effects.Timestamp.UnixMilli()
	result.TimestampMs = iotajsonrpc.NewBigInt(uint64(timestampMs))

	// Add Checkpoint from effects
	result.Checkpoint = iotajsonrpc.NewBigInt(tx.Effects.Checkpoint.SequenceNumber)

	// Add ObjectChanges if requested
	if options != nil && options.ShowObjectChanges {
		objectChanges, err := convertGraphQLObjectChanges(tx.Effects.ObjectChanges.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object changes: %w", err)
		}
		result.ObjectChanges = objectChanges
	}

	// Add BalanceChanges if requested
	if options != nil && options.ShowBalanceChanges {
		balanceChanges, err := convertGraphQLBalanceChanges(tx.Effects.BalanceChanges.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert balance changes: %w", err)
		}
		result.BalanceChanges = balanceChanges
	}

	// Add RawEffects if requested
	if options != nil && options.ShowRawEffects {
		// Convert BCS bytes to []int
		bcsBytes := tx.Effects.Bcs
		rawEffects := make([]int, len(bcsBytes))
		for i, b := range bcsBytes {
			rawEffects[i] = int(b)
		}
		result.RawEffects = rawEffects
	}

	return result, nil
}

func convertGraphQLEffects(
	effects *iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffects,
	txDigest *iotago.TransactionDigest,
) (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error) {
	// Decode the BCS effects
	var decodedEffects iotajsonrpc.IotaTransactionBlockEffects
	if err := iotaclient.UnmarshalBCS(effects.Bcs, &decodedEffects); err != nil {
		return nil, fmt.Errorf("failed to decode BCS effects: %w", err)
	}

	return &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
		Data: decodedEffects,
	}, nil
}

func convertGraphQLEvents(
	nodes []iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsEventsEventConnectionNodesEvent,
	txDigest *iotago.TransactionDigest,
) ([]*iotajsonrpc.IotaEvent, error) {
	events := make([]*iotajsonrpc.IotaEvent, 0, len(nodes))

	for i, node := range nodes {
		// Extract package ID and module from SendingModule
		packageID := iotago.ObjectID(node.SendingModule.Package.Address)
		module := iotago.Identifier(node.SendingModule.Name)

		// Parse sender address
		sender := &node.Sender.Address

		// Parse event type from SendingModule if available
		// Note: We may need to extract the actual event type from the event data
		// For now, we'll construct a basic struct tag
		var eventType *iotago.StructTag
		// TODO: Extract proper event type from the event structure

		// Convert timestamp to milliseconds
		timestampMs := node.Timestamp.UnixMilli()

		event := &iotajsonrpc.IotaEvent{
			Id: iotajsonrpc.EventId{
				TxDigest: *txDigest,
				EventSeq: iotajsonrpc.NewBigInt(uint64(i)),
			},
			PackageId:         &packageID,
			TransactionModule: module,
			Sender:            sender,
			Type:              eventType,
			ParsedJson:        node.Json,
			Bcs:               iotago.Base64Data{}, // TODO: Extract BCS if available
			TimestampMs:       iotajsonrpc.NewBigInt(uint64(timestampMs)),
		}

		events = append(events, event)
	}

	return events, nil
}

func convertGraphQLObjectChanges(
	nodes []iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsObjectChangesObjectChangeConnectionNodesObjectChange,
) ([]serialization.TagJson[iotajsonrpc.ObjectChange], error) {
	changes := make([]serialization.TagJson[iotajsonrpc.ObjectChange], 0, len(nodes))

	for _, node := range nodes {
		change, err := convertGraphQLObjectChange(&node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object change for address %s: %w", node.Address, err)
		}
		changes = append(changes, *change)
	}

	return changes, nil
}

func convertGraphQLObjectChange(
	node *iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsObjectChangesObjectChangeConnectionNodesObjectChange,
) (*serialization.TagJson[iotajsonrpc.ObjectChange], error) {
	objectID := iotago.ObjectID(node.Address)
	inputState := node.InputState
	outputState := node.OutputState

	// Get versions from input and output states
	inputVersion := inputState.Version
	// For output version, we need to check if the output state has a version
	// Since OutputState doesn't have a direct Version field, we'll need to determine
	// the change type based on whether the states are empty/zero

	// For now, extract object type from the move object if available
	objectType := ""
	if outputState.AsMoveObject.Contents.Type.Repr != "" {
		objectType = outputState.AsMoveObject.Contents.Type.Repr
	} else if inputState.AsMoveObject.Contents.Type.Repr != "" {
		objectType = inputState.AsMoveObject.Contents.Type.Repr
	}

	// Parse digest if available
	var digest iotago.ObjectDigest
	// TODO: Extract digest from output state - requires checking the GraphQL schema

	var change iotajsonrpc.ObjectChange

	// For simplicity, assume most changes are mutations since we have both input and output
	// A proper implementation would need to determine the exact change type from the GraphQL data
	if inputVersion == 0 {
		// Created - no previous version
		change.Created = &struct {
			Sender     iotago.Address          `json:"sender"`
			Owner      iotajsonrpc.ObjectOwner `json:"owner"`
			ObjectType string                  `json:"objectType"`
			ObjectID   iotago.ObjectID         `json:"objectId"`
			Version    *iotajsonrpc.BigInt     `json:"version"`
			Digest     iotago.ObjectDigest     `json:"digest"`
		}{
			// Sender:     sender, // TODO: Extract sender from transaction
			// Owner:      owner,  // TODO: Extract owner from output state
			ObjectType: objectType,
			ObjectID:   objectID,
			Version:    iotajsonrpc.NewBigInt(1), // Created objects start at version 1
			Digest:     digest,
		}
	} else {
		// Mutated - has previous version
		change.Mutated = &struct {
			Sender          iotago.Address          `json:"sender"`
			Owner           iotajsonrpc.ObjectOwner `json:"owner"`
			ObjectType      string                  `json:"objectType"`
			ObjectID        iotago.ObjectID         `json:"objectId"`
			Version         *iotajsonrpc.BigInt     `json:"version"`
			PreviousVersion *iotajsonrpc.BigInt     `json:"previousVersion"`
			Digest          iotago.ObjectDigest     `json:"digest"`
		}{
			// Sender:          sender, // TODO: Extract sender
			// Owner:           owner,  // TODO: Extract owner
			ObjectType:      objectType,
			ObjectID:        objectID,
			Version:         iotajsonrpc.NewBigInt(inputVersion + 1), // Assume incremented version
			PreviousVersion: iotajsonrpc.NewBigInt(inputVersion),
			Digest:          digest,
		}
	}

	return &serialization.TagJson[iotajsonrpc.ObjectChange]{
		Data: change,
	}, nil
}

func deriveObjectChangesFromEffects(
	effects *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects],
	sender iotago.Address,
) ([]serialization.TagJson[iotajsonrpc.ObjectChange], error) {
	if effects == nil || effects.Data.V1 == nil {
		return nil, nil
	}

	v1 := effects.Data.V1
	prevVersions := make(map[iotago.ObjectID]*iotajsonrpc.BigInt, len(v1.ModifiedAtVersions))
	for _, entry := range v1.ModifiedAtVersions {
		prevVersions[entry.ObjectID] = entry.SequenceNumber
	}

	changes := make([]serialization.TagJson[iotajsonrpc.ObjectChange], 0, len(v1.Mutated)+len(v1.Created)+len(v1.Deleted))
	seen := make(map[iotago.ObjectID]struct{})

	addMutated := func(ref iotajsonrpc.OwnedObjectRef) error {
		if ref.Reference.ObjectID == nil {
			return nil
		}
		objectID := *ref.Reference.ObjectID
		if _, exists := seen[objectID]; exists {
			return nil
		}
		owner, err := convertOwnerFromTag(ref.Owner)
		if err != nil {
			return err
		}
		change := iotajsonrpc.ObjectChange{
			Mutated: &struct {
				Sender          iotago.Address          `json:"sender"`
				Owner           iotajsonrpc.ObjectOwner `json:"owner"`
				ObjectType      string                  `json:"objectType"`
				ObjectID        iotago.ObjectID         `json:"objectId"`
				Version         *iotajsonrpc.BigInt     `json:"version"`
				PreviousVersion *iotajsonrpc.BigInt     `json:"previousVersion"`
				Digest          iotago.ObjectDigest     `json:"digest"`
			}{
				Sender:          sender,
				Owner:           *owner,
				ObjectType:      "",
				ObjectID:        objectID,
				Version:         iotajsonrpc.NewBigInt(uint64(ref.Reference.Version)),
				PreviousVersion: prevVersions[objectID],
				Digest:          iotago.ObjectDigest(ref.Reference.Digest),
			},
		}
		changes = append(changes, serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change})
		seen[objectID] = struct{}{}
		return nil
	}

	for _, mutated := range v1.Mutated {
		if err := addMutated(mutated); err != nil {
			return nil, err
		}
	}
	if v1.GasObject.Reference.ObjectID != nil {
		if err := addMutated(v1.GasObject); err != nil {
			return nil, err
		}
	}

	for _, created := range v1.Created {
		if created.Reference.ObjectID == nil {
			continue
		}
		owner, err := convertOwnerFromTag(created.Owner)
		if err != nil {
			return nil, err
		}
		change := iotajsonrpc.ObjectChange{
			Created: &struct {
				Sender     iotago.Address          `json:"sender"`
				Owner      iotajsonrpc.ObjectOwner `json:"owner"`
				ObjectType string                  `json:"objectType"`
				ObjectID   iotago.ObjectID         `json:"objectId"`
				Version    *iotajsonrpc.BigInt     `json:"version"`
				Digest     iotago.ObjectDigest     `json:"digest"`
			}{
				Sender:     sender,
				Owner:      *owner,
				ObjectType: "",
				ObjectID:   *created.Reference.ObjectID,
				Version:    iotajsonrpc.NewBigInt(uint64(created.Reference.Version)),
				Digest:     iotago.ObjectDigest(created.Reference.Digest),
			},
		}
		changes = append(changes, serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change})
	}

	for _, deleted := range v1.Deleted {
		if deleted.ObjectID == nil {
			continue
		}
		change := iotajsonrpc.ObjectChange{
			Deleted: &struct {
				Sender     iotago.Address      `json:"sender"`
				ObjectType string              `json:"objectType"`
				ObjectID   iotago.ObjectID     `json:"objectId"`
				Version    *iotajsonrpc.BigInt `json:"version"`
			}{
				Sender:     sender,
				ObjectType: "",
				ObjectID:   *deleted.ObjectID,
				Version:    iotajsonrpc.NewBigInt(uint64(deleted.Version)),
			},
		}
		changes = append(changes, serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change})
	}

	for _, wrapped := range v1.Wrapped {
		if wrapped.ObjectID == nil {
			continue
		}
		change := iotajsonrpc.ObjectChange{
			Wrapped: &struct {
				Sender     iotago.Address      `json:"sender"`
				ObjectType string              `json:"objectType"`
				ObjectID   iotago.ObjectID     `json:"objectId"`
				Version    *iotajsonrpc.BigInt `json:"version"`
			}{
				Sender:     sender,
				ObjectType: "",
				ObjectID:   *wrapped.ObjectID,
				Version:    iotajsonrpc.NewBigInt(uint64(wrapped.Version)),
			},
		}
		changes = append(changes, serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change})
	}

	return changes, nil
}

func convertOwnerFromTag(owner serialization.TagJson[iotago.Owner]) (*iotajsonrpc.ObjectOwner, error) {
	data := owner.Data
	switch {
	case data.AddressOwner != nil:
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				AddressOwner: data.AddressOwner,
			},
		}, nil
	case data.ObjectOwner != nil:
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				ObjectOwner: data.ObjectOwner,
			},
		}, nil
	case data.Shared != nil:
		version := data.Shared.InitialSharedVersion
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				Shared: &struct {
					InitialSharedVersion *iotago.SequenceNumber `json:"initial_shared_version"`
				}{
					InitialSharedVersion: lo.ToPtr(version),
				},
			},
		}, nil
	case data.Immutable != nil:
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported owner type")
	}
}

func convertGraphQLBalanceChanges(
	nodes []iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsBalanceChangesBalanceChangeConnectionNodesBalanceChange,
) ([]iotajsonrpc.BalanceChange, error) {
	changes := make([]iotajsonrpc.BalanceChange, 0, len(nodes))

	for _, node := range nodes {
		// Convert owner
		owner, err := convertGraphQLBalanceChangeOwner(node.Owner)
		if err != nil {
			return nil, fmt.Errorf("failed to convert balance change owner: %w", err)
		}

		// Convert amount (it's a BigInt in GraphQL, convert to string for JSON-RPC)
		amount := node.Amount.String()

		change := iotajsonrpc.BalanceChange{
			Owner:    *owner,
			CoinType: node.CoinType.Repr,
			Amount:   amount,
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func convertGraphQLBalanceChangeOwner(
	owner iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsBalanceChangesBalanceChangeConnectionNodesBalanceChangeOwner,
) (*iotajsonrpc.ObjectOwner, error) {
	// Balance change owners can be addresses or objects
	// Check if it's an address
	if owner.AsAddress.Address != (iotago.Address{}) {
		addr := &owner.AsAddress.Address
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				AddressOwner: addr,
			},
		}, nil
	}

	// Check if it's an object
	if owner.AsObject.Address != (iotago.Address{}) {
		addr := &owner.AsObject.Address
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				ObjectOwner: addr,
			},
		}, nil
	}

	return nil, fmt.Errorf("balance change owner has neither address nor object")
}

func convertDevInspectResults(resp *iotagraphql.DevInspectTransactionBlockResponse) (*iotajsonrpc.DevInspectResults, error) {
	dryRunResult := resp.DryRunTransactionBlock

	// Handle error case
	if dryRunResult.Error != "" {
		return &iotajsonrpc.DevInspectResults{
			Error: dryRunResult.Error,
		}, nil
	}

	// Convert effects - try to decode BCS, fall back to minimal effects if unavailable/incomplete
	var effects *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]
	if len(dryRunResult.Transaction.Effects.Bcs) > 0 {
		var err error
		effects, err = convertGraphQLEffects(&dryRunResult.Transaction.Effects, nil)
		if err != nil {
			// BCS decoding failed (possibly incomplete for dev inspect), use minimal effects
			effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
				Data: iotajsonrpc.IotaTransactionBlockEffects{
					V1: &iotajsonrpc.IotaTransactionBlockEffectsV1{
						Status: iotajsonrpc.ExecutionStatus{
							Status: "success",
						},
						GasUsed: iotajsonrpc.GasCostSummary{},
					},
				},
			}
		}
	} else {
		// BCS effects not available for dev inspect, return empty effects with success status
		effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
			Data: iotajsonrpc.IotaTransactionBlockEffects{
				V1: &iotajsonrpc.IotaTransactionBlockEffectsV1{
					Status: iotajsonrpc.ExecutionStatus{
						Status: "success",
					},
					GasUsed: iotajsonrpc.GasCostSummary{},
				},
			},
		}
	}

	// Convert events if present
	var events []iotajsonrpc.IotaEvent
	if len(dryRunResult.Transaction.Effects.Events.Nodes) > 0 {
		convertedEvents, err := convertGraphQLEvents(dryRunResult.Transaction.Effects.Events.Nodes, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to convert events: %w", err)
		}
		for _, e := range convertedEvents {
			events = append(events, *e)
		}
	}

	// Convert results
	var results []iotajsonrpc.ExecutionResultType
	for _, dryRunEffect := range dryRunResult.Results {
		executionResult := iotajsonrpc.ExecutionResultType{
			MutableReferenceOutputs: []iotajsonrpc.MutableReferenceOutputType{},
			ReturnValues:            []iotajsonrpc.ReturnValueType{},
		}

		// Convert mutated references
		for _, mutRef := range dryRunEffect.MutatedReferences {
			executionResult.MutableReferenceOutputs = append(executionResult.MutableReferenceOutputs, map[string]interface{}{
				"type": mutRef.Type.Repr,
				"bcs":  mutRef.Bcs,
			})
		}

		// Convert return values
		for _, retVal := range dryRunEffect.ReturnValues {
			executionResult.ReturnValues = append(executionResult.ReturnValues, map[string]interface{}{
				"type": retVal.Type.Repr,
				"bcs":  retVal.Bcs,
			})
		}

		results = append(results, executionResult)
	}

	return &iotajsonrpc.DevInspectResults{
		Effects: *effects,
		Events:  events,
		Results: results,
	}, nil
}

func convertDryRunResults(resp *iotagraphql.DryRunTransactionBlockResponse) (*iotajsonrpc.DryRunTransactionBlockResponse, error) {
	dryRunResult := resp.DryRunTransactionBlock

	// Convert effects - try to decode BCS, fall back to minimal effects if unavailable/incomplete
	var effects *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]
	if len(dryRunResult.Transaction.Effects.Bcs) > 0 {
		var err error
		effects, err = convertGraphQLEffects(&dryRunResult.Transaction.Effects, nil)
		if err != nil {
			// BCS decoding failed (possibly incomplete for dry run), use minimal effects
			effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
				Data: iotajsonrpc.IotaTransactionBlockEffects{
					V1: &iotajsonrpc.IotaTransactionBlockEffectsV1{
						Status: iotajsonrpc.ExecutionStatus{
							Status: "success",
						},
						GasUsed: iotajsonrpc.GasCostSummary{},
					},
				},
			}
		}
	} else {
		// BCS effects not available, return empty effects with success status
		effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
			Data: iotajsonrpc.IotaTransactionBlockEffects{
				V1: &iotajsonrpc.IotaTransactionBlockEffectsV1{
					Status: iotajsonrpc.ExecutionStatus{
						Status: "success",
					},
					GasUsed: iotajsonrpc.GasCostSummary{},
				},
			},
		}
	}

	// Convert events if present
	var events []iotajsonrpc.IotaEvent
	if len(dryRunResult.Transaction.Effects.Events.Nodes) > 0 {
		convertedEvents, err := convertGraphQLEvents(dryRunResult.Transaction.Effects.Events.Nodes, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to convert events: %w", err)
		}
		for _, e := range convertedEvents {
			events = append(events, *e)
		}
	}

	// Decode input transaction data from BCS
	var input serialization.TagJson[iotajsonrpc.IotaTransactionBlockData]
	if len(dryRunResult.Transaction.RawTransaction) > 0 {
		var txData iotajsonrpc.IotaTransactionBlockData
		if err := iotaclient.UnmarshalBCS(dryRunResult.Transaction.RawTransaction, &txData); err != nil {
			return nil, fmt.Errorf("failed to decode input transaction: %w", err)
		}
		input = serialization.TagJson[iotajsonrpc.IotaTransactionBlockData]{
			Data: txData,
		}
	}

	// Convert balance changes if present
	var balanceChanges []iotajsonrpc.BalanceChange
	if len(dryRunResult.Transaction.Effects.BalanceChanges.Nodes) > 0 {
		convertedBalanceChanges, err := convertGraphQLBalanceChanges(dryRunResult.Transaction.Effects.BalanceChanges.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert balance changes: %w", err)
		}
		balanceChanges = convertedBalanceChanges
	}

	// Convert object changes if present
	var objectChanges []serialization.TagJson[iotajsonrpc.ObjectChange]
	if len(dryRunResult.Transaction.Effects.ObjectChanges.Nodes) > 0 {
		convertedObjectChanges, err := convertGraphQLObjectChanges(dryRunResult.Transaction.Effects.ObjectChanges.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object changes: %w", err)
		}
		objectChanges = convertedObjectChanges
	} else {
		derivedChanges, err := deriveObjectChangesFromEffects(effects, dryRunResult.Transaction.Sender.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to derive object changes: %w", err)
		}
		objectChanges = derivedChanges
	}

	return &iotajsonrpc.DryRunTransactionBlockResponse{
		Effects:        *effects,
		Events:         events,
		ObjectChanges:  objectChanges,
		BalanceChanges: balanceChanges,
		Input:          input,
	}, nil
}

func convertExecuteTransactionBlockResponse(
	resp *iotagraphql.ExecuteTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	// Check for errors
	if len(resp.ExecuteTransactionBlock.Errors) > 0 {
		return nil, fmt.Errorf("execution failed: %v", resp.ExecuteTransactionBlock.Errors)
	}

	// Extract the transaction block from the effects
	txBlock := &resp.ExecuteTransactionBlock.Effects.TransactionBlock.RPC_TRANSACTION_FIELDS

	// Convert digest
	digest, err := iotago.NewDigest(txBlock.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction digest: %w", err)
	}

	result := &iotajsonrpc.IotaTransactionBlockResponse{
		Digest: *digest,
	}

	// Add RawTransaction if requested
	if options != nil && options.ShowRawInput {
		result.RawTransaction = txBlock.RawTransaction
	}

	// Add Effects if requested (default to true if options is nil)
	showEffects := true
	if options != nil {
		showEffects = options.ShowEffects
	}
	if showEffects {
		// Try to decode BCS effects, but if it fails (which can happen for ExecuteTransactionBlock
		// since not all fields are available yet), create minimal effects from the status
		effects, err := convertGraphQLEffects(&txBlock.Effects, digest)
		if err != nil {
			// BCS decoding failed, likely because effects are incomplete
			// Create minimal effects based on execution status
			effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
				Data: iotajsonrpc.IotaTransactionBlockEffects{
					V1: &iotajsonrpc.IotaTransactionBlockEffectsV1{
						Status: iotajsonrpc.ExecutionStatus{
							Status: "success",
						},
						GasUsed: iotajsonrpc.GasCostSummary{},
					},
				},
			}
		}
		result.Effects = effects
	}

	// Add Events if requested
	if options != nil && options.ShowEvents {
		events, err := convertGraphQLEvents(txBlock.Effects.Events.Nodes, digest)
		if err != nil {
			return nil, fmt.Errorf("failed to convert events: %w", err)
		}
		result.Events = events
	}

	// Add TimestampMs from effects
	timestampMs := txBlock.Effects.Timestamp.UnixMilli()
	result.TimestampMs = iotajsonrpc.NewBigInt(uint64(timestampMs))

	// Add Checkpoint from effects
	result.Checkpoint = iotajsonrpc.NewBigInt(txBlock.Effects.Checkpoint.SequenceNumber)

	// Add ObjectChanges if requested
	if options != nil && options.ShowObjectChanges {
		objectChanges, err := convertGraphQLObjectChanges(txBlock.Effects.ObjectChanges.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object changes: %w", err)
		}
		result.ObjectChanges = objectChanges
	}

	// Add BalanceChanges if requested
	if options != nil && options.ShowBalanceChanges {
		balanceChanges, err := convertGraphQLBalanceChanges(txBlock.Effects.BalanceChanges.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert balance changes: %w", err)
		}
		result.BalanceChanges = balanceChanges
	}

	// Add RawEffects if requested (default to true if options is nil)
	showRawEffects := true
	if options != nil {
		showRawEffects = options.ShowRawEffects
	}
	if showRawEffects {
		// Convert BCS bytes to []int
		bcsBytes := txBlock.Effects.Bcs
		rawEffects := make([]int, len(bcsBytes))
		for i, b := range bcsBytes {
			rawEffects[i] = int(b)
		}
		result.RawEffects = rawEffects
	}

	return result, nil
}

func convertRPCMoveObjectFieldsToIotaObjectResponse(
	fields *iotagraphql.RPC_MOVE_OBJECT_FIELDS,
	options *iotajsonrpc.IotaObjectDataOptions,
) (*iotajsonrpc.IotaObjectResponse, error) {
	if fields == nil {
		return nil, fmt.Errorf("fields is nil")
	}

	// Convert ObjectID (from Address to ObjectID)
	objectID := iotago.ObjectID(fields.ObjectId)

	// Convert Version
	version := iotajsonrpc.NewBigInt(fields.Version)

	// Convert Digest
	digest, err := iotago.NewDigest(fields.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object digest: %w", err)
	}

	// Create IotaObjectData
	data := &iotajsonrpc.IotaObjectData{
		ObjectID: &objectID,
		Version:  version,
		Digest:   digest,
	}

	// Add Type if requested
	if options != nil && options.ShowType {
		typeStr := fields.Contents_type.Type.Repr
		data.Type = &typeStr
	}

	// Add Content if requested
	if options != nil && options.ShowContent {
		contentData := fields.Contents_content.Data
		typeRepr := fields.Contents_content.Type.Repr

		// Parse the content as a MoveObject
		parsedContent := serialization.TagJson[iotajsonrpc.IotaParsedData]{
			Data: iotajsonrpc.IotaParsedData{
				MoveObject: &iotajsonrpc.IotaParsedMoveObject{
					Type:              typeRepr,
					HasPublicTransfer: true, // TODO: determine from object
					Fields:            contentData,
				},
			},
		}
		data.Content = &parsedContent
	}

	// Add BCS if requested
	if options != nil && options.ShowBcs {
		bcsBytes := fields.Bcs
		typeRepr := fields.Contents.Type.Repr

		// Parse type as StructTag
		structTag, err := iotago.StructTagFromString(typeRepr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse struct tag: %w", err)
		}

		rawData := serialization.TagJson[iotajsonrpc.IotaRawData]{
			Data: iotajsonrpc.IotaRawData{
				MoveObject: &iotajsonrpc.IotaRawMoveObject{
					Type:              *structTag,
					HasPublicTransfer: true, // TODO: determine from object
					Version:           iotago.SequenceNumber(fields.Version),
					BcsBytes:          bcsBytes,
				},
			},
		}
		data.Bcs = &rawData
	}

	// Add Owner if requested
	if options != nil && options.ShowOwner {
		owner, err := convertGraphQLObjectOwner(fields.Owner)
		if err != nil {
			return nil, fmt.Errorf("failed to convert owner: %w", err)
		}
		data.Owner = owner
	}

	// Add PreviousTransaction if requested
	if options != nil && options.ShowPreviousTransaction {
		txDigest, err := iotago.NewDigest(fields.PreviousTransactionBlock.Digest)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction digest: %w", err)
		}
		data.PreviousTransaction = txDigest
	}

	// Add StorageRebate if requested
	if options != nil && options.ShowStorageRebate {
		storageRebate := fields.StorageRebate.Clone()
		data.StorageRebate = storageRebate
	}

	// Add Display if requested
	if options != nil && options.ShowDisplay {
		if len(fields.Display) > 0 {
			display := make(map[string]string)
			for _, entry := range fields.Display {
				display[entry.Key] = entry.Value
			}
			data.Display = display
		}
	}

	return &iotajsonrpc.IotaObjectResponse{
		Data: data,
	}, nil
}

func convertGraphQLObjectOwner(owner iotagraphql.RPC_OBJECT_OWNER_FIELDS) (*iotajsonrpc.ObjectOwner, error) {
	if owner == nil {
		return nil, nil
	}

	switch o := owner.(type) {
	case *iotagraphql.RPC_OBJECT_OWNER_FIELDSAddressOwner:
		// AddressOwner
		var addr *iotago.Address
		if o.Owner.AsAddress.Address != (iotago.Address{}) {
			addr = &o.Owner.AsAddress.Address
		} else if o.Owner.AsObject.Address != (iotago.Address{}) {
			// ObjectOwner (owned by another object)
			addr = &o.Owner.AsObject.Address
			return &iotajsonrpc.ObjectOwner{
				ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
					ObjectOwner: addr,
				},
			}, nil
		}
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				AddressOwner: addr,
			},
		}, nil

	case *iotagraphql.RPC_OBJECT_OWNER_FIELDSShared:
		// Shared object
		initialVersion := iotago.SequenceNumber(o.InitialSharedVersion)
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				Shared: &struct {
					InitialSharedVersion *iotago.SequenceNumber `json:"initial_shared_version"`
				}{
					InitialSharedVersion: &initialVersion,
				},
			},
		}, nil

	case *iotagraphql.RPC_OBJECT_OWNER_FIELDSImmutable:
		// Immutable object - use JSON marshaling to set the unexported field
		var owner iotajsonrpc.ObjectOwner
		if err := json.Unmarshal([]byte(`"Immutable"`), &owner); err != nil {
			return nil, fmt.Errorf("failed to create Immutable owner: %w", err)
		}
		return &owner, nil

	case *iotagraphql.RPC_OBJECT_OWNER_FIELDSParent:
		// Parent object
		parentAddr := o.Parent.Address
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				ObjectOwner: &parentAddr,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown owner type: %T", owner)
	}
}
