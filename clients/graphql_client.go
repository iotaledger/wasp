package clients

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
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

var defaultFalse = false

type GraphQLClient struct {
	url        string
	client     graphql.Client
	httpClient *http.Client
}

func NewGraphQLClient(url string) *GraphQLClient {
	return NewGraphQLClientWithTimeout(url, 30*time.Second)
}

func NewGraphQLClientWithTimeout(url string, timeout time.Duration) *GraphQLClient {
	httpClient := &http.Client{
		Timeout: timeout,
	}
	return &GraphQLClient{
		url:        strings.TrimRight(url, "/"),
		client:     graphql.NewClient(url, httpClient),
		httpClient: httpClient,
	}
}

// GetGraphQLClient returns the underlying GraphQL client for custom GraphQL queries.
func (c *GraphQLClient) GetGraphQLClient() graphql.Client {
	return c.client
}

// extractObjectOptions extracts boolean options for object queries.
// It uses options from opts, or defaults to false.
func extractObjectOptions(
	opts *iotajsonrpc.IotaObjectDataOptions,
) *objectDataShowOptions {
	if opts != nil {
		return &objectDataShowOptions{
			ShowBcs:                 &opts.ShowBcs,
			ShowOwner:               &opts.ShowOwner,
			ShowPreviousTransaction: &opts.ShowPreviousTransaction,
			ShowContent:             &opts.ShowContent,
			ShowDisplay:             &opts.ShowDisplay,
			ShowType:                &opts.ShowType,
			ShowStorageRebate:       &opts.ShowStorageRebate,
		}
	}

	return &objectDataShowOptions{
		ShowBcs:                 &defaultFalse,
		ShowOwner:               &defaultFalse,
		ShowPreviousTransaction: &defaultFalse,
		ShowContent:             &defaultFalse,
		ShowDisplay:             &defaultFalse,
		ShowType:                &defaultFalse,
		ShowStorageRebate:       &defaultFalse,
	}
}

// extractTransactionOptions extracts boolean options for transaction queries.
// It uses options from opts, or defaults to false.
func extractTransactionOptions(
	opts *iotajsonrpc.IotaTransactionBlockResponseOptions,
) *transactionShowOptions {
	if opts != nil {
		return &transactionShowOptions{
			ShowBalanceChanges: &opts.ShowBalanceChanges,
			ShowEffects:        &opts.ShowEffects,
			ShowRawEffects:     &opts.ShowRawEffects,
			ShowEvents:         &opts.ShowEvents,
			ShowInput:          &opts.ShowInput,
			ShowObjectChanges:  &opts.ShowObjectChanges,
			ShowRawInput:       &opts.ShowRawInput,
		}
	}

	return &transactionShowOptions{
		ShowBalanceChanges: &defaultFalse,
		ShowEffects:        &defaultFalse,
		ShowRawEffects:     &defaultFalse,
		ShowEvents:         &defaultFalse,
		ShowInput:          &defaultFalse,
		ShowObjectChanges:  &defaultFalse,
		ShowRawInput:       &defaultFalse,
	}
}

// transactionShowOptions encapsulates the boolean show options for transaction queries
type transactionShowOptions struct {
	ShowBalanceChanges *bool
	ShowEffects        *bool
	ShowRawEffects     *bool
	ShowEvents         *bool
	ShowInput          *bool
	ShowObjectChanges  *bool
	ShowRawInput       *bool
}

// objectDataShowOptions encapsulates the boolean show options for object data queries
type objectDataShowOptions struct {
	ShowBcs                 *bool
	ShowContent             *bool
	ShowDisplay             *bool
	ShowType                *bool
	ShowOwner               *bool
	ShowPreviousTransaction *bool
	ShowStorageRebate       *bool
}

// convertToShowOptions converts IotaTransactionBlockResponseOptions to transactionShowOptions
func convertToShowOptions(reqOptions *iotajsonrpc.IotaTransactionBlockResponseOptions) *transactionShowOptions {
	if reqOptions == nil {
		// Return default false values for all options instead of nil pointers
		// GraphQL @include directives require boolean values, not null

		return &transactionShowOptions{
			ShowBalanceChanges: &defaultFalse,
			ShowEffects:        &defaultFalse,
			ShowRawEffects:     &defaultFalse,
			ShowEvents:         &defaultFalse,
			ShowInput:          &defaultFalse,
			ShowObjectChanges:  &defaultFalse,
			ShowRawInput:       &defaultFalse,
		}
	}

	return &transactionShowOptions{
		ShowBalanceChanges: &reqOptions.ShowBalanceChanges,
		ShowEffects:        &reqOptions.ShowEffects,
		ShowRawEffects:     &reqOptions.ShowRawEffects,
		ShowEvents:         &reqOptions.ShowEvents,
		ShowInput:          &reqOptions.ShowInput,
		ShowObjectChanges:  &reqOptions.ShowObjectChanges,
		ShowRawInput:       &reqOptions.ShowRawInput,
	}
}

// convertToObjectDataShowOptions converts IotaObjectDataOptions to objectDataShowOptions
func convertToObjectDataShowOptions(reqOptions *iotajsonrpc.IotaObjectDataOptions) *objectDataShowOptions {
	if reqOptions == nil {
		// Return default false values for all options instead of nil pointers
		// GraphQL @include directives require boolean values, not null

		return &objectDataShowOptions{
			ShowBcs:                 &defaultFalse,
			ShowContent:             &defaultFalse,
			ShowDisplay:             &defaultFalse,
			ShowType:                &defaultFalse,
			ShowOwner:               &defaultFalse,
			ShowPreviousTransaction: &defaultFalse,
			ShowStorageRebate:       &defaultFalse,
		}
	}

	return &objectDataShowOptions{
		ShowBcs:                 &reqOptions.ShowBcs,
		ShowContent:             &reqOptions.ShowContent,
		ShowDisplay:             &reqOptions.ShowDisplay,
		ShowType:                &reqOptions.ShowType,
		ShowOwner:               &reqOptions.ShowOwner,
		ShowPreviousTransaction: &reqOptions.ShowPreviousTransaction,
		ShowStorageRebate:       &reqOptions.ShowStorageRebate,
	}
}

// bigIntToUint64 safely converts a BigInt to uint64, returning an error if it doesn't fit.
func bigIntToUint64(b *iotajsonrpc.BigInt, fieldName string) (uint64, error) {
	if b == nil {
		return 0, fmt.Errorf("%s is nil", fieldName)
	}
	if !b.IsUint64() {
		return 0, fmt.Errorf("%s value %s exceeds uint64 maximum", fieldName, b.String())
	}
	return b.Uint64(), nil
}

// dereferenceObjectRefSlice converts a slice of ObjectRef pointers to a slice of ObjectRef values.
func dereferenceObjectRefSlice(refs []*iotago.ObjectRef) []iotago.ObjectRef {
	result := make([]iotago.ObjectRef, len(refs))
	for i, ref := range refs {
		result[i] = *ref
	}
	return result
}

// hasNextPage checks if pagination should continue.
func hasNextPage(hasNext bool, cursor string) bool {
	return hasNext && cursor != ""
}

// cursorToObjectID tries to parse a GraphQL cursor into an ObjectID, accepting
// hex (with or without 0x prefix) and base64-encoded cursor formats.
func cursorToObjectID(cursor string) (*iotago.ObjectID, error) {
	if cursor == "" {
		return nil, fmt.Errorf("cursor is empty")
	}

	if objID, err := iotago.ObjectIDFromHex(cursor); err == nil {
		return objID, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("cursor is not base64 or hex: %w", err)
	}

	if objID, err := iotago.ObjectIDFromHex(string(decoded)); err == nil {
		return objID, nil
	}

	if len(decoded) == iotago.AddressLen {
		var arr [iotago.AddressLen]byte
		copy(arr[:], decoded)
		return iotago.ObjectIDFromArray(arr), nil
	}

	return nil, fmt.Errorf("cursor does not represent an object ID")
}

// Interfaces for dynamic field conversion to eliminate duplication
// These interfaces abstract over the GraphQL-generated types to allow unified handling

type dynamicFieldNameInfo struct {
	JSON []byte
	Type struct{ Repr string }
	Bcs  iotago.Base64Data
}

type dynamicFieldMoveObjectInfo struct {
	TypeRepr string
	Address  iotago.ObjectID
	Version  iotago.SequenceNumber
	Digest   string
}

type dynamicFieldMoveValueInfo struct {
	TypeRepr string
}

// validateRequired validates that a required parameter is not nil.
func validateRequired(val interface{}, paramName string) error {
	if val == nil {
		return fmt.Errorf("%s is required", paramName)
	}
	return nil
}

// Query builds and executes a custom GraphQL query returning the raw response bytes.
func (c *GraphQLClient) Query(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	reqBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return rawBytes, nil
}

var _ L1Client = &GraphQLClient{}

func (c *GraphQLClient) GetDynamicFieldObject(
	ctx context.Context,
	req iotaclient.GetDynamicFieldObjectRequest,
) (*iotajsonrpc.IotaObjectResponse, error) {
	return nil, fmt.Errorf("not implemented: %s", "GetDynamicFieldObject")
}

func (c *GraphQLClient) GetDynamicFields(
	ctx context.Context,
	req iotaclient.GetDynamicFieldsRequest,
) (*iotajsonrpc.DynamicFieldPage, error) {
	var cursor *string
	if req.Cursor != nil {
		cursorStr := req.Cursor.String()
		cursor = &cursorStr
	}

	objResp, objErr := iotagraphql.GetObjectDynamicFields(ctx, c.client, *req.ParentObjectID, req.Limit, cursor)
	if objErr == nil && len(objResp.Object.DynamicFields.Nodes) > 0 {
		return convertObjectDynamicFieldsResponse(objResp)
	}

	ownerResp, ownerErr := iotagraphql.GetDynamicFields(ctx, c.client, *req.ParentObjectID, req.Limit, cursor)
	if ownerErr == nil && len(ownerResp.Owner.DynamicFields.Nodes) > 0 {
		return convertOwnerDynamicFieldsResponse(ownerResp)
	}

	if objErr != nil {
		return nil, fmt.Errorf("failed to get dynamic fields as object: %w", objErr)
	}
	if ownerErr != nil {
		return nil, fmt.Errorf("failed to get dynamic fields as owner: %w", ownerErr)
	}

	return &iotajsonrpc.DynamicFieldPage{
		Data:        []iotajsonrpc.DynamicFieldInfo{},
		HasNextPage: false,
		NextCursor:  nil,
	}, nil
}

