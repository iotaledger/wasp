package iotaclient

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
)

type Client struct {
	transport transport

	// If WaitUntilEffectsVisible is set, it takes effect on any sent transaction with WaitForLocalExecution. It is
	// necessary because if the L1 node is overloaded, it may return an effects cert without actually having ececuted
	// the tx locally.
	WaitUntilEffectsVisible *WaitParams
}

type WaitParams struct {
	Attempts             int
	DelayBetweenAttempts time.Duration
}

var (
	WaitForEffectsDisabled *WaitParams = nil
	WaitForEffectsEnabled  *WaitParams = &WaitParams{
		Attempts:             5,
		DelayBetweenAttempts: 2 * time.Second,
	}
)

type transport interface {
	Call(ctx context.Context, v any, method iotaconn.JsonRPCMethod, args ...any) error
	Subscribe(ctx context.Context, v chan<- []byte, method iotaconn.JsonRPCMethod, args ...any) error
	WaitUntilStopped()
}

func (c *Client) WaitUntilStopped() {
	c.transport.WaitUntilStopped()
}

type RetryCondition[T any] func(result T, err error) bool

// Retry retries a function until the condition is met or the context is cancelled
func Retry[T any](
	ctx context.Context,
	f func() (T, error),
	shouldRetry RetryCondition[T],
	params *WaitParams,
) (T, error) {
	var result T
	var err error

	// If params is nil, just run once without retrying
	if params == nil {
		return f()
	}

	for i := range params.Attempts {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		result, err = f()
		if !shouldRetry(result, err) {
			return result, nil
		}
		// no need to wait after last attempt
		if i < params.Attempts-1 {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(params.DelayBetweenAttempts):
			}
		}
	}

	// failed all attempts, but we still might return incomplete result
	return result, fmt.Errorf("retry failed after %d attempts: %v", params.Attempts, err)
}

// RetryOnError retries a function until the error is nil or the context is cancelled
func RetryOnError[T any](ctx context.Context, f func() (T, error), params *WaitParams) (T, error) {
	return Retry(ctx, f, DefaultRetryCondition[T](), params)
}

// DefaultRetryCondition returns a RetryCondition that only retries on error
func DefaultRetryCondition[T any]() RetryCondition[T] {
	return func(result T, err error) bool {
		return err != nil
	}
}
