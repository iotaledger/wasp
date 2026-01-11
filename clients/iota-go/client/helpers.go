package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	bcs "github.com/iotaledger/bcs-go"
)

const (
	DefaultGasBudget = 10_000_000
	DefaultGasPrice  = 1000
	MinGasBudget     = 1_000_000
	MaxGasBudget     = 50_000_000_000
)

type WaitParams struct {
	Attempts             int
	DelayBetweenAttempts time.Duration
}

var WaitForEffectsDisabled *WaitParams = nil
var WaitForEffectsEnabled = &WaitParams{
	Attempts:             5,
	DelayBetweenAttempts: 2 * time.Second,
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

// UnmarshalBCS is a shortcut for bcs.Unmarshal that also verifies
// that the consumed bytes is exactly len(data).
func UnmarshalBCS[Obj any](data []byte, obj *Obj) error {
	r := bytes.NewReader(data)

	if _, err := bcs.UnmarshalStreamInto(r, obj); err != nil {
		return err
	}
	if r.Len() != 0 {
		return errors.New("excess bytes")
	}
	return nil
}
