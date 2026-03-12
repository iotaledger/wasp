// Package iotagraphql provides a GraphQL client for interacting with IOTA nodes.
package iotagraphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/gorilla/websocket"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

const (
	SingleCoinFundsFromFaucetAmount = uint64(2_000_000_000)
	FundsFromFaucetAmount           = SingleCoinFundsFromFaucetAmount * 5
)

type GraphQLClient struct {
	url                     string
	faucetURL               string
	client                  graphql.Client
	httpClient              *http.Client
	WaitUntilEffectsVisible *WaitParams
	FaucetRetryParams       *WaitParams
	tickingTime             time.Duration
	log                     log.Logger
}

// newWebSocketClient creates a new WebSocket client, dials the connection, and returns it ready for subscriptions.
func (c *GraphQLClient) newWebSocketClient(ctx context.Context) (graphql.WebSocketClient, error) {
	url := c.url + "/subscriptions"
	c.log.LogDebugf("dialing WebSocket connection to %s", url)
	wsClient := graphql.NewClientUsingWebSocket(url, &WebSocketDialer{log: c.log})
	if _, err := wsClient.Start(ctx); err != nil {
		return nil, err
	}
	return wsClient, nil
}

func NewGraphQLClient(url, faucetURL string) *GraphQLClient {
	return NewGraphQLClientWithTimeout(url, faucetURL, 30*time.Second, nil)
}

func NewGraphQLClientWithWaitParams(url string, faucetURL string, waitParams *WaitParams) *GraphQLClient {
	return NewGraphQLClientWithTimeout(url, faucetURL, 30*time.Second, waitParams)
}

type WebSocketDialer struct {
	websocket.Dialer
	log log.Logger
}

func (w *WebSocketDialer) DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (graphql.WSConn, error) {
	conn, resp, err := w.Dialer.DialContext(ctx, urlStr, requestHeader)
	if err != nil && resp != nil {
		w.log.LogErrorf("dialing WebSocket failed: url=%s status=%d", urlStr, resp.StatusCode)
	}
	return conn, err
}

func NewGraphQLClientWithTimeout(url, faucetURL string, timeout time.Duration, waitParams *WaitParams) *GraphQLClient {
	httpClient := &http.Client{

		Timeout: timeout,
	}

	return &GraphQLClient{
		url:                     strings.TrimRight(url, "/"),
		faucetURL:               faucetURL,
		client:                  graphql.NewClient(url, httpClient),
		httpClient:              httpClient,
		WaitUntilEffectsVisible: waitParams,
		tickingTime:             250 * time.Millisecond,
		log:                     log.EmptyLogger,
	}
}

func (c *GraphQLClient) WithLogger(logger log.Logger) *GraphQLClient {
	c.log = logger
	return c
}

// RequestFundsFromFaucet requests test funds for the provided address from the faucet endpoint.
// If FaucetRetryParams is configured, it waits for the coins to be visible on the ledger.
func (c *GraphQLClient) RequestFundsFromFaucet(ctx context.Context, address iotago.Address) error {
	params := c.FaucetRetryParams
	if params == nil {
		params = c.WaitUntilEffectsVisible
	}
	if params == nil {
		params = &WaitParams{
			Attempts:             20,
			DelayBetweenAttempts: 500 * time.Millisecond,
		}
	}

	initial, err := c.getIotaBalanceSnapshot(ctx, address)
	for i := 0; err != nil && i < params.Attempts; i++ {
		if waitErr := waitWithContext(ctx, 1, params.DelayBetweenAttempts); waitErr != nil {
			return waitErr
		}
		initial, err = c.getIotaBalanceSnapshot(ctx, address)
	}
	if err != nil {
		return fmt.Errorf("failed to get initial balance before faucet request: %w", err)
	}

	if err := requestFundsFromFaucetRaw(ctx, address, c.faucetURL); err != nil {
		return err
	}

	for i := 0; i < params.Attempts; i++ {
		current, err := c.getIotaBalanceSnapshot(ctx, address)
		if err == nil && (current.Total.Cmp(initial.Total) > 0 || current.CoinObjectCount > initial.CoinObjectCount) {
			return nil
		}
		if i < params.Attempts-1 {
			if waitErr := waitWithContext(ctx, 1, params.DelayBetweenAttempts); waitErr != nil {
				return waitErr
			}
		}
	}

	return fmt.Errorf("timeout waiting for faucet coins to be visible")
}