func convertOwnerDynamicFieldsResponse(resp *iotagraphql.GetDynamicFieldsResponse) (*iotajsonrpc.DynamicFieldPage, error) {
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
	if hasNextPage(resp.Owner.DynamicFields.PageInfo.HasNextPage, resp.Owner.DynamicFields.PageInfo.EndCursor) {
		nextCursor = iotago.MustObjectIDFromHex(resp.Owner.DynamicFields.PageInfo.EndCursor)
	}

	return &iotajsonrpc.DynamicFieldPage{
		Data:        data,
		HasNextPage: resp.Owner.DynamicFields.PageInfo.HasNextPage,
		NextCursor:  nextCursor,
	}, nil
}

func convertObjectDynamicFieldsResponse(resp *iotagraphql.GetObjectDynamicFieldsResponse) (*iotajsonrpc.DynamicFieldPage, error) {
	nodes := resp.Object.DynamicFields.Nodes
	data := make([]iotajsonrpc.DynamicFieldInfo, len(nodes))
	for i, node := range nodes {
		converted, err := convertObjectDynamicFieldToInfo(&node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert dynamic field at index %d: %w", i, err)
		}
		data[i] = *converted
	}

	var nextCursor *iotago.ObjectID
	if hasNextPage(resp.Object.DynamicFields.PageInfo.HasNextPage, resp.Object.DynamicFields.PageInfo.EndCursor) {
		nextCursor = iotago.MustObjectIDFromHex(resp.Object.DynamicFields.PageInfo.EndCursor)
	}

	return &iotajsonrpc.DynamicFieldPage{
		Data:        data,
		HasNextPage: resp.Object.DynamicFields.PageInfo.HasNextPage,
		NextCursor:  nextCursor,
	}, nil
}

func (c *GraphQLClient) GetOwnedObjects(
	ctx context.Context,
	req iotaclient.GetOwnedObjectsRequest,
) (*iotajsonrpc.ObjectsPage, error) {
	if err := validateRequired(req.Address, "address"); err != nil {
		return nil, err
	}

	var cursorPtr *string
	if req.Cursor != nil {
		cursorPtr = lo.ToPtr(req.Cursor.String())
	}

	var opts *objectDataShowOptions
	if req.Query != nil {
		opts = convertToObjectDataShowOptions(req.Query.Options)
	} else {
		opts = convertToObjectDataShowOptions(nil)
	}

	var filter iotagraphql.ObjectFilter
	if req.Query != nil && req.Query.Filter != nil {
		if req.Query.Filter.StructType != nil {
			filter.Type = req.Query.Filter.StructType.String()
		}

		if req.Query.Filter.Package != nil {
			filter.Type = req.Query.Filter.Package.String()
		}
	}

	resp, err := iotagraphql.GetOwnedObjects(ctx, c.client, *req.Address, req.Limit, cursorPtr,
		opts.ShowBcs, opts.ShowContent, opts.ShowDisplay, opts.ShowType, opts.ShowOwner, opts.ShowPreviousTransaction, opts.ShowStorageRebate, &filter)
	if err != nil {
		return nil, err
	}

	nodes := resp.Address.Objects.Nodes
	objects := make([]iotajsonrpc.IotaObjectResponse, 0, len(nodes))
	for _, node := range nodes {
		obj, err := convertRPCMoveObjectFieldsToIotaObjectResponse(&node.RPC_MOVE_OBJECT_FIELDS, req.Query.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object: %w", err)
		}
		objects = append(objects, *obj)
	}

	// Note: GraphQL cursors are opaque and may not always be parseable as ObjectIDs
	var nextCursor *iotago.ObjectID
	if hasNextPage(resp.Address.Objects.PageInfo.HasNextPage, resp.Address.Objects.PageInfo.EndCursor) {
		if parsed, err := cursorToObjectID(resp.Address.Objects.PageInfo.EndCursor); err == nil {
			nextCursor = parsed
		}
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
	return nil, fmt.Errorf("not implemented: %s", "QueryEvents")
}

func (c *GraphQLClient) QueryTransactionBlocks(
	ctx context.Context,
	req iotaclient.QueryTransactionBlocksRequest,
) (*iotajsonrpc.TransactionBlocksPage, error) {
	var first, last *int
	var after, before *string

	if req.Limit != nil {
		if req.DescendingOrder {
			last = req.Limit
		} else {
			first = req.Limit
		}
	}

	if req.Cursor != nil {
		cursorStr := req.Cursor.String()
		if req.DescendingOrder {
			before = &cursorStr
		} else {
			after = &cursorStr
		}
	}

	var filter *iotagraphql.TransactionBlockFilter
	if req.Query != nil && req.Query.Filter != nil {
		filter = convertTransactionFilterToGraphQL(req.Query.Filter)
	}

	var opts *iotajsonrpc.IotaTransactionBlockResponseOptions
	if req.Query != nil {
		opts = req.Query.Options
	}
	txOpts := extractTransactionOptions(opts)

	resp, err := iotagraphql.QueryTransactionBlocks(ctx, c.client, first, last, before, after,
		txOpts.ShowBalanceChanges, txOpts.ShowEffects, txOpts.ShowRawEffects, txOpts.ShowEvents, txOpts.ShowInput, txOpts.ShowObjectChanges, txOpts.ShowRawInput, filter)
	if err != nil {
		return nil, err
	}

	nodes := resp.TransactionBlocks.Nodes
	data := make([]iotajsonrpc.IotaTransactionBlockResponse, 0, len(nodes))

	for _, node := range nodes {
		txResp, err := convertQueryTransactionBlockNodeToResponse(&node, req.Query)
		if err != nil {
			return nil, fmt.Errorf("failed to convert transaction block: %w", err)
		}
		data = append(data, *txResp)
	}

	var nextCursor *iotago.TransactionDigest
	if hasNextPage(resp.TransactionBlocks.PageInfo.HasNextPage, resp.TransactionBlocks.PageInfo.EndCursor) {
		nextCursor = iotago.MustNewDigest(resp.TransactionBlocks.PageInfo.EndCursor)
	}

	return &iotajsonrpc.TransactionBlocksPage{
		Data:        data,
		NextCursor:  nextCursor,
		HasNextPage: resp.TransactionBlocks.PageInfo.HasNextPage,
	}, nil
}

func (c *GraphQLClient) ResolveNameServiceAddress(ctx context.Context, iotaName string) (*iotago.Address, error) {
	return nil, fmt.Errorf("not implemented: %s", "ResolveNameServiceAddress")
}

func (c *GraphQLClient) ResolveNameServiceNames(
	ctx context.Context,
	req iotaclient.ResolveNameServiceNamesRequest,
) (*iotajsonrpc.IotaNamePage, error) {
	return nil, fmt.Errorf("not implemented: %s", "ResolveNameServiceNames")
}

func (c *GraphQLClient) DevInspectTransactionBlock(
	ctx context.Context,
	req iotaclient.DevInspectTransactionBlockRequest,
) (*iotajsonrpc.DevInspectResults, error) {
	txBytes := req.TxKindBytes.String()

	gasPrice := uint64(iotaclient.DefaultGasPrice)
	if req.GasPrice != nil {
		var err error
		gasPrice, err = bigIntToUint64(req.GasPrice, "gasPrice")
		if err != nil {
			return nil, err
		}
	}

	txMeta := iotagraphql.TransactionMetadata{
		Sender:     *req.SenderAddress,
		GasPrice:   gasPrice,
		GasBudget:  iotaclient.DefaultGasBudget,
		GasSponsor: *req.SenderAddress,
	}

	opts := convertToShowOptions(req.Options)

	resp, err := iotagraphql.DevInspectTransactionBlock(ctx, c.client, txBytes, txMeta,
		opts.ShowBalanceChanges, opts.ShowEffects, opts.ShowRawEffects, opts.ShowEvents, opts.ShowInput, opts.ShowObjectChanges, opts.ShowRawInput)
	if err != nil {
		return nil, err
	}

	result, err := convertDevInspectResults(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DevInspectResults: %w", err)
	}

	return result, nil
}

func (c *GraphQLClient) DryRunTransaction(
	ctx context.Context,
	req iotaclient.DryRunTransactionRequest,
) (*iotajsonrpc.DryRunTransactionBlockResponse, error) {
	txBytes := req.TxDataBytes.String()
	opts := convertToShowOptions(req.Options)

	resp, err := iotagraphql.DryRunTransactionBlock(ctx, c.client, txBytes,
		opts.ShowBalanceChanges, opts.ShowEffects, opts.ShowRawEffects, opts.ShowEvents, opts.ShowInput, opts.ShowObjectChanges, opts.ShowRawInput)
	if err != nil {
		return nil, err
	}

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
		sigBytes := sig.Bytes()
		if sigBytes == nil {
			return nil, fmt.Errorf("signature %d has nil bytes", i)
		}
		signatures[i] = iotago.Base64Data(sigBytes).String()
	}
	opts := convertToShowOptions(req.Options)

	resp, err := iotagraphql.ExecuteTransactionBlock(ctx, c.client, txBytes, signatures,
		opts.ShowBalanceChanges, opts.ShowEffects, opts.ShowRawEffects, opts.ShowEvents, opts.ShowInput, opts.ShowObjectChanges, opts.ShowRawInput)
	if err != nil {
		return nil, err
	}

	return convertExecuteTransactionBlockResponse(resp, req.Options)
}

func (c *GraphQLClient) GetCommitteeInfo(
	ctx context.Context,
	epoch *iotajsonrpc.BigInt,
) (*iotajsonrpc.CommitteeInfo, error) {
	return nil, fmt.Errorf("not implemented: %s", "GetCommitteeInfo")
}

func (c *GraphQLClient) GetLatestIotaSystemState(ctx context.Context) (*iotajsonrpc.IotaSystemStateSummary, error) {
	resp, err := iotagraphql.GetLatestIotaSystemState(ctx, c.client)
	if err != nil {
		return nil, err
	}

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
	return nil, fmt.Errorf("not implemented: %s", "GetStakes")
}

