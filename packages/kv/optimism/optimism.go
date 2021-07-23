// optimism package implements primitives needed for te optimistic read of the chain's state
package optimism

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"golang.org/x/xerrors"
)

// OptimisticKVStoreReader implements KVReader interfaces. It wraps any kv.KVStoreReader together with the baseline
// of the state index, which, in turn, is linked to the global state index object.
// The user of the OptimisticKVStoreReader announces it starts reading the state with SetBaseline. Then the user reads
// the state with Get's. If sequense of Get's is finished without error, it means state wasn't invalidated from
// the baseline until the last read.
// If returned error is ErrStateHasBeenInvalidated ir means state was invalidated since SetBaseline.
// In this case read can be repeated. Any other error is a database error
type OptimisticKVStoreReader struct {
	kvstore  kv.KVStoreReader
	baseline coreutil.StateBaseline
}

type ErrorStateInvalidated struct {
	error
}

var ErrStateHasBeenInvalidated = &ErrorStateInvalidated{xerrors.New("virtual state has been invalidated")}

// NewOptimisticKVStoreReader creates an instance of the optimistic reader
func NewOptimisticKVStoreReader(store kv.KVStoreReader, baseline coreutil.StateBaseline) *OptimisticKVStoreReader {
	ret := &OptimisticKVStoreReader{
		kvstore:  store,
		baseline: baseline,
	}
	ret.SetBaseline()
	return ret
}

// SetBaseline sets the baseline for the read.
// Each and check if it wasn't invalidated by the global variable (the state manager)
func (o *OptimisticKVStoreReader) SetBaseline() {
	o.baseline.Set()
}

// IsStateValid check the validity of the baseline
func (o *OptimisticKVStoreReader) IsStateValid() bool {
	return o.baseline.IsValid()
}

func (o *OptimisticKVStoreReader) Get(key kv.Key) ([]byte, error) {
	println("!!!!!!!!!")
	println(o.baseline)
	if !o.baseline.IsValid() {
		return nil, ErrStateHasBeenInvalidated
	}
	ret, err := o.kvstore.Get(key)
	if err != nil {
		return nil, err
	}
	if !o.baseline.IsValid() {
		return nil, ErrStateHasBeenInvalidated
	}
	return ret, nil
}

func (o *OptimisticKVStoreReader) Has(key kv.Key) (bool, error) {
	if !o.baseline.IsValid() {
		return false, ErrStateHasBeenInvalidated
	}
	ret, err := o.kvstore.Has(key)
	if err != nil {
		return false, err
	}
	if !o.baseline.IsValid() {
		return false, ErrStateHasBeenInvalidated
	}
	return ret, nil
}

func (o *OptimisticKVStoreReader) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	if err := o.kvstore.Iterate(prefix, f); err != nil {
		return err
	}
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	return nil
}

func (o *OptimisticKVStoreReader) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	if err := o.kvstore.IterateKeys(prefix, f); err != nil {
		return err
	}
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	return nil
}

func (o *OptimisticKVStoreReader) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	if err := o.kvstore.IterateSorted(prefix, f); err != nil {
		return err
	}
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	return nil
}

func (o *OptimisticKVStoreReader) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	if err := o.kvstore.IterateKeysSorted(prefix, f); err != nil {
		return err
	}
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	return nil
}

func (o *OptimisticKVStoreReader) MustGet(key kv.Key) []byte {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	ret := o.kvstore.MustGet(key)
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	return ret
}

func (o *OptimisticKVStoreReader) MustHas(key kv.Key) bool {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	ret := o.kvstore.MustHas(key)
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	return ret
}

func (o *OptimisticKVStoreReader) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	o.kvstore.MustIterate(prefix, f)
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
}

func (o *OptimisticKVStoreReader) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	o.kvstore.MustIterateKeys(prefix, f)
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
}

func (o *OptimisticKVStoreReader) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	o.kvstore.MustIterateSorted(prefix, f)
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
}

func (o *OptimisticKVStoreReader) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	o.kvstore.MustIterateKeysSorted(prefix, f)
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
}

func RetryOnStateInvalidated(fun func() error, retryDelay time.Duration, timeout time.Time) error {
	err := fun()
	if err != nil {
		if errors.Is(err, ErrStateHasBeenInvalidated) {
			if time.Now().Before(timeout) {
				time.Sleep(retryDelay)
				return RetryOnStateInvalidated(fun, retryDelay, timeout)
			}
			return fmt.Errorf("Retrying timed out. Last error: %w", err)
		}
	}
	return err
}
