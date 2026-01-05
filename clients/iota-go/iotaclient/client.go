package iotaclient

import (
	"context"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/client"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

// Re-export constants from iotagraphql
const (
	SingleCoinFundsFromFaucetAmount = iotagraphql.SingleCoinFundsFromFaucetAmount
	FundsFromFaucetAmount           = iotagraphql.FundsFromFaucetAmount
	DefaultGasBudget                = iotagraphql.DefaultGasBudget
	DefaultGasPrice                 = iotagraphql.DefaultGasPrice
	MinGasBudget                    = iotagraphql.MinGasBudget
	MaxGasBudget                    = iotagraphql.MaxGasBudget
)

// Re-export types from client package
type (
	RetryCondition[T any] = client.RetryCondition[T]
)

// Re-export variables from iotagraphql
var (
	WaitForEffectsDisabled = iotagraphql.WaitForEffectsDisabled
	WaitForEffectsEnabled  = iotagraphql.WaitForEffectsEnabled
)

// Re-export utility functions - these delegate to the client package
func Retry[T any](
	ctx context.Context,
	f func() (T, error),
	shouldRetry RetryCondition[T],
	params *iotagraphql.WaitParams,
) (T, error) {
	return client.Retry(ctx, f, shouldRetry, params)
}

func DefaultRetryCondition[T any]() RetryCondition[T] {
	return client.DefaultRetryCondition[T]()
}

func UnmarshalBCS[Obj any](data []byte, obj *Obj) error {
	return client.UnmarshalBCS(data, obj)
}

// Re-export non-generic functions from iotagraphql
var RequestFundsFromFaucet = iotagraphql.RequestFundsFromFaucet

// Client wraps the GraphQL client so callers depending on the legacy iotaclient
// package path can continue to work with the new GraphQL implementation.
type Client struct {
	*iotagraphql.GraphQLClient
}

func NewClient(apiURL string, waitUntilEffectsVisible *iotagraphql.WaitParams) *Client {
	graphqlURL := iotaconn.GraphQLURL(apiURL)
	return &Client{
		GraphQLClient: iotagraphql.NewGraphQLClientWithWaitParams(graphqlURL, waitUntilEffectsVisible),
	}
}

// NewWebsocket keeps the existing signature but currently returns an HTTP-based GraphQL client.
func NewWebsocket(ctx context.Context, wsURL string, waitUntilEffectsVisible *iotagraphql.WaitParams, log log.Logger) (*Client, error) {
	_ = ctx
	_ = log
	return NewClient(wsURL, waitUntilEffectsVisible), nil
}

// WaitUntilStopped is a no-op placeholder to keep websocket-dependent code compiling.
func (c *Client) WaitUntilStopped() {}

// SubscribeEvent is currently unsupported on the GraphQL client.
func (c *Client) SubscribeEvent(
	ctx context.Context,
	filter *graphqltypes.EventFilter,
	resultCh chan<- *graphqltypes.IotaEvent,
) error {
	_ = ctx
	_ = filter
	_ = resultCh
	return nil
}

// SubscribeTransaction is currently unsupported on the GraphQL client.
func (c *Client) SubscribeTransaction(
	ctx context.Context,
	filter *graphqltypes.TransactionFilter,
	resultCh chan<- *serialization.TagJson[graphqltypes.IotaTransactionBlockEffects],
) error {
	_ = ctx
	_ = filter
	_ = resultCh
	return nil
}