func (c *GraphQLClient) GetStakesByIds(ctx context.Context, stakedIotaIds []iotago.ObjectID) ([]*iotajsonrpc.DelegatedStake, error) {
	return nil, fmt.Errorf("not implemented: %s", "GetStakesByIds")
}

func (c *GraphQLClient) GetValidatorsApy(ctx context.Context) (*iotajsonrpc.ValidatorsApy, error) {
	return nil, fmt.Errorf("not implemented: %s", "GetValidatorsApy")
}

func (c *GraphQLClient) BatchTransaction(
	ctx context.Context,
	req iotaclient.BatchTransactionRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "BatchTransaction")
}

func (c *GraphQLClient) MergeCoins(
	ctx context.Context,
	req iotaclient.MergeCoinsRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "MergeCoins")
}

func (c *GraphQLClient) MoveCall(
	ctx context.Context,
	req iotaclient.MoveCallRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "MoveCall")
}

func (c *GraphQLClient) fetchObjectRefs(ctx context.Context, objectIDs []*iotago.ObjectID) ([]*iotago.ObjectRef, error) {
	refs := make([]*iotago.ObjectRef, 0, len(objectIDs))
	for _, objID := range objectIDs {
		objResp, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: objID})
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", objID.String(), err)
		}
		if objResp.Data == nil {
			return nil, fmt.Errorf("object %s not found", objID.String())
		}
		ref := objResp.Data.Ref()
		refs = append(refs, &ref)
	}
	return refs, nil
}

func createInputObjectsFromRefs(refs []*iotago.ObjectRef) []iotajsonrpc.InputObjectKind {
	inputObjects := make([]iotajsonrpc.InputObjectKind, 0, len(refs))
	for _, ref := range refs {
		inputObjects = append(inputObjects, iotajsonrpc.InputObjectKind{
			"ImmOrOwnedMoveObject": map[string]interface{}{
				"objectId": ref.ObjectID.String(),
				"version":  ref.Version,
				"digest":   ref.Digest.String(),
			},
		})
	}
	return inputObjects
}

func (c *GraphQLClient) Pay(
	ctx context.Context,
	req iotaclient.PayRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	coinRefs, err := c.fetchObjectRefs(ctx, req.InputCoins)
	if err != nil {
		return nil, err
	}

	amounts := make([]uint64, len(req.Amount))
	for i, amt := range req.Amount {
		val, convErr := bigIntToUint64(amt, fmt.Sprintf("amount[%d]", i))
		if convErr != nil {
			return nil, convErr
		}
		amounts[i] = val
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	if err = ptb.Pay(coinRefs, req.Recipients, amounts); err != nil {
		return nil, fmt.Errorf("failed to build Pay transaction: %w", err)
	}
	pt := ptb.Finish()

	gasBudget := uint64(iotaclient.DefaultGasBudget)
	if req.GasBudget != nil {
		gasBudget, err = bigIntToUint64(req.GasBudget, "gasBudget")
		if err != nil {
			return nil, err
		}
	}

	// Gas must be provided explicitly because input coins are used in the Pay command
	if req.Gas == nil {
		return nil, fmt.Errorf("gas parameter is required for Pay via GraphQL (input coins cannot be used as gas)")
	}

	gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: req.Gas,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get gas object %s: %w", req.Gas.String(), err)
	}
	if gasObj.Data == nil {
		return nil, fmt.Errorf("gas object %s not found", req.Gas.String())
	}

	ref := gasObj.Data.Ref()
	gasPayment := []*iotago.ObjectRef{&ref}

	allRefs := append(append([]*iotago.ObjectRef{}, coinRefs...), gasPayment...)
	inputObjects := createInputObjectsFromRefs(allRefs)

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

	gasRefs := dereferenceObjectRefSlice(gasPayment)

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
	ptb := iotago.NewProgrammableTransactionBuilder()
	if err := ptb.PayAllIota(req.Recipient); err != nil {
		return nil, fmt.Errorf("failed to build PayAllIota transaction: %w", err)
	}
	pt := ptb.Finish()

	var err error
	gasBudget := uint64(iotaclient.DefaultGasBudget)
	if req.GasBudget != nil {
		gasBudget, err = bigIntToUint64(req.GasBudget, "gasBudget")
		if err != nil {
			return nil, err
		}
	}

	gasPayment := make([]*iotago.ObjectRef, 0, len(req.InputCoins))
	inputObjects := make([]iotajsonrpc.InputObjectKind, 0)

	var objResp *iotajsonrpc.IotaObjectResponse
	for _, coinID := range req.InputCoins {
		objResp, err = c.GetObject(ctx, iotaclient.GetObjectRequest{
			ObjectID: coinID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", coinID.String(), err)
		}

		if objResp.Data == nil {
			return nil, fmt.Errorf("object %s not found", coinID.String())
		}

		objRef := objResp.Data.Ref()
		gasPayment = append(gasPayment, &objRef)

		inputObjects = append(inputObjects, iotajsonrpc.InputObjectKind{
			"ImmOrOwnedMoveObject": map[string]interface{}{
				"objectId": objResp.Data.ObjectID.String(),
				"version":  objResp.Data.Version.Uint64(),
				"digest":   objResp.Data.Digest.String(),
			},
		})
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

	gasRefs := dereferenceObjectRefSlice(gasPayment)

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
	return nil, fmt.Errorf("not implemented: %s", "PayIota")
}

func (c *GraphQLClient) Publish(
	ctx context.Context,
	req iotaclient.PublishRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "Publish")
}

func (c *GraphQLClient) RequestAddStake(
	ctx context.Context,
	req iotaclient.RequestAddStakeRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "RequestAddStake")
}

func (c *GraphQLClient) RequestWithdrawStake(
	ctx context.Context,
	req iotaclient.RequestWithdrawStakeRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "RequestWithdrawStake")
}

func (c *GraphQLClient) SplitCoin(
	ctx context.Context,
	req iotaclient.SplitCoinRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "SplitCoin")
}

func (c *GraphQLClient) SplitCoinEqual(
	ctx context.Context,
	req iotaclient.SplitCoinEqualRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "SplitCoinEqual")
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
	if transferErr := ptb.TransferObject(req.Recipient, objectRef); transferErr != nil {
		return nil, fmt.Errorf("failed to build TransferObject transaction: %w", transferErr)
	}
	pt := ptb.Finish()

	gasBudget := uint64(iotaclient.DefaultGasBudget)
	if req.GasBudget != nil {
		gasBudget, err = bigIntToUint64(req.GasBudget, "gasBudget")
		if err != nil {
			return nil, err
		}
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
	ref := objResp.Data.Ref()
	return &ref, nil
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
	const pageLimit = int(50)
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
	return nil, fmt.Errorf("not implemented: %s", "TransferIota")
}

func (c *GraphQLClient) GetCoinObjsForTargetAmount(
	ctx context.Context,
	address *iotago.Address,
	targetAmount uint64,
	gasAmount uint64,
) (iotajsonrpc.Coins, error) {
	// This would require fetching coins via GraphQL and filtering/selecting client-side
	// Similar to GetIotaCoinsOwnedByAddress but with additional logic
	return nil, fmt.Errorf("not implemented: %s", "GetCoinObjsForTargetAmount")
}

func isResponseComplete(
	res *iotajsonrpc.IotaTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) bool {
	// In Rebased, it can happen that Effects are available before ObjectChanges are.
	// This function checks if ShowEffects/ShowObjectChanges are enabled, and validates the state of the response.

	if options == nil {
		return true
	}

	if options.ShowObjectChanges {
		if res.ObjectChanges == nil {
			return false
		}
		if options.ShowEffects && res.Effects == nil {
			return false
		}
		return true
	}

	if options.ShowEffects {
		return res.Effects != nil
	}

	return true
}