type iotaBalanceSnapshot struct {
	Total           *big.Int
	CoinObjectCount uint64
}

func (c *GraphQLClient) getIotaBalanceSnapshot(ctx context.Context, address iotago.Address) (iotaBalanceSnapshot, error) {
	balance, err := c.GetBalance(ctx, GetBalanceRequest{Owner: &address})
	if err != nil {
		return iotaBalanceSnapshot{}, err
	}
	if balance == nil || balance.TotalBalance == nil || balance.TotalBalance.Int == nil {
		return iotaBalanceSnapshot{}, fmt.Errorf("balance is nil")
	}

	total := new(big.Int).Set(balance.TotalBalance.Int)
	coinObjectCount := uint64(0)
	if balance.CoinObjectCount != nil {
		coinObjectCount = balance.CoinObjectCount.Uint64()
	}

	return iotaBalanceSnapshot{
		Total:           total,
		CoinObjectCount: coinObjectCount,
	}, nil
}

func waitWithContext(ctx context.Context, attempts int, delay time.Duration) error {
	for i := 0; i < attempts; i++ {
		if delay <= 0 {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil
}

func requestFundsFromFaucetRaw(ctx context.Context, address iotago.Address, faucetURL string) error {
	payload := map[string]any{
		"FixedAmountRequest": map[string]string{
			"recipient": address.String(),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal faucet request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, faucetURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create faucet request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("faucet request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("faucet returned unexpected status %s", res.Status)
	}

	var parsed struct {
		Error any `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		// The response body is informational; don't fail just because decoding failed.
		return nil
	}

	switch v := parsed.Error.(type) {
	case nil:
		return nil
	case string:
		if v == "" {
			return nil
		}
		return fmt.Errorf("faucet error: %s", v)
	default:
		return fmt.Errorf("faucet returned an error")
	}
}

func (c *GraphQLClient) Query(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func bigIntToUint64(b *BigInt, fieldName string) (uint64, error) {
	if b == nil {
		return 0, fmt.Errorf("%s is nil", fieldName)
	}
	if !b.IsUint64() {
		return 0, fmt.Errorf("%s value %s exceeds uint64 maximum", fieldName, b.String())
	}
	return b.Uint64(), nil
}

func validateRequired(val interface{}, paramName string) error {
	if val == nil {
		return fmt.Errorf("%s is required", paramName)
	}
	return nil
}

func (c *GraphQLClient) GetDynamicFieldObject(
	ctx context.Context,
	req GetDynamicFieldObjectRequest,
) (*GetDynamicFieldObjectResponse, error) {
	if req.ParentObjectID == nil {
		return nil, fmt.Errorf("parent object ID is required")
	}
	if req.Name == nil {
		return nil, fmt.Errorf("dynamic field name is required")
	}

	valueJSON, err := json.Marshal(req.Name.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value to JSON: %w", err)
	}

	bcsData, err := bcs.Marshal(&valueJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to BCS-encode value: %w", err)
	}

	nameInput := graphqltypes.DynamicFieldName{
		Type: req.Name.Type,
		Bcs:  bcsData,
	}

	showBcs := true
	showPreviousTransaction := true
	showDisplay := true
	showStorageRebate := true

	objResp, objErr := graphqltypes.GetDynamicFieldObject(ctx, c.client, *req.ParentObjectID, nameInput,
		&showBcs, &showPreviousTransaction, &showDisplay, &showStorageRebate)

	return objResp, objErr
}

func (c *GraphQLClient) GetDynamicFields(
	ctx context.Context,
	req GetDynamicFieldsRequest,
) (*graphqltypes.GetDynamicFieldsResponse, error) {
	return graphqltypes.GetDynamicFields(ctx, c.client, *req.ParentObjectID, nil, req.Cursor)
}

func (c *GraphQLClient) GetOwnedObjects(
	ctx context.Context,
	req GetOwnedObjectsRequest,
) (*graphqltypes.GetOwnedObjectsResponse, error) {
	if err := validateRequired(req.Address, "address"); err != nil {
		return nil, err
	}

	filter := req.Filter
	opts := showAllObjectOptions()

	resp, err := graphqltypes.GetOwnedObjects(ctx, c.client, *req.Address, req.Limit, req.Cursor,
		opts.Bcs, opts.Content, opts.Display, opts.Type, opts.Owner, opts.PreviousTransaction, opts.StorageRebate, filter)

	return resp, err
}

func (c *GraphQLClient) DryRunTransaction(
	ctx context.Context,
	txDataBytes iotago.Base64Data,
) (*graphqltypes.DryRunTransactionBlockResponse, error) {
	txBytes := txDataBytes.String()

	resp, err := graphqltypes.DryRunTransactionBlock(ctx, c.client, txBytes)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GraphQLClient) ExecuteTransactionBlock(
	ctx context.Context,
	txDataBytes iotago.Base64Data,
	signatures []*iotasigner.Signature,
) (*graphqltypes.ExecuteTransactionBlockResponse, error) {
	if len(signatures) == 0 {
		return nil, fmt.Errorf("at least one signature is required")
	}
	txBytes := txDataBytes.String()
	sigStrings := make([]string, len(signatures))
	for i, sig := range signatures {
		sigBytes := sig.Bytes()
		if sigBytes == nil {
			return nil, fmt.Errorf("signature %d has nil bytes", i)
		}
		sigStrings[i] = iotago.Base64Data(sigBytes).String()
	}

	resp, err := graphqltypes.ExecuteTransactionBlock(ctx, c.client, txBytes, sigStrings)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GraphQLClient) GetLatestIotaSystemState(ctx context.Context) (*GetLatestIotaSystemStateResponse, error) {
	resp, err := graphqltypes.GetLatestIotaSystemState(ctx, c.client)
	return resp, err
}

func (c *GraphQLClient) GetReferenceGasPrice(ctx context.Context) (*BigInt, error) {
	resp, err := graphqltypes.GetReferenceGasPrice(ctx, c.client)
	if err != nil {
		return nil, err
	}
	return resp.Epoch.ReferenceGasPrice.Clone(), nil
}

func (c *GraphQLClient) MergeCoins(
	ctx context.Context,
	req MergeCoinsRequest,
) (*TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "MergeCoins")
}

func (c *GraphQLClient) fetchObjectRefs(ctx context.Context, objectIDs []iotago.ObjectID) ([]*iotago.ObjectRef, error) {
	refs := make([]*iotago.ObjectRef, 0, len(objectIDs))
	for _, objID := range objectIDs {
		objResp, err := c.GetObject(ctx, objID)
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", objID.String(), err)
		}
		if objResp.Object.IsNotFound() {
			return nil, fmt.Errorf("object %s not found", objID.String())
		}
		ref, err := objResp.Object.ObjectRef()
		if err != nil {
			return nil, fmt.Errorf("failed to get object ref for %s: %w", objID.String(), err)
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func (c *GraphQLClient) Pay(
	ctx context.Context,
	req PayRequest,
) (*TransactionBytes, error) {
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

	gasBudget := uint64(DefaultGasBudget)
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

	gasObj, err := c.GetObject(ctx, *req.Gas)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas object %s: %w", req.Gas.String(), err)
	}
	if gasObj.Object.IsNotFound() {
		return nil, fmt.Errorf("gas object %s not found", req.Gas.String())
	}

	gasRef, err := gasObj.Object.ObjectRef()
	if err != nil {
		return nil, fmt.Errorf("failed to get gas object ref %s: %w", req.Gas.String(), err)
	}
	gasPayment := []*iotago.ObjectRef{gasRef}

	tx := iotago.NewProgrammable(
		req.Signer,
		pt,
		gasPayment,
		gasBudget,
		DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	return &TransactionBytes{TxBytes: txBytes}, nil
}

func (c *GraphQLClient) PayAllIota(
	ctx context.Context,
	req PayAllIotaRequest,
) (*TransactionBytes, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	if err := ptb.PayAllIota(req.Recipient); err != nil {
		return nil, fmt.Errorf("failed to build PayAllIota transaction: %w", err)
	}
	pt := ptb.Finish()

	var err error
	gasBudget := uint64(DefaultGasBudget)
	if req.GasBudget != nil {
		gasBudget, err = bigIntToUint64(req.GasBudget, "gasBudget")
		if err != nil {
			return nil, err
		}
	}

	gasPayment := make([]*iotago.ObjectRef, 0, len(req.InputCoins))

	for _, coinID := range req.InputCoins {
		var objResp *graphqltypes.GetObjectResponse
		objResp, err = c.GetObject(ctx, coinID)
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", coinID.String(), err)
		}
		if objResp.Object.IsNotFound() {
			return nil, fmt.Errorf("object %s not found", coinID.String())
		}

		var objRef *iotago.ObjectRef
		objRef, err = objResp.Object.ObjectRef()
		if err != nil {
			return nil, fmt.Errorf("failed to get ref for %s: %w", coinID.String(), err)
		}
		gasPayment = append(gasPayment, objRef)
	}

	tx := iotago.NewProgrammable(
		req.Signer,
		pt,
		gasPayment,
		gasBudget,
		DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	return &TransactionBytes{TxBytes: txBytes}, nil
}

func (c *GraphQLClient) PayIota(
	ctx context.Context,
	req PayIotaRequest,
) (*TransactionBytes, error) {
	coinRefs, err := c.fetchObjectRefs(ctx, req.InputCoins)
	if err != nil {
		return nil, err
	}
	if len(req.Recipients) != len(req.Amount) {
		return nil, fmt.Errorf("recipients and amounts mismatch. Got %d recipients but %d amounts", len(req.Recipients), len(req.Amount))
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
	if err = ptb.PayIota(req.Recipients, amounts); err != nil {
		return nil, fmt.Errorf("failed to build PayIota transaction: %w", err)
	}
	pt := ptb.Finish()

	gasBudget := uint64(DefaultGasBudget)
	if req.GasBudget != nil {
		gasBudget, err = bigIntToUint64(req.GasBudget, "gasBudget")
		if err != nil {
			return nil, err
		}
	}

	tx := iotago.NewProgrammable(
		req.Signer,
		pt,
		coinRefs,
		gasBudget,
		DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	return &TransactionBytes{TxBytes: txBytes}, nil
}

func (c *GraphQLClient) Publish(
	ctx context.Context,
	req PublishRequest,
) (*TransactionBytes, error) {
	if req.Sender == nil {
		return nil, fmt.Errorf("Publish: sender address is required")
	}
	if len(req.CompiledModules) == 0 {
		return nil, fmt.Errorf("Publish: at least one compiled module is required")
	}

	modules := make([][]byte, len(req.CompiledModules))
	for i, module := range req.CompiledModules {
		if module == nil {
			return nil, fmt.Errorf("Publish: compiled module at index %d is nil", i)
		}
		modules[i] = module.Data()
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	capArg := ptb.PublishUpgradeable(modules, req.Dependencies)
	ptb.TransferArgs(req.Sender, []iotago.Argument{capArg})
	pt := ptb.Finish()

	gasBudget := uint64(DefaultGasBudget)
	if req.GasBudget != nil {
		var err error
		gasBudget, err = bigIntToUint64(req.GasBudget, "gasBudget")
		if err != nil {
			return nil, err
		}
	}

	gasRef, err := c.resolveGasObject(ctx, req.Sender, req.Gas, nil)
	if err != nil {
		return nil, err
	}
	gasPayment := []*iotago.ObjectRef{gasRef}

	tx := iotago.NewProgrammable(
		req.Sender,
		pt,
		gasPayment,
		gasBudget,
		DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	return &TransactionBytes{TxBytes: txBytes}, nil
}

func (c *GraphQLClient) loadObjectRef(ctx context.Context, objectID *iotago.ObjectID) (*iotago.ObjectRef, error) {
	if objectID == nil {
		return nil, fmt.Errorf("object ID is nil")
	}
	objResp, err := c.GetObject(ctx, *objectID)
	if err != nil {
		return nil, err
	}
	if objResp.Object.IsNotFound() {
		return nil, fmt.Errorf("object %s not found", objectID.String())
	}
	return objResp.Object.ObjectRef()
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
		coins, err := c.GetCoins(ctx, GetCoinsRequest{
			Owner:  signer,
			Limit:  pageLimit,
			Cursor: cursor,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch coins for gas selection: %w", err)
		}
		for _, coin := range coins.Address.Coins.Nodes {
			if transferObjectID != nil && coin.Address == *transferObjectID {
				continue
			}
			return coin.ObjectRef()
		}
		if !coins.Address.Coins.PageInfo.HasNextPage || coins.Address.Coins.PageInfo.EndCursor == "" {
			break
		}
		cursor = &coins.Address.Coins.PageInfo.EndCursor
	}
	return nil, fmt.Errorf("no suitable gas coin found; provide Gas explicitly")
}

func (c *GraphQLClient) TransferIota(
	ctx context.Context,
	req TransferIotaRequest,
) (*TransactionBytes, error) {
	return nil, fmt.Errorf("not implemented: %s", "TransferIota")
}

func (c *GraphQLClient) TransferObject(
	ctx context.Context,
	req TransferObjectRequest,
) (*TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("TransferObject: signer address is required")
	}
	if req.ObjectID == nil {
		return nil, fmt.Errorf("TransferObject: object ID is required")
	}
	if req.Recipient == nil {
		return nil, fmt.Errorf("TransferObject: recipient address is required")
	}

	objResp, err := c.GetObject(ctx, *req.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("TransferObject: failed to get object: %w", err)
	}
	objRef, err := objResp.Object.ObjectRef()
	if err != nil {
		return nil, fmt.Errorf("TransferObject: failed to get object ref: %w", err)
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	objArg := ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: objRef})
	ptb.TransferArgs(req.Recipient, []iotago.Argument{objArg})
	pt := ptb.Finish()

	gasBudget := uint64(DefaultGasBudget)
	if req.GasBudget != nil {
		gasBudget, err = bigIntToUint64(req.GasBudget, "gasBudget")
		if err != nil {
			return nil, err
		}
	}

	gasRef, err := c.resolveGasObject(ctx, req.Signer, req.Gas, nil)
	if err != nil {
		return nil, err
	}

	tx := iotago.NewProgrammable(
		req.Signer,
		pt,
		[]*iotago.ObjectRef{gasRef},
		gasBudget,
		DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("TransferObject: failed to marshal transaction: %w", err)
	}

	return &TransactionBytes{TxBytes: txBytes}, nil
}

func (c *GraphQLClient) GetCoinObjsForTargetAmount(
	ctx context.Context,
	address iotago.Address,
	targetAmount uint64,
	gasAmount uint64,
) (Coins, error) {
	coins, err := c.GetCoins(
		ctx, GetCoinsRequest{
			Owner: &address,
			Limit: 50,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call GetCoins(): %w", err)
	}
	pickedCoins, err := PickupCoins(Coins(coins.Address.Coins.Nodes), new(big.Int).SetUint64(targetAmount), gasAmount, 0, 25)
	if err != nil {
		return nil, err
	}
	return pickedCoins.Coins, nil
}

func (c *GraphQLClient) SignAndExecuteTransaction(
	ctx context.Context,
	txnBytes []byte,
	signer iotasigner.Signer,
) (*graphqltypes.ExecuteTransactionBlockResponse, error) {
	signature, err := signer.SignTransactionBlock(txnBytes, iotasigner.DefaultIntent())
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction block: %w", err)
	}
	resp, err := c.ExecuteTransactionBlock(ctx, txnBytes, []*iotasigner.Signature{signature})
	if err != nil {
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	txDigest := resp.ExecuteTransactionBlock.Effects.TransactionBlock.Digest
	if err := c.waitForEffectsIndexed(ctx, txDigest); err != nil {
		return resp, fmt.Errorf("transaction succeeded but effects not yet indexed: %w", err)
	}

	return resp, nil
}

func (c *GraphQLClient) waitForEffectsIndexed(ctx context.Context, txDigest string) error {
	params := c.WaitUntilEffectsVisible
	if params == nil {
		params = WaitForEffectsEnabled
	}

	var objectChanges []graphqltypes.ObjectChangeData
	for i := range params.Attempts {
		res, err := graphqltypes.GetTransactionBlock(ctx, c.client, txDigest)
		if err == nil &&
			res.TransactionBlock.Effects.Checkpoint.SequenceNumber > 0 &&
			res.TransactionBlock.Effects.GasEffects.GasObject.Digest != "" {
			objectChanges = res.TransactionBlock.Effects.ObjectChanges.Nodes
			break
		}

		if i == params.Attempts-1 {
			return fmt.Errorf("transaction %s not indexed after %d attempts", txDigest, params.Attempts)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(params.DelayBetweenAttempts):
		}
	}

	for _, change := range objectChanges {
		if change.IdDeleted || change.OutputState.Digest == "" {
			continue
		}

		objectID := change.Address
		targetVersion := change.OutputState.Version

		if err := c.waitForObjectAtVersion(ctx, objectID, targetVersion, params); err != nil {
			return err
		}
	}

	return nil
}

func (c *GraphQLClient) waitForObjectAtVersion(
	ctx context.Context,
	objectID iotago.ObjectID,
	minVersion uint64,
	params *WaitParams,
) error {
	for i := range params.Attempts {
		showNone := false
		res, err := graphqltypes.GetObject(ctx, c.client, objectID,
			&showNone, &showNone, &showNone, &showNone, &showNone, &showNone, &showNone)
		if err == nil && !res.Object.IsNotFound() && res.Object.Version >= minVersion {
			return nil
		}

		if i == params.Attempts-1 {
			return fmt.Errorf("object %s did not reach version %d after %d attempts",
				objectID, minVersion, params.Attempts)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(params.DelayBetweenAttempts):
		}
	}
	return nil // unreachable
}

func (c *GraphQLClient) UpdateObjectRef(
	ctx context.Context,
	ref *iotago.ObjectRef,
) (*iotago.ObjectRef, error) {
	res, err := c.GetObject(ctx, *ref.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the object of ObjectRef: %w", err)
	}

	return res.Object.ObjectRef()
}

func (c *GraphQLClient) MintToken(
	ctx context.Context,
	signer iotasigner.Signer,
	packageID iotago.PackageID,
	tokenName string,
	treasuryCap *iotago.ObjectRef,
	mintAmount uint64,
	maxRetries int,
) (*graphqltypes.ExecuteTransactionBlockResponse, error) {
	var err error
	var txnBytes []byte
	var txnResponse *graphqltypes.ExecuteTransactionBlockResponse
	var gasPayments []*iotago.ObjectRef

	for i := 0; i < maxRetries; i++ {
		updatedTreasuryCap, updateErr := c.UpdateObjectRef(ctx, treasuryCap)
		if updateErr != nil {
			return nil, fmt.Errorf("failed to update treasuryCap: %w", updateErr)
		}
		if updatedTreasuryCap.Version > treasuryCap.Version {
			treasuryCap = updatedTreasuryCap
		}

		ptb := iotago.NewProgrammableTransactionBuilder()
		ptb.Command(
			iotago.Command{
				MoveCall: &iotago.ProgrammableMoveCall{
					Package:       &packageID,
					Module:        tokenName,
					Function:      "mint",
					TypeArguments: []iotago.TypeTag{},
					Arguments: []iotago.Argument{
						ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: treasuryCap}),
						ptb.MustForceSeparatePure(mintAmount),
						ptb.MustForceSeparatePure(signer.Address()),
					},
				},
			},
		)
		pt := ptb.Finish()

		gasPayments, err = c.FindCoinsForGasPayment(ctx, signer.Address(), pt, DefaultGasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to find gas payment: %w", err)
		}

		tx := iotago.NewProgrammable(
			signer.Address(),
			pt,
			gasPayments,
			DefaultGasBudget,
			DefaultGasPrice,
		)
		txnBytes, err = bcs.Marshal(&tx)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tx: %w", err)
		}

		txnResponse, err = c.SignAndExecuteTransaction(ctx, txnBytes, signer)
		if err == nil {
			return txnResponse, nil
		}
		time.Sleep(c.tickingTime)
	}
	return nil, fmt.Errorf("can't execute MintToken in time: %w", err)
}

func (c *GraphQLClient) GetAllBalances(ctx context.Context, owner iotago.Address) ([]*Balance, error) {
	resp, err := graphqltypes.GetAllBalances(ctx, c.client, owner, nil, nil)
	if err != nil {
		return nil, err
	}
	balances := make([]*Balance, 0, len(resp.Address.Balances.Nodes))
	for _, node := range resp.Address.Balances.Nodes {
		bal, err := convertGraphQLBalance(node.CoinType.Repr, node.CoinObjectCount, node.TotalBalance)
		if err != nil {
			return nil, err
		}
		balances = append(balances, bal)
	}
	return balances, nil
}

func (c *GraphQLClient) GetAllCoins(ctx context.Context, req GetAllCoinsRequest) (*GetAllCoinsResponse, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("owner address is required")
	}

	var limitPtr *int
	if req.Limit > 0 {
		limitPtr = &req.Limit
	}

	resp, err := graphqltypes.GetAllCoins(ctx, c.client, *req.Owner, limitPtr, req.Cursor)
	return resp, err
}

func (c *GraphQLClient) GetBalance(ctx context.Context, req GetBalanceRequest) (*Balance, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("owner address is required")
	}
	var coinTypePtr *string
	if req.CoinType != "" {
		s := string(req.CoinType)
		coinTypePtr = &s
	}
	resp, err := graphqltypes.GetBalance(ctx, c.client, *req.Owner, coinTypePtr)
	if err != nil {
		return nil, err
	}
	balance := resp.Address.Balance
	return convertGraphQLBalance(balance.CoinType.Repr, balance.CoinObjectCount, balance.TotalBalance)
}

func (c *GraphQLClient) GetCoinMetadata(ctx context.Context, coinType CoinType) (*IotaCoinMetadata, error) {
	if coinType == "" {
		return nil, fmt.Errorf("coin type is required")
	}
	resp, err := graphqltypes.GetCoinMetadata(ctx, c.client, string(coinType))
	if err != nil {
		return nil, err
	}
	meta := resp.CoinMetadata
	objID := &meta.Address
	return &IotaCoinMetadata{
		Name:        meta.Name,
		Symbol:      meta.Symbol,
		Decimals:    uint8(meta.Decimals), // #nosec G115 -- decimals is always < 256
		Description: meta.Description,
		IconURL:     meta.IconUrl,
		ID:          objID,
	}, nil
}

func (c *GraphQLClient) GetCoins(ctx context.Context, req GetCoinsRequest) (*graphqltypes.GetCoinsResponse, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("owner address is required")
	}

	var limitPtr *int
	if req.Limit > 0 {
		limitPtr = &req.Limit
	}

	var coinTypePtr *string
	if req.CoinType != nil {
		s := string(*req.CoinType)
		coinTypePtr = &s
	}

	resp, err := graphqltypes.GetCoins(ctx, c.client, *req.Owner, limitPtr, req.Cursor, coinTypePtr)
	return resp, err
}

func (c *GraphQLClient) GetTotalSupply(ctx context.Context, coinType CoinType) (*Supply, error) {
	resp, err := graphqltypes.GetLatestIotaSystemState(ctx, c.client)
	if err != nil {
		return nil, err
	}

	return &Supply{Value: resp.Epoch.IotaTotalSupply.Clone()}, err
}

type objectShowOptions struct {
	Bcs, Owner, PreviousTransaction, Content, Display, Type, StorageRebate *bool
}

func showAllObjectOptions() objectShowOptions {
	t := true
	return objectShowOptions{&t, &t, &t, &t, &t, &t, &t}
}

func (c *GraphQLClient) GetObject(ctx context.Context, objectID iotago.ObjectID) (*graphqltypes.GetObjectResponse, error) {
	opts := showAllObjectOptions()

	return Retry(
		ctx,
		func() (*graphqltypes.GetObjectResponse, error) {
			return graphqltypes.GetObject(ctx, c.client, objectID,
				opts.Bcs, opts.Owner, opts.PreviousTransaction, opts.Content, opts.Display, opts.Type, opts.StorageRebate)
		},
		func(resp *graphqltypes.GetObjectResponse, err error) bool {
			return resp != nil && resp.Object.IsNotFound()
		},
		c.WaitUntilEffectsVisible,
	)
}

func (c *GraphQLClient) GetTransactionBlock(ctx context.Context, digest iotago.TransactionDigest) (*graphqltypes.GetTransactionBlockResponse, error) {
	return graphqltypes.GetTransactionBlock(ctx, c.client, digest.String())
}

func (c *GraphQLClient) TryGetPastObject(
	ctx context.Context,
	objectID iotago.ObjectID,
	version uint64,
) (*TryGetPastObjectResponse, error) {
	opts := showAllObjectOptions()

	return graphqltypes.TryGetPastObject(ctx, c.client, objectID, &version,
		opts.Bcs, opts.Owner, opts.PreviousTransaction, opts.Content, opts.Display, opts.Type, opts.StorageRebate)
}

func (c *GraphQLClient) Health(ctx context.Context) error {
	return fmt.Errorf("not implemented: %s", "Health")
}

func (c *GraphQLClient) GetIotaClient() *GraphQLClient {
	return c
}

func (c *GraphQLClient) DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error) {
	return iotago.PackageID{}, fmt.Errorf("not implemented: %s", "DeployISCContracts")
}

func (c *GraphQLClient) FindCoinsForGasPayment(
	ctx context.Context,
	owner *iotago.Address,
	pt iotago.ProgrammableTransaction,
	gasBudget uint64,
) ([]*iotago.ObjectRef, error) {
	coinType := IotaCoinType
	coinPage, err := c.GetCoins(
		ctx, GetCoinsRequest{
			CoinType: &coinType,
			Owner:    owner,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch coins for gas payment: %w", err)
	}
	gasPayments, err := PickupCoinsWithFilter(
		Coins(coinPage.Address.Coins.Nodes),
		gasBudget,
		func(c Coin) bool {
			addr := c.ObjectID()
			return !pt.IsInInputObjects(&addr)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pickup coins for gas payment: %w", err)
	}

	return gasPayments.CoinRefs()
}

func (c *GraphQLClient) SignAndExecuteTxWithRetry(
	ctx context.Context,
	signer iotasigner.Signer,
	pt iotago.ProgrammableTransaction,
	gasCoin *iotago.ObjectRef,
	gasBudget uint64,
	gasPrice uint64,
) (*ExecuteTransactionBlockResponse, error) {
	var err error
	var txnBytes []byte
	var txnResponse *graphqltypes.ExecuteTransactionBlockResponse
	var gasPayments []*iotago.ObjectRef
	var updatedGasCoin *iotago.ObjectRef
	for i := 0; i < 5; i++ {
		if gasCoin == nil {
			gasPayments, err = c.FindCoinsForGasPayment(ctx, signer.Address(), pt, gasBudget)
			if err != nil {
				return nil, fmt.Errorf("failed to find gas payment: %w", err)
			}
		} else {
			updatedGasCoin, err = c.UpdateObjectRef(ctx, gasCoin)
			if err != nil {
				return nil, fmt.Errorf("failed to update gas payment: %w", err)
			}
			if updatedGasCoin.Version > gasCoin.Version {
				gasCoin = updatedGasCoin
			}
			gasPayments = []*iotago.ObjectRef{gasCoin}
		}

		tx := iotago.NewProgrammable(
			signer.Address(),
			pt,
			gasPayments,
			gasBudget,
			gasPrice,
		)
		txnBytes, err = bcs.Marshal(&tx)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tx: %w", err)
		}

		txnResponse, err = c.SignAndExecuteTransaction(ctx, txnBytes, signer)
		if err == nil {
			return txnResponse, nil
		}
		time.Sleep(c.tickingTime)
	}
	return nil, fmt.Errorf("can't execute the transaction in time: %w", err)
}

func convertGraphQLBalance(coinTypeRepr string, coinObjectCount uint64, totalBalance BigInt) (*Balance, error) {
	// When coinTypeRepr is empty, default to IOTA coin type (matching GraphQL query default)
	if coinTypeRepr == "" {
		coinTypeRepr = "0x2::iota::IOTA"
	}
	coinType, err := CoinTypeFromString(coinTypeRepr)
	if err != nil {
		return nil, fmt.Errorf("invalid coin type %s: %w", coinTypeRepr, err)
	}

	// Handle nil totalBalance by defaulting to zero
	totalBalancePtr := NewBigInt(0)
	if totalBalance.Int != nil {
		totalBalancePtr = totalBalance.Clone()
	}

	return &Balance{
		CoinType:        coinType,
		CoinObjectCount: NewBigInt(coinObjectCount),
		TotalBalance:    totalBalancePtr,
	}, nil
}
