package iotaclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iotaledger/hive.go/log"

	api "github.com/iotaledger/wasp/v2/clients/iota-go/client"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
)

const (
	SingleCoinFundsFromFaucetAmount = uint64(1_000_000_000)
	FundsFromFaucetAmount           = SingleCoinFundsFromFaucetAmount * 2
	DefaultGasBudget                = api.DefaultGasBudget
	DefaultGasPrice                 = api.DefaultGasPrice
	MinGasBudget                    = api.MinGasBudget
	MaxGasBudget                    = api.MaxGasBudget
)

type (
	IotaClient                        = api.IotaClient
	RetryCondition[T any]             = api.RetryCondition[T]
	WaitParams                        = api.WaitParams
	GetDynamicFieldObjectRequest      = api.GetDynamicFieldObjectRequest
	GetDynamicFieldsRequest           = api.GetDynamicFieldsRequest
	GetOwnedObjectsRequest            = api.GetOwnedObjectsRequest
	QueryEventsRequest                = api.QueryEventsRequest
	QueryTransactionBlocksRequest     = api.QueryTransactionBlocksRequest
	ResolveNameServiceNamesRequest    = api.ResolveNameServiceNamesRequest
	DevInspectTransactionBlockRequest = api.DevInspectTransactionBlockRequest
	DryRunTransactionRequest          = api.DryRunTransactionRequest
	ExecuteTransactionBlockRequest    = api.ExecuteTransactionBlockRequest
	BatchTransactionRequest           = api.BatchTransactionRequest
	MergeCoinsRequest                 = api.MergeCoinsRequest
	MoveCallRequest                   = api.MoveCallRequest
	PayRequest                        = api.PayRequest
	PayAllIotaRequest                 = api.PayAllIotaRequest
	PayIotaRequest                    = api.PayIotaRequest
	PublishRequest                    = api.PublishRequest
	RequestAddStakeRequest            = api.RequestAddStakeRequest
	RequestWithdrawStakeRequest       = api.RequestWithdrawStakeRequest
	SplitCoinRequest                  = api.SplitCoinRequest
	SplitCoinEqualRequest             = api.SplitCoinEqualRequest
	TransferObjectRequest             = api.TransferObjectRequest
	TransferIotaRequest               = api.TransferIotaRequest
	GetAllCoinsRequest                = api.GetAllCoinsRequest
	GetBalanceRequest                 = api.GetBalanceRequest
	GetCoinsRequest                   = api.GetCoinsRequest
	GetCheckpointsRequest             = api.GetCheckpointsRequest
	GetObjectRequest                  = api.GetObjectRequest
	GetTransactionBlockRequest        = api.GetTransactionBlockRequest
	MultiGetObjectsRequest            = api.MultiGetObjectsRequest
	MultiGetTransactionBlocksRequest  = api.MultiGetTransactionBlocksRequest
	TryGetPastObjectRequest           = api.TryGetPastObjectRequest
	TryMultiGetPastObjectsRequest     = api.TryMultiGetPastObjectsRequest
	SignAndExecuteTransactionRequest  = api.SignAndExecuteTransactionRequest
)

var (
	WaitForEffectsDisabled = api.WaitForEffectsDisabled
	WaitForEffectsEnabled  = api.WaitForEffectsEnabled
)

// Client wraps the GraphQL client so callers depending on the legacy iotaclient
// package path can continue to work with the new GraphQL implementation.
type Client struct {
	*iotagraphql.GraphQLClient
}

func NewClient(apiURL string, waitUntilEffectsVisible *WaitParams) *Client {
	graphqlURL := iotaconn.GraphQLURL(apiURL)
	return &Client{
		GraphQLClient: iotagraphql.NewGraphQLClientWithWaitParams(graphqlURL, waitUntilEffectsVisible),
	}
}

// NewWebsocket keeps the existing signature but currently returns an HTTP-based GraphQL client.
func NewWebsocket(ctx context.Context, wsURL string, waitUntilEffectsVisible *WaitParams, log log.Logger) (*Client, error) {
	_ = ctx
	_ = log
	return NewClient(wsURL, waitUntilEffectsVisible), nil
}

func Retry[T any](
	ctx context.Context,
	f func() (T, error),
	shouldRetry RetryCondition[T],
	params *WaitParams,
) (T, error) {
	return api.Retry(ctx, f, shouldRetry, params)
}

func RetryOnError[T any](ctx context.Context, f func() (T, error), params *WaitParams) (T, error) {
	return api.RetryOnError(ctx, f, params)
}

func DefaultRetryCondition[T any]() RetryCondition[T] {
	return api.DefaultRetryCondition[T]()
}

func UnmarshalBCS[Obj any](data []byte, obj *Obj) error {
	return api.UnmarshalBCS(data, obj)
}

// RequestFundsFromFaucet requests test funds for the provided address from the faucet endpoint.
func RequestFundsFromFaucet(ctx context.Context, address *iotago.Address, faucetURL string) error {
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

// WaitUntilStopped is a no-op placeholder to keep websocket-dependent code compiling.
func (c *Client) WaitUntilStopped() {}

// SubscribeEvent is currently unsupported on the GraphQL client.
func (c *Client) SubscribeEvent(
	ctx context.Context,
	filter *iotajsonrpc.EventFilter,
	resultCh chan<- *iotajsonrpc.IotaEvent,
) error {
	_ = ctx
	_ = filter
	_ = resultCh
	return fmt.Errorf("event subscriptions are not supported by the GraphQL client")
}

// SubscribeTransaction is currently unsupported on the GraphQL client.
func (c *Client) SubscribeTransaction(
	ctx context.Context,
	filter *iotajsonrpc.TransactionFilter,
	resultCh chan<- *serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects],
) error {
	_ = ctx
	_ = filter
	_ = resultCh
	return fmt.Errorf("transaction subscriptions are not supported by the GraphQL client")
}