func (c *GraphQLClient) SignAndExecuteTransaction(
	ctx context.Context,
	req *iotaclient.SignAndExecuteTransactionRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	signature, err := req.Signer.SignTransactionBlock(req.TxDataBytes, iotasigner.DefaultIntent())
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction block: %w", err)
	}
	resp, err := c.ExecuteTransactionBlock(
		ctx,
		iotaclient.ExecuteTransactionBlockRequest{
			TxDataBytes: req.TxDataBytes,
			Signatures:  []*iotasigner.Signature{signature},
			Options:     req.Options,
			RequestType: iotajsonrpc.TxnRequestTypeWaitForLocalExecution,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	if !isResponseComplete(resp, req.Options) {
		resp, err = c.GetTransactionBlock(
			ctx, iotaclient.GetTransactionBlockRequest{
				Digest:  &resp.Digest,
				Options: req.Options,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("GetTransactionBlock failed: %w", err)
		}
	}

	return resp, err
}

func (c *GraphQLClient) UpdateObjectRef(
	ctx context.Context,
	ref *iotago.ObjectRef,
) (*iotago.ObjectRef, error) {
	return nil, fmt.Errorf("not implemented: %s", "UpdateObjectRef")
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
	return nil, fmt.Errorf("not implemented: %s", "MintToken")
}

func (c *GraphQLClient) GetIotaCoinsOwnedByAddress(ctx context.Context, address *iotago.Address) (iotajsonrpc.Coins, error) {
	return nil, fmt.Errorf("not implemented: %s", "GetIotaCoinsOwnedByAddress")
}

func (c *GraphQLClient) BatchGetObjectsOwnedByAddress(
	ctx context.Context,
	address *iotago.Address,
	options *iotajsonrpc.IotaObjectDataOptions,
	filterType string,
) ([]iotajsonrpc.IotaObjectResponse, error) {
	return nil, fmt.Errorf("not implemented: %s", "BatchGetObjectsOwnedByAddress")
}

func (c *GraphQLClient) BatchGetFilteredObjectsOwnedByAddress(
	ctx context.Context,
	address *iotago.Address,
	options *iotajsonrpc.IotaObjectDataOptions,
	filter func(*iotajsonrpc.IotaObjectData) bool,
) ([]iotajsonrpc.IotaObjectResponse, error) {
	return nil, fmt.Errorf("not implemented: %s", "BatchGetFilteredObjectsOwnedByAddress")
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

	var limitPtr *int
	if req.Limit > 0 {
		limitPtr = &req.Limit
	}

	var cursorPtr *string
	if req.Cursor != nil {
		cursorPtr = lo.ToPtr(req.Cursor.String())
	}

	resp, err := iotagraphql.GetAllCoins(ctx, c.client, *req.Owner, limitPtr, cursorPtr)
	if err != nil {
		return nil, err
	}

	nodes := resp.Address.Coins.Nodes
	coins := make([]*iotajsonrpc.Coin, 0, len(nodes))
	for _, node := range nodes {
		coin, err := convertGraphQLAllCoin(&node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert coin: %w", err)
		}
		coins = append(coins, coin)
	}

	var nextCursor *string
	if hasNextPage(resp.Address.Coins.PageInfo.HasNextPage, resp.Address.Coins.PageInfo.EndCursor) {
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
	objID := &meta.Address
	return &iotajsonrpc.IotaCoinMetadata{
		Name:        meta.Name,
		Symbol:      meta.Symbol,
		Decimals:    uint8(meta.Decimals), // #nosec G115 -- decimals is always < 256
		Description: meta.Description,
		IconUrl:     meta.IconUrl,
		Id:          objID,
	}, nil
}

func (c *GraphQLClient) GetCoins(ctx context.Context, req iotaclient.GetCoinsRequest) (*iotajsonrpc.CoinPage, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("owner address is required")
	}

	var limitPtr *int
	if req.Limit > 0 {
		limitPtr = &req.Limit
	}

	cursorPtr := req.Cursor

	resp, err := iotagraphql.GetCoins(ctx, c.client, *req.Owner, limitPtr, cursorPtr, req.CoinType)
	if err != nil {
		return nil, err
	}

	nodes := resp.Address.Coins.Nodes
	coins := make([]*iotajsonrpc.Coin, 0, len(nodes))
	for _, node := range nodes {
		coin, err := convertGraphQLCoin(&node)
		if err != nil {
			return nil, fmt.Errorf("failed to convert coin: %w", err)
		}
		coins = append(coins, coin)
	}

	var nextCursor *string
	if hasNextPage(resp.Address.Coins.PageInfo.HasNextPage, resp.Address.Coins.PageInfo.EndCursor) {
		nextCursor = &resp.Address.Coins.PageInfo.EndCursor
	}

	return &iotajsonrpc.CoinPage{
		Data:        coins,
		NextCursor:  nextCursor,
		HasNextPage: resp.Address.Coins.PageInfo.HasNextPage,
	}, nil
}

// coinNode defines the common interface for coin GraphQL nodes
type coinNode interface {
	GetAddress() iotago.Address
	GetDigest() string
	GetVersion() uint64
	GetCoinBalance() iotajsonrpc.BigInt
	GetContentsTypeRepr() string
	GetPreviousTxDigest() string
}

// coinNodeWrapper wraps GetCoinsAddressCoinsCoinConnectionNodesCoin
type coinNodeWrapper struct {
	*iotagraphql.GetCoinsAddressCoinsCoinConnectionNodesCoin
}

func (c coinNodeWrapper) GetAddress() iotago.Address         { return c.Address }
func (c coinNodeWrapper) GetDigest() string                  { return c.Digest }
func (c coinNodeWrapper) GetVersion() uint64                 { return c.Version }
func (c coinNodeWrapper) GetCoinBalance() iotajsonrpc.BigInt { return c.CoinBalance }
func (c coinNodeWrapper) GetContentsTypeRepr() string        { return c.Contents.Type.Repr }
func (c coinNodeWrapper) GetPreviousTxDigest() string        { return c.PreviousTransactionBlock.Digest }

// allCoinNodeWrapper wraps GetAllCoinsAddressCoinsCoinConnectionNodesCoin
type allCoinNodeWrapper struct {
	*iotagraphql.GetAllCoinsAddressCoinsCoinConnectionNodesCoin
}

func (c allCoinNodeWrapper) GetAddress() iotago.Address         { return c.Address }
func (c allCoinNodeWrapper) GetDigest() string                  { return c.Digest }
func (c allCoinNodeWrapper) GetVersion() uint64                 { return c.Version }
func (c allCoinNodeWrapper) GetCoinBalance() iotajsonrpc.BigInt { return c.CoinBalance }
func (c allCoinNodeWrapper) GetContentsTypeRepr() string        { return c.Contents.Type.Repr }
func (c allCoinNodeWrapper) GetPreviousTxDigest() string        { return c.PreviousTransactionBlock.Digest }

func convertCoinNode(node coinNode) (*iotajsonrpc.Coin, error) {
	coinType, err := iotajsonrpc.CoinTypeFromString(node.GetContentsTypeRepr())
	if err != nil {
		return nil, fmt.Errorf("invalid coin type %s: %w", node.GetContentsTypeRepr(), err)
	}

	objID := node.GetAddress()
	digest, err := iotago.NewDigest(node.GetDigest())
	if err != nil {
		return nil, fmt.Errorf("invalid digest %s: %w", node.GetDigest(), err)
	}

	txDigest, err := iotago.NewDigest(node.GetPreviousTxDigest())
	if err != nil {
		return nil, fmt.Errorf("invalid transaction digest %s: %w", node.GetPreviousTxDigest(), err)
	}

	balance := node.GetCoinBalance()
	return &iotajsonrpc.Coin{
		CoinType:            coinType,
		CoinObjectID:        &objID,
		Version:             iotajsonrpc.NewBigInt(node.GetVersion()),
		Digest:              digest,
		Balance:             balance.Clone(),
		PreviousTransaction: *txDigest,
	}, nil
}

func convertGraphQLCoin(node *iotagraphql.GetCoinsAddressCoinsCoinConnectionNodesCoin) (*iotajsonrpc.Coin, error) {
	return convertCoinNode(coinNodeWrapper{node})
}

func convertGraphQLAllCoin(node *iotagraphql.GetAllCoinsAddressCoinsCoinConnectionNodesCoin) (*iotajsonrpc.Coin, error) {
	return convertCoinNode(allCoinNodeWrapper{node})
}

func (c *GraphQLClient) GetTotalSupply(ctx context.Context, coinType string) (*iotajsonrpc.Supply, error) {
	return nil, fmt.Errorf("GetTotalSupply is not yet implemented for GraphQL client")
}

func (c *GraphQLClient) GetChainIdentifier(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not implemented: %s", "GetChainIdentifier")
}

func (c *GraphQLClient) GetCheckpoint(ctx context.Context, checkpointID *iotajsonrpc.BigInt) (*iotajsonrpc.Checkpoint, error) {
	return nil, fmt.Errorf("not implemented: %s", "GetCheckpoint")
}

func (c *GraphQLClient) GetCheckpoints(ctx context.Context, req iotaclient.GetCheckpointsRequest) (*iotajsonrpc.CheckpointPage, error) {
	return nil, fmt.Errorf("not implemented: %s", "GetCheckpoints")
}

func (c *GraphQLClient) GetEvents(ctx context.Context, digest *iotago.TransactionDigest) ([]*iotajsonrpc.IotaEvent, error) {
	return nil, fmt.Errorf("not implemented: %s", "GetEvents")
}

func (c *GraphQLClient) GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not implemented: %s", "GetLatestCheckpointSequenceNumber")
}

func (c *GraphQLClient) GetObject(ctx context.Context, req iotaclient.GetObjectRequest) (*iotajsonrpc.IotaObjectResponse, error) {
	if req.ObjectID == nil {
		return nil, fmt.Errorf("object ID is required")
	}
	objAddr := *req.ObjectID

	objOpts := extractObjectOptions(req.Options)

	resp, err := iotagraphql.GetObject(ctx, c.client, objAddr,
		objOpts.ShowBcs, objOpts.ShowOwner, objOpts.ShowPreviousTransaction, objOpts.ShowContent, objOpts.ShowDisplay, objOpts.ShowType, objOpts.ShowStorageRebate)
	if err != nil {
		return nil, err
	}

	return convertGraphQLObjectToIotaObjectResponse(&resp.Object, req.Options)
}

func (c *GraphQLClient) GetProtocolConfig(
	ctx context.Context,
	version *iotajsonrpc.BigInt,
) (*iotajsonrpc.ProtocolConfig, error) {
	return nil, fmt.Errorf("not implemented: %s", "GetProtocolConfig")
}

func (c *GraphQLClient) GetTotalTransactionBlocks(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not implemented: %s", "GetTotalTransactionBlocks")
}

func (c *GraphQLClient) GetTransactionBlock(ctx context.Context, req iotaclient.GetTransactionBlockRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if req.Digest == nil {
		return nil, fmt.Errorf("transaction digest is required")
	}

	txOpts := extractTransactionOptions(req.Options)

	resp, err := iotagraphql.GetTransactionBlock(ctx, c.client, req.Digest.String(),
		txOpts.ShowBalanceChanges, txOpts.ShowEffects, txOpts.ShowRawEffects, txOpts.ShowEvents, txOpts.ShowInput, txOpts.ShowObjectChanges, txOpts.ShowRawInput)
	if err != nil {
		return nil, err
	}

	return convertGraphQLTransactionBlockToResponse(&resp.TransactionBlock, req.Options)
}

func (c *GraphQLClient) MultiGetObjects(ctx context.Context, req iotaclient.MultiGetObjectsRequest) ([]iotajsonrpc.IotaObjectResponse, error) {
	return nil, fmt.Errorf("not implemented: %s", "MultiGetObjects")
}

func (c *GraphQLClient) MultiGetTransactionBlocks(
	ctx context.Context,
	req iotaclient.MultiGetTransactionBlocksRequest,
) ([]*iotajsonrpc.IotaTransactionBlockResponse, error) {
	return nil, fmt.Errorf("not implemented: %s", "MultiGetTransactionBlocks")
}

func (c *GraphQLClient) TryGetPastObject(
	ctx context.Context,
	req iotaclient.TryGetPastObjectRequest,
) (*iotajsonrpc.IotaPastObjectResponse, error) {
	if req.ObjectID == nil {
		return nil, fmt.Errorf("object ID is required")
	}
	objAddr := *req.ObjectID
	version := req.Version

	objOpts := extractObjectOptions(req.Options)

	resp, err := iotagraphql.TryGetPastObject(ctx, c.client, objAddr, &version,
		objOpts.ShowBcs, objOpts.ShowOwner, objOpts.ShowPreviousTransaction, objOpts.ShowContent, objOpts.ShowDisplay, objOpts.ShowType, objOpts.ShowStorageRebate)
	if err != nil {
		return nil, err
	}

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
	return nil, fmt.Errorf("not implemented: %s", "TryMultiGetPastObjects")
}

func (c *GraphQLClient) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	return fmt.Errorf("not implemented: %s", "RequestFunds")
}

func (c *GraphQLClient) Health(ctx context.Context) error {
	return fmt.Errorf("not implemented: %s", "Health")
}

func (c *GraphQLClient) L2() L2Client {
	// Not implemented for GraphQL client
	return nil
}

func (c *GraphQLClient) IotaClient() *iotaclient.Client {
	// Not implemented for GraphQL client
	return nil
}

func (c *GraphQLClient) DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error) {
	return iotago.PackageID{}, fmt.Errorf("not implemented: %s", "DeployISCContracts")
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
	return nil, fmt.Errorf("not implemented: %s", "SignAndExecuteTxWithRetry")
}

func (c *GraphQLClient) WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error) {
	return nil, fmt.Errorf("not implemented: %s", "WaitForNextVersionForTesting")
}

func convertGraphQLBalance(coinTypeRepr string, coinObjectCount uint64, totalBalance iotajsonrpc.BigInt) (*iotajsonrpc.Balance, error) {
	// When coinTypeRepr is empty, default to IOTA coin type (matching GraphQL query default)
	if coinTypeRepr == "" {
		coinTypeRepr = "0x2::iota::IOTA"
	}
	coinType, err := iotajsonrpc.CoinTypeFromString(coinTypeRepr)
	if err != nil {
		return nil, fmt.Errorf("invalid coin type %s: %w", coinTypeRepr, err)
	}

	// Handle nil totalBalance by defaulting to zero
	totalBalancePtr := iotajsonrpc.NewBigInt(0)
	if totalBalance.Int != nil {
		totalBalancePtr = totalBalance.Clone()
	}

	return &iotajsonrpc.Balance{
		CoinType:        coinType,
		CoinObjectCount: iotajsonrpc.NewBigInt(coinObjectCount),
		TotalBalance:    totalBalancePtr,
	}, nil
}

// convertDynamicFieldToInfo is a unified function to convert dynamic field info
// from either Owner or Object GraphQL queries
func convertDynamicFieldToInfo(nameInfo dynamicFieldNameInfo, moveObject *dynamicFieldMoveObjectInfo, moveValue *dynamicFieldMoveValueInfo) (*iotajsonrpc.DynamicFieldInfo, error) {
	// Convert Name field
	var nameValue any
	if err := json.Unmarshal(nameInfo.JSON, &nameValue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal name JSON: %w", err)
	}

	name := iotago.DynamicFieldName{
		Type:  nameInfo.Type.Repr,
		Value: nameValue,
	}

	var fieldType serialization.TagJson[iotago.DynamicFieldType]
	var objectType string
	var objectID iotago.ObjectID
	var version iotago.SequenceNumber
	var digest iotago.ObjectDigest

	if moveObject != nil {
		// This is a DynamicObject
		fieldType = serialization.TagJson[iotago.DynamicFieldType]{
			Data: iotago.DynamicFieldType{
				DynamicObject: &serialization.EmptyEnum{},
			},
		}
		objectType = moveObject.TypeRepr
		objectID = moveObject.Address
		version = moveObject.Version
		digestPtr, err := iotago.NewDigest(moveObject.Digest)
		if err != nil {
			return nil, fmt.Errorf("failed to parse object digest: %w", err)
		}
		digest = *digestPtr
	} else if moveValue != nil {
		// This is a DynamicField
		fieldType = serialization.TagJson[iotago.DynamicFieldType]{
			Data: iotago.DynamicFieldType{
				DynamicField: &serialization.EmptyEnum{},
			},
		}
		objectType = moveValue.TypeRepr
		// For DynamicField, ObjectID, Version, and Digest are zero values
	} else {
		return nil, fmt.Errorf("either moveObject or moveValue must be provided")
	}

	return &iotajsonrpc.DynamicFieldInfo{
		Name:       name,
		BcsName:    nameInfo.Bcs,
		Type:       fieldType,
		ObjectType: objectType,
		ObjectID:   objectID,
		Version:    version,
		Digest:     digest,
	}, nil
}

func convertGraphQLDynamicFieldToInfo(
	node *iotagraphql.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicField,
) (*iotajsonrpc.DynamicFieldInfo, error) {
	if node == nil {
		return nil, fmt.Errorf("convertGraphQLDynamicFieldToInfo: node is nil")
	}

	nameInfo := dynamicFieldNameInfo{
		JSON: node.GetName().Json,
		Type: struct{ Repr string }{Repr: node.GetName().Type.Repr},
		Bcs:  node.GetName().Bcs,
	}

	value := node.GetValue()
	switch v := value.(type) {
	case *iotagraphql.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveObject:
		return convertDynamicFieldToInfo(nameInfo, &dynamicFieldMoveObjectInfo{
			TypeRepr: v.GetContents().Type.Repr,
			Address:  v.Address,
			Version:  v.Version,
			Digest:   v.Digest,
		}, nil)
	case *iotagraphql.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveValue:
		return convertDynamicFieldToInfo(nameInfo, nil, &dynamicFieldMoveValueInfo{
			TypeRepr: v.Type.Repr,
		})
	default:
		return nil, fmt.Errorf("unknown value type: %T", value)
	}
}

func convertObjectDynamicFieldToInfo(
	node *iotagraphql.GetObjectDynamicFieldsObjectDynamicFieldsDynamicFieldConnectionNodesDynamicField,
) (*iotajsonrpc.DynamicFieldInfo, error) {
	if node == nil {
		return nil, fmt.Errorf("convertObjectDynamicFieldToInfo: node is nil")
	}

	nameInfo := dynamicFieldNameInfo{
		JSON: node.GetName().Json,
		Type: struct{ Repr string }{Repr: node.GetName().Type.Repr},
		Bcs:  node.GetName().Bcs,
	}

	value := node.GetValue()
	switch v := value.(type) {
	case *iotagraphql.GetObjectDynamicFieldsObjectDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveObject:
		return convertDynamicFieldToInfo(nameInfo, &dynamicFieldMoveObjectInfo{
			TypeRepr: v.GetContents().Type.Repr,
			Address:  v.Address,
			Version:  v.Version,
			Digest:   v.Digest,
		}, nil)
	case *iotagraphql.GetObjectDynamicFieldsObjectDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveValue:
		return convertDynamicFieldToInfo(nameInfo, nil, &dynamicFieldMoveValueInfo{
			TypeRepr: v.Type.Repr,
		})
	default:
		return nil, fmt.Errorf("unknown value type: %T", value)
	}
}

func applyGraphQLObjectOptions(
	data *iotajsonrpc.IotaObjectData,
	obj *iotagraphql.GetObjectObject,
	options *iotajsonrpc.IotaObjectDataOptions,
) error {
	if options == nil {
		return nil
	}

	if options.ShowType {
		typeStr := obj.AsMoveObjectType.Contents.Type.Repr
		data.Type = &typeStr
	}

	if options.ShowContent {
		contentData := obj.AsMoveObjectContent.Contents.Data
		typeRepr := obj.AsMoveObjectContent.Contents.Type.Repr
		parsedContent := serialization.TagJson[iotajsonrpc.IotaParsedData]{
			Data: iotajsonrpc.IotaParsedData{
				MoveObject: &iotajsonrpc.IotaParsedMoveObject{
					Type:              typeRepr,
					HasPublicTransfer: true,
					Fields:            contentData,
				},
			},
		}
		data.Content = &parsedContent
	}

	if options.ShowBcs {
		structTag, err := iotago.StructTagFromString(obj.AsMoveObject.Contents.Type.Repr)
		if err != nil {
			return fmt.Errorf("failed to parse struct tag: %w", err)
		}
		rawData := serialization.TagJson[iotajsonrpc.IotaRawData]{
			Data: iotajsonrpc.IotaRawData{
				MoveObject: &iotajsonrpc.IotaRawMoveObject{
					Type:              *structTag,
					HasPublicTransfer: true,
					Version:           obj.Version,
					BcsBytes:          obj.AsMoveObject.Contents.Bcs,
				},
			},
		}
		data.Bcs = &rawData
	}

	if options.ShowOwner {
		owner, err := convertGraphQLOwner(obj.Owner)
		if err != nil {
			return fmt.Errorf("failed to convert owner: %w", err)
		}
		data.Owner = owner
	}

	if options.ShowPreviousTransaction {
		txDigest, err := iotago.NewDigest(obj.PreviousTransactionBlock.Digest)
		if err != nil {
			return fmt.Errorf("failed to parse transaction digest: %w", err)
		}
		data.PreviousTransaction = txDigest
	}

	if options.ShowStorageRebate {
		data.StorageRebate = obj.StorageRebate.Clone()
	}

	if options.ShowDisplay && len(obj.Display) > 0 {
		display := make(map[string]string)
		for _, entry := range obj.Display {
			display[entry.Key] = entry.Value
		}
		data.Display = display
	}

	return nil
}

func convertGraphQLObjectToIotaObjectResponse(
	obj *iotagraphql.GetObjectObject,
	options *iotajsonrpc.IotaObjectDataOptions,
) (*iotajsonrpc.IotaObjectResponse, error) {
	if obj == nil {
		return nil, fmt.Errorf("object is nil")
	}

	digest, err := iotago.NewDigest(obj.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object digest: %w", err)
	}

	data := &iotajsonrpc.IotaObjectData{
		ObjectID: &obj.ObjectId,
		Version:  iotajsonrpc.NewBigInt(obj.Version),
		Digest:   digest,
	}

	if err := applyGraphQLObjectOptions(data, obj, options); err != nil {
		return nil, err
	}

	return &iotajsonrpc.IotaObjectResponse{Data: data}, nil
}

func applyRPCObjectFieldsOptions(
	data *iotajsonrpc.IotaObjectData,
	fields *iotagraphql.RPC_OBJECT_FIELDS,
	options *iotajsonrpc.IotaObjectDataOptions,
) error {
	if options == nil {
		return nil
	}

	if options.ShowType {
		typeStr := fields.AsMoveObjectType.Contents.Type.Repr
		data.Type = &typeStr
	}

	if options.ShowContent {
		parsedContent := serialization.TagJson[iotajsonrpc.IotaParsedData]{
			Data: iotajsonrpc.IotaParsedData{
				MoveObject: &iotajsonrpc.IotaParsedMoveObject{
					Type:              fields.AsMoveObjectContent.Contents.Type.Repr,
					HasPublicTransfer: true,
					Fields:            fields.AsMoveObjectContent.Contents.Data,
				},
			},
		}
		data.Content = &parsedContent
	}

	if options.ShowBcs {
		structTag, err := iotago.StructTagFromString(fields.AsMoveObject.Contents.Type.Repr)
		if err != nil {
			return fmt.Errorf("failed to parse struct tag: %w", err)
		}
		rawData := serialization.TagJson[iotajsonrpc.IotaRawData]{
			Data: iotajsonrpc.IotaRawData{
				MoveObject: &iotajsonrpc.IotaRawMoveObject{
					Type:              *structTag,
					HasPublicTransfer: true,
					Version:           fields.Version,
					BcsBytes:          fields.AsMoveObject.Contents.Bcs,
				},
			},
		}
		data.Bcs = &rawData
	}

	if options.ShowOwner {
		owner, err := convertGraphQLOwner(fields.Owner)
		if err != nil {
			return fmt.Errorf("failed to convert owner: %w", err)
		}
		data.Owner = owner
	}

	if options.ShowPreviousTransaction {
		txDigest, err := iotago.NewDigest(fields.PreviousTransactionBlock.Digest)
		if err != nil {
			return fmt.Errorf("failed to parse transaction digest: %w", err)
		}
		data.PreviousTransaction = txDigest
	}

	if options.ShowStorageRebate {
		data.StorageRebate = fields.StorageRebate.Clone()
	}

	if options.ShowDisplay && len(fields.Display) > 0 {
		display := make(map[string]string)
		for _, entry := range fields.Display {
			display[entry.Key] = entry.Value
		}
		data.Display = display
	}

	return nil
}

func convertRPCObjectFieldsToIotaObjectData(
	fields *iotagraphql.RPC_OBJECT_FIELDS,
	options *iotajsonrpc.IotaObjectDataOptions,
) (*iotajsonrpc.IotaObjectData, error) {
	if fields == nil {
		return nil, fmt.Errorf("fields is nil")
	}

	digest, err := iotago.NewDigest(fields.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object digest: %w", err)
	}

	data := &iotajsonrpc.IotaObjectData{
		ObjectID: &fields.ObjectId,
		Version:  iotajsonrpc.NewBigInt(fields.Version),
		Digest:   digest,
	}

	if err := applyRPCObjectFieldsOptions(data, fields, options); err != nil {
		return nil, err
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
		objID := resp.Current.Address
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
			pastObject.VersionTooHigh = &iotajsonrpc.VersionTooHigh{
				ObjectID:      resp.Current.Address,
				AskedVersion:  requestedVersion,
				LatestVersion: currentVersion,
			}
		} else {
			// Version not found (possibly deleted or pruned)
			objID := resp.Current.Address
			pastObject.VersionNotFound = &iotajsonrpc.VersionNotFoundData{
				ObjectID:       &objID,
				SequenceNumber: requestedVersion,
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
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				Shared: &struct {
					InitialSharedVersion *iotago.SequenceNumber `json:"initial_shared_version"`
				}{
					InitialSharedVersion: &o.InitialSharedVersion,
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

// convertTransactionFilterToGraphQL converts JSON-RPC TransactionFilter to GraphQL TransactionBlockFilter
func convertTransactionFilterToGraphQL(filter *iotajsonrpc.TransactionFilter) *iotagraphql.TransactionBlockFilter {
	if filter == nil {
		return nil
	}

	result := &iotagraphql.TransactionBlockFilter{}

	// Map the fields from JSON-RPC to GraphQL format
	if filter.FromAddress != nil {
		result.SignAddress = *filter.FromAddress
	}
	if filter.ToAddress != nil {
		result.RecvAddress = *filter.ToAddress
	}
	if filter.InputObject != nil {
		result.InputObject = *filter.InputObject
	}
	if filter.ChangedObject != nil {
		result.ChangedObject = *filter.ChangedObject
	}
	if filter.MoveFunction != nil {
		// Format: "package::module::function"
		result.Function = fmt.Sprintf("%s::%s::%s",
			filter.MoveFunction.Package.String(),
			filter.MoveFunction.Module,
			filter.MoveFunction.Function)
	}

	// Note: Some filters don't have direct mappings:
	// - Checkpoint -> AfterCheckpoint/AtCheckpoint/BeforeCheckpoint
	// - FromAndToAddress, FromOrToAddress -> need special handling
	// - TransactionKind -> Kind
	// For now, we only map the fields that have direct equivalents

	return result
}

// convertQueryTransactionBlockNodeToResponse converts a GraphQL transaction node to JSON-RPC response
func applyQueryNodeOptions(
	result *iotajsonrpc.IotaTransactionBlockResponse,
	node *iotagraphql.QueryTransactionBlocksTransactionBlocksTransactionBlockConnectionNodesTransactionBlock,
	digest *iotago.Digest,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) error {
	if options == nil {
		return nil
	}

	var decodedEffects *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]
	decodeEffects := func() (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error) {
		if decodedEffects != nil {
			return decodedEffects, nil
		}
		effects, err := convertGraphQLEffects(&node.Effects)
		if err != nil {
			return nil, err
		}
		decodedEffects = effects
		return effects, nil
	}

	if options.ShowRawInput {
		result.RawTransaction = node.Bcs
	}

	if options.ShowEffects {
		effects, err := decodeEffects()
		if err != nil {
			return fmt.Errorf("failed to convert effects: %w", err)
		}
		result.Effects = effects
	}

	if options.ShowEvents {
		events, err := convertGraphQLEvents(node.Effects.Events.Nodes, digest)
		if err != nil {
			return fmt.Errorf("failed to convert events: %w", err)
		}
		result.Events = events
	}

	if options.ShowObjectChanges {
		effects, err := decodeEffects()
		if err != nil {
			// Fall back to GraphQL nodes if BCS effects are unavailable
			objectChanges, convErr := convertGraphQLObjectChanges(node.Effects.ObjectChanges.Nodes)
			if convErr != nil {
				return fmt.Errorf("failed to convert object changes: %w", convErr)
			}
			result.ObjectChanges = objectChanges
			return nil
		}

		objectChanges, err := deriveObjectChangesFromEffects(effects, node.Sender.Address)
		if err != nil {
			return fmt.Errorf("failed to convert object changes: %w", err)
		}
		result.ObjectChanges = objectChanges
	}

	if options.ShowBalanceChanges {
		balanceChanges, err := convertGraphQLBalanceChanges(node.Effects.BalanceChanges.Nodes)
		if err != nil {
			return fmt.Errorf("failed to convert balance changes: %w", err)
		}
		result.BalanceChanges = balanceChanges
	}

	if options.ShowRawEffects {
		result.RawEffects = node.Effects.Bcs
	}

	return nil
}

func convertQueryTransactionBlockNodeToResponse(
	node *iotagraphql.QueryTransactionBlocksTransactionBlocksTransactionBlockConnectionNodesTransactionBlock,
	query *iotajsonrpc.IotaTransactionBlockResponseQuery,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if node == nil {
		return nil, fmt.Errorf("transaction node is nil")
	}

	digest, err := iotago.NewDigest(node.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction digest: %w", err)
	}

	result := &iotajsonrpc.IotaTransactionBlockResponse{Digest: *digest}
	// #nosec G115 -- timestamps from blockchain are always positive
	result.TimestampMs = iotajsonrpc.NewBigInt(uint64(node.Effects.Timestamp.UnixMilli()))
	result.Checkpoint = iotajsonrpc.NewBigInt(node.Effects.Checkpoint.SequenceNumber)

	var options *iotajsonrpc.IotaTransactionBlockResponseOptions
	if query != nil {
		options = query.Options
	}

	if err := applyQueryNodeOptions(result, node, digest, options); err != nil {
		return nil, err
	}

	return result, nil
}

func applyShowEffects(
	result *iotajsonrpc.IotaTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
	decodeEffects func() (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error),
) error {
	if options == nil || !options.ShowEffects {
		return nil
	}
	effects, err := decodeEffects()
	if err != nil {
		return fmt.Errorf("failed to convert effects: %w", err)
	}
	result.Effects = effects
	return nil
}

func applyShowEvents(
	result *iotajsonrpc.IotaTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
	eventNodes []iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsEventsEventConnectionNodesEvent,
	digest *iotago.TransactionDigest,
) error {
	if options == nil || !options.ShowEvents {
		return nil
	}
	events, err := convertGraphQLEvents(eventNodes, digest)
	if err != nil {
		return fmt.Errorf("failed to convert events: %w", err)
	}
	result.Events = events
	return nil
}

func applyShowObjectChanges(
	result *iotajsonrpc.IotaTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
	decodeEffects func() (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error),
	objectChangeNodes []iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsObjectChangesObjectChangeConnectionNodesObjectChange,
	senderAddress iotago.Address,
) error {
	if options == nil || !options.ShowObjectChanges {
		return nil
	}
	effects, err := decodeEffects()
	if err != nil {
		// Fall back to GraphQL nodes if BCS effects are unavailable
		objectChanges, convErr := convertGraphQLObjectChanges(objectChangeNodes)
		if convErr != nil {
			return fmt.Errorf("failed to convert object changes: %w", convErr)
		}
		result.ObjectChanges = objectChanges
		return nil
	}
	objectChanges, err := deriveObjectChangesFromEffects(effects, senderAddress)
	if err != nil {
		return fmt.Errorf("failed to convert object changes: %w", err)
	}
	result.ObjectChanges = objectChanges
	return nil
}

func applyShowBalanceChanges(
	result *iotajsonrpc.IotaTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
	balanceChangeNodes []iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsBalanceChangesBalanceChangeConnectionNodesBalanceChange,
) error {
	if options == nil || !options.ShowBalanceChanges {
		return nil
	}
	balanceChanges, err := convertGraphQLBalanceChanges(balanceChangeNodes)
	if err != nil {
		return fmt.Errorf("failed to convert balance changes: %w", err)
	}
	result.BalanceChanges = balanceChanges
	return nil
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

	var decodedEffects *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]
	decodeEffects := func() (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error) {
		if decodedEffects != nil {
			return decodedEffects, nil
		}
		effects, err := convertGraphQLEffects(&tx.Effects)
		if err != nil {
			return nil, err
		}
		decodedEffects = effects
		return effects, nil
	}

	if options != nil && options.ShowRawInput {
		result.RawTransaction = tx.Bcs
	}

	if err := applyShowEffects(result, options, decodeEffects); err != nil {
		return nil, err
	}

	if err := applyShowEvents(result, options, tx.Effects.Events.Nodes, digest); err != nil {
		return nil, err
	}

	timestampMs := tx.Effects.Timestamp.UnixMilli()
	result.TimestampMs = iotajsonrpc.NewBigInt(uint64(timestampMs)) // #nosec G115 -- timestamp is always positive

	result.Checkpoint = iotajsonrpc.NewBigInt(tx.Effects.Checkpoint.SequenceNumber)

	if err := applyShowObjectChanges(result, options, decodeEffects, tx.Effects.ObjectChanges.Nodes, tx.Sender.Address); err != nil {
		return nil, err
	}

	if err := applyShowBalanceChanges(result, options, tx.Effects.BalanceChanges.Nodes); err != nil {
		return nil, err
	}

	if options != nil && options.ShowRawEffects {
		result.RawEffects = tx.Effects.Bcs
	}

	return result, nil
}

func convertGraphQLEffects(
	effects *iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffects,
) (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error) {
	var decodedEffects iotajsonrpc.IotaTransactionBlockEffects
	if err := iotaclient.UnmarshalBCS(effects.Bcs, &decodedEffects); err != nil {
		return nil, fmt.Errorf("failed to decode BCS effects: %w", err)
	}

	return &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
		Data: decodedEffects,
	}, nil
}

//nolint:unparam // error return kept for API consistency
func convertGraphQLEvents(
	nodes []iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsEventsEventConnectionNodesEvent,
	txDigest *iotago.TransactionDigest,
) ([]*iotajsonrpc.IotaEvent, error) {
	events := make([]*iotajsonrpc.IotaEvent, 0, len(nodes))

	var digestValue iotago.TransactionDigest
	if txDigest != nil {
		digestValue = *txDigest
	}

	for i, node := range nodes {
		packageID := node.SendingModule.Package.Address
		module := node.SendingModule.Name
		sender := &node.Sender.Address

		// Parse event type from SendingModule if available
		// Note: We may need to extract the actual event type from the event data
		// For now, we'll construct a basic struct tag
		var eventType *iotago.StructTag
		// TODO: Extract proper event type from the event structure

		timestampMs := node.Timestamp.UnixMilli()

		event := &iotajsonrpc.IotaEvent{
			Id: iotajsonrpc.EventId{
				TxDigest: digestValue,
				EventSeq: iotajsonrpc.NewBigInt(uint64(i)), // #nosec G115
			},
			PackageId:         &packageID,
			TransactionModule: module,
			Sender:            sender,
			Type:              eventType,
			ParsedJson:        node.Json,
			Bcs:               iotago.Base64Data{},                        // TODO: Extract BCS if available
			TimestampMs:       iotajsonrpc.NewBigInt(uint64(timestampMs)), // #nosec G115 -- timestamp is always positive
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

//nolint:unparam // error return kept for API consistency
func convertGraphQLObjectChange(
	node *iotagraphql.RPC_TRANSACTION_FIELDSEffectsTransactionBlockEffectsObjectChangesObjectChangeConnectionNodesObjectChange,
) (*serialization.TagJson[iotajsonrpc.ObjectChange], error) {
	objectID := node.Address
	inputState := node.InputState
	outputState := node.OutputState

	objectType := ""
	if outputState.AsMoveObject.Contents.Type.Repr != "" {
		objectType = outputState.AsMoveObject.Contents.Type.Repr
	} else if inputState.AsMoveObject.Contents.Type.Repr != "" {
		objectType = inputState.AsMoveObject.Contents.Type.Repr
	}

	var change iotajsonrpc.ObjectChange
	switch {
	case len(outputState.AsMovePackage.Modules.Nodes) > 0:
		modules := make([]string, 0, len(outputState.AsMovePackage.Modules.Nodes))
		for _, module := range outputState.AsMovePackage.Modules.Nodes {
			modules = append(modules, module.Name)
		}
		change.Published = &struct {
			PackageID iotago.ObjectID     `json:"packageId"`
			Version   *iotajsonrpc.BigInt `json:"version"`
			Digest    iotago.ObjectDigest `json:"digest"`
			Nodules   []string            `json:"nodules"`
		}{
			PackageID: objectID,
			Version:   nil,
			Digest:    iotago.ObjectDigest{},
			Nodules:   modules,
		}
	case inputState.Version == 0 && objectType != "":
		change.Created = &struct {
			Sender     iotago.Address          `json:"sender"`
			Owner      iotajsonrpc.ObjectOwner `json:"owner"`
			ObjectType string                  `json:"objectType"`
			ObjectID   iotago.ObjectID         `json:"objectId"`
			Version    *iotajsonrpc.BigInt     `json:"version"`
			Digest     iotago.ObjectDigest     `json:"digest"`
		}{
			ObjectType: objectType,
			ObjectID:   objectID,
			Version:    nil,
			Digest:     iotago.ObjectDigest{},
		}
	case objectType != "":
		change.Mutated = &struct {
			Sender          iotago.Address          `json:"sender"`
			Owner           iotajsonrpc.ObjectOwner `json:"owner"`
			ObjectType      string                  `json:"objectType"`
			ObjectID        iotago.ObjectID         `json:"objectId"`
			Version         *iotajsonrpc.BigInt     `json:"version"`
			PreviousVersion *iotajsonrpc.BigInt     `json:"previousVersion"`
			Digest          iotago.ObjectDigest     `json:"digest"`
		}{
			ObjectType:      objectType,
			ObjectID:        objectID,
			PreviousVersion: iotajsonrpc.NewBigInt(inputState.Version),
			Digest:          iotago.ObjectDigest{},
		}
	default:
		if inputState.Version > 0 {
			change.Deleted = &struct {
				Sender     iotago.Address      `json:"sender"`
				ObjectType string              `json:"objectType"`
				ObjectID   iotago.ObjectID     `json:"objectId"`
				Version    *iotajsonrpc.BigInt `json:"version"`
			}{
				ObjectType: objectType,
				ObjectID:   objectID,
				Version:    iotajsonrpc.NewBigInt(inputState.Version),
			}
		}
	}

	return &serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change}, nil
}

func createMutatedChange(
	sender iotago.Address,
	ref iotajsonrpc.OwnedObjectRef,
	prevVersion *iotajsonrpc.BigInt,
) (*serialization.TagJson[iotajsonrpc.ObjectChange], error) {
	owner, err := convertOwnerFromTag(ref.Owner)
	if err != nil {
		return nil, err
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
			ObjectID:        *ref.Reference.ObjectID,
			Version:         iotajsonrpc.NewBigInt(ref.Reference.Version),
			PreviousVersion: prevVersion,
			Digest:          ref.Reference.Digest,
		},
	}
	return &serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change}, nil
}

func createCreatedChange(sender iotago.Address, ref iotajsonrpc.OwnedObjectRef) (*serialization.TagJson[iotajsonrpc.ObjectChange], error) {
	owner, err := convertOwnerFromTag(ref.Owner)
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
			ObjectID:   *ref.Reference.ObjectID,
			Version:    iotajsonrpc.NewBigInt(ref.Reference.Version),
			Digest:     ref.Reference.Digest,
		},
	}
	return &serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change}, nil
}

func createDeletedChange(sender iotago.Address, ref iotajsonrpc.IotaObjectRef) serialization.TagJson[iotajsonrpc.ObjectChange] {
	change := iotajsonrpc.ObjectChange{
		Deleted: &struct {
			Sender     iotago.Address      `json:"sender"`
			ObjectType string              `json:"objectType"`
			ObjectID   iotago.ObjectID     `json:"objectId"`
			Version    *iotajsonrpc.BigInt `json:"version"`
		}{
			Sender:     sender,
			ObjectType: "",
			ObjectID:   *ref.ObjectID,
			Version:    iotajsonrpc.NewBigInt(ref.Version),
		},
	}
	return serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change}
}

func createWrappedChange(sender iotago.Address, ref iotajsonrpc.IotaObjectRef) serialization.TagJson[iotajsonrpc.ObjectChange] {
	change := iotajsonrpc.ObjectChange{
		Wrapped: &struct {
			Sender     iotago.Address      `json:"sender"`
			ObjectType string              `json:"objectType"`
			ObjectID   iotago.ObjectID     `json:"objectId"`
			Version    *iotajsonrpc.BigInt `json:"version"`
		}{
			Sender:     sender,
			ObjectType: "",
			ObjectID:   *ref.ObjectID,
			Version:    iotajsonrpc.NewBigInt(ref.Version),
		},
	}
	return serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change}
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
		change, err := createMutatedChange(sender, ref, prevVersions[objectID])
		if err != nil {
			return err
		}
		changes = append(changes, *change)
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
		change, err := createCreatedChange(sender, created)
		if err != nil {
			return nil, err
		}
		changes = append(changes, *change)
	}

	for _, deleted := range v1.Deleted {
		if deleted.ObjectID == nil {
			continue
		}
		changes = append(changes, createDeletedChange(sender, deleted))
	}

	for _, wrapped := range v1.Wrapped {
		if wrapped.ObjectID == nil {
			continue
		}
		changes = append(changes, createWrappedChange(sender, wrapped))
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
		owner, err := convertGraphQLBalanceChangeOwner(node.Owner)
		if err != nil {
			return nil, fmt.Errorf("failed to convert balance change owner: %w", err)
		}

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
	if owner.AsAddress.Address != (iotago.Address{}) {
		addr := &owner.AsAddress.Address
		return &iotajsonrpc.ObjectOwner{
			ObjectOwnerInternal: &iotajsonrpc.ObjectOwnerInternal{
				AddressOwner: addr,
			},
		}, nil
	}

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

	if dryRunResult.Error != "" {
		return &iotajsonrpc.DevInspectResults{
			Error: dryRunResult.Error,
		}, nil
	}

	var effects *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]
	if len(dryRunResult.Transaction.Effects.Bcs) > 0 {
		var err error
		effects, err = convertGraphQLEffects(&dryRunResult.Transaction.Effects)
		if err != nil {
			// BCS decoding failed (possibly incomplete for dev inspect), use minimal effects
			effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
				Data: iotajsonrpc.IotaTransactionBlockEffects{
					V1: &iotajsonrpc.IotaTransactionBlockEffectsV1{
						Status: iotajsonrpc.ExecutionStatus{
							Status: iotajsonrpc.ExecutionStatusSuccess,
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
						Status: iotajsonrpc.ExecutionStatusSuccess,
					},
					GasUsed: iotajsonrpc.GasCostSummary{},
				},
			},
		}
	}

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

	var results []iotajsonrpc.ExecutionResultType
	for _, dryRunEffect := range dryRunResult.Results {
		executionResult := iotajsonrpc.ExecutionResultType{
			MutableReferenceOutputs: []iotajsonrpc.MutableReferenceOutputType{},
			ReturnValues:            []iotajsonrpc.ReturnValueType{},
		}

		for _, mutRef := range dryRunEffect.MutatedReferences {
			executionResult.MutableReferenceOutputs = append(executionResult.MutableReferenceOutputs, map[string]interface{}{
				"type": mutRef.Type.Repr,
				"bcs":  mutRef.Bcs,
			})
		}

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

	var effects *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]
	if len(dryRunResult.Transaction.Effects.Bcs) > 0 {
		var err error
		effects, err = convertGraphQLEffects(&dryRunResult.Transaction.Effects)
		if err != nil {
			// BCS decoding failed (possibly incomplete for dry run), use minimal effects
			effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
				Data: iotajsonrpc.IotaTransactionBlockEffects{
					V1: &iotajsonrpc.IotaTransactionBlockEffectsV1{
						Status: iotajsonrpc.ExecutionStatus{
							Status: iotajsonrpc.ExecutionStatusSuccess,
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
						Status: iotajsonrpc.ExecutionStatusSuccess,
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

	var input serialization.TagJson[iotajsonrpc.IotaTransactionBlockData]
	if len(dryRunResult.Transaction.Bcs) > 0 {
		var txData iotajsonrpc.IotaTransactionBlockData
		if err := iotaclient.UnmarshalBCS(dryRunResult.Transaction.Bcs, &txData); err != nil {
			return nil, fmt.Errorf("failed to decode input transaction: %w", err)
		}
		input = serialization.TagJson[iotajsonrpc.IotaTransactionBlockData]{
			Data: txData,
		}
	}

	var balanceChanges []iotajsonrpc.BalanceChange
	if len(dryRunResult.Transaction.Effects.BalanceChanges.Nodes) > 0 {
		convertedBalanceChanges, err := convertGraphQLBalanceChanges(dryRunResult.Transaction.Effects.BalanceChanges.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert balance changes: %w", err)
		}
		balanceChanges = convertedBalanceChanges
	}

	var objectChanges []serialization.TagJson[iotajsonrpc.ObjectChange]
	derivedChanges, err := deriveObjectChangesFromEffects(effects, dryRunResult.Transaction.Sender.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to derive object changes: %w", err)
	}

	if len(derivedChanges) > 0 {
		objectChanges = derivedChanges
	} else if len(dryRunResult.Transaction.Effects.ObjectChanges.Nodes) > 0 {
		convertedObjectChanges, err := convertGraphQLObjectChanges(dryRunResult.Transaction.Effects.ObjectChanges.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object changes: %w", err)
		}
		objectChanges = convertedObjectChanges
	}

	return &iotajsonrpc.DryRunTransactionBlockResponse{
		Effects:        *effects,
		Events:         events,
		ObjectChanges:  objectChanges,
		BalanceChanges: balanceChanges,
		Input:          input,
	}, nil
}

func applyExecuteShowEffects(
	result *iotajsonrpc.IotaTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
	decodeEffects func() (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error),
) {
	showEffects := true
	if options != nil {
		showEffects = options.ShowEffects
	}
	if !showEffects {
		return
	}
	effects, err := decodeEffects()
	if err != nil {
		result.Effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{
			Data: iotajsonrpc.IotaTransactionBlockEffects{
				V1: &iotajsonrpc.IotaTransactionBlockEffectsV1{
					Status:  iotajsonrpc.ExecutionStatus{Status: iotajsonrpc.ExecutionStatusSuccess},
					GasUsed: iotajsonrpc.GasCostSummary{},
				},
			},
		}
	} else {
		result.Effects = effects
	}
}

func applyExecuteShowRawEffects(
	result *iotajsonrpc.IotaTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
	bcs iotago.Base64Data,
) {
	showRawEffects := true
	if options != nil {
		showRawEffects = options.ShowRawEffects
	}
	if showRawEffects {
		result.RawEffects = bcs
	}
}

func applyExecuteTransactionOptions(
	result *iotajsonrpc.IotaTransactionBlockResponse,
	txBlock *iotagraphql.RPC_TRANSACTION_FIELDS,
	digest *iotago.Digest,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) error {
	var decodedEffects *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]
	decodeEffects := func() (*serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects], error) {
		if decodedEffects != nil {
			return decodedEffects, nil
		}
		effects, err := convertGraphQLEffects(&txBlock.Effects)
		if err != nil {
			return nil, err
		}
		decodedEffects = effects
		return effects, nil
	}

	if options != nil && options.ShowRawInput {
		result.RawTransaction = txBlock.Bcs
	}

	applyExecuteShowEffects(result, options, decodeEffects)

	if err := applyShowEvents(result, options, txBlock.Effects.Events.Nodes, digest); err != nil {
		return err
	}

	if err := applyShowObjectChanges(result, options, decodeEffects, txBlock.Effects.ObjectChanges.Nodes, txBlock.Sender.Address); err != nil {
		return err
	}

	if err := applyShowBalanceChanges(result, options, txBlock.Effects.BalanceChanges.Nodes); err != nil {
		return err
	}

	applyExecuteShowRawEffects(result, options, txBlock.Effects.Bcs)

	return nil
}

func convertExecuteTransactionBlockResponse(
	resp *iotagraphql.ExecuteTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	if len(resp.ExecuteTransactionBlock.Errors) > 0 {
		return nil, fmt.Errorf("execution failed: %v", resp.ExecuteTransactionBlock.Errors)
	}

	txBlock := &resp.ExecuteTransactionBlock.Effects.TransactionBlock.RPC_TRANSACTION_FIELDS

	digest, err := iotago.NewDigest(txBlock.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction digest: %w", err)
	}

	result := &iotajsonrpc.IotaTransactionBlockResponse{Digest: *digest}

	// #nosec G115 -- timestamps from blockchain are always positive
	result.TimestampMs = iotajsonrpc.NewBigInt(uint64(txBlock.Effects.Timestamp.UnixMilli()))
	result.Checkpoint = iotajsonrpc.NewBigInt(txBlock.Effects.Checkpoint.SequenceNumber)

	if err := applyExecuteTransactionOptions(result, txBlock, digest, options); err != nil {
		return nil, err
	}

	return result, nil
}

func applyRPCMoveObjectFieldsOptions(
	data *iotajsonrpc.IotaObjectData,
	fields *iotagraphql.RPC_MOVE_OBJECT_FIELDS,
	options *iotajsonrpc.IotaObjectDataOptions,
) error {
	if options == nil {
		return nil
	}

	if options.ShowType {
		typeStr := fields.Contents_type.Type.Repr
		data.Type = &typeStr
	}

	if options.ShowContent {
		parsedContent := serialization.TagJson[iotajsonrpc.IotaParsedData]{
			Data: iotajsonrpc.IotaParsedData{
				MoveObject: &iotajsonrpc.IotaParsedMoveObject{
					Type:              fields.Contents_content.Type.Repr,
					HasPublicTransfer: true,
					Fields:            fields.Contents_content.Data,
				},
			},
		}
		data.Content = &parsedContent
	}

	if options.ShowBcs {
		structTag, err := iotago.StructTagFromString(fields.Contents.Type.Repr)
		if err != nil {
			return fmt.Errorf("failed to parse struct tag: %w", err)
		}
		rawData := serialization.TagJson[iotajsonrpc.IotaRawData]{
			Data: iotajsonrpc.IotaRawData{
				MoveObject: &iotajsonrpc.IotaRawMoveObject{
					Type:              *structTag,
					HasPublicTransfer: true,
					Version:           fields.Version,
					BcsBytes:          fields.Bcs,
				},
			},
		}
		data.Bcs = &rawData
	}

	if options.ShowOwner {
		owner, err := convertGraphQLObjectOwner(fields.Owner)
		if err != nil {
			return fmt.Errorf("failed to convert owner: %w", err)
		}
		data.Owner = owner
	}

	if options.ShowPreviousTransaction {
		txDigest, err := iotago.NewDigest(fields.PreviousTransactionBlock.Digest)
		if err != nil {
			return fmt.Errorf("failed to parse transaction digest: %w", err)
		}
		data.PreviousTransaction = txDigest
	}

	if options.ShowStorageRebate {
		data.StorageRebate = fields.StorageRebate.Clone()
	}

	if options.ShowDisplay && len(fields.Display) > 0 {
		display := make(map[string]string)
		for _, entry := range fields.Display {
			display[entry.Key] = entry.Value
		}
		data.Display = display
	}

	return nil
}

func convertRPCMoveObjectFieldsToIotaObjectResponse(
	fields *iotagraphql.RPC_MOVE_OBJECT_FIELDS,
	options *iotajsonrpc.IotaObjectDataOptions,
) (*iotajsonrpc.IotaObjectResponse, error) {
	if fields == nil {
		return nil, fmt.Errorf("fields is nil")
	}

	digest, err := iotago.NewDigest(fields.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse object digest: %w", err)
	}

	data := &iotajsonrpc.IotaObjectData{
		ObjectID: &fields.ObjectId,
		Version:  iotajsonrpc.NewBigInt(fields.Version),
		Digest:   digest,
	}

	if err := applyRPCMoveObjectFieldsOptions(data, fields, options); err != nil {
		return nil, err
	}

	return &iotajsonrpc.IotaObjectResponse{Data: data}, nil
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
		initialVersion := o.InitialSharedVersion
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
