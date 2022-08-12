// optimism package implements primitives needed for te optimistic read of the chain's state
package optimism

import (
	"errors"
	"time"

	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"golang.org/x/xerrors"
)

// OptimisticKVStoreReader implements KVReader interfaces. It wraps any kv.KVStoreReader together with the baseline
// of the state index, which, in turn, is linked to the global state index object.
// The user of the OptimisticKVStoreReader announces it starts reading the state with SetBaseline. Then the user reads
// the state with Get's. If sequense of Get's is finished without error, it means state wasn't invalidated from
// the baseline until the last read.
// If returned error is ErrorStateInvalidated ir means state was invalidated since SetBaseline.
// In this case read can be repeated. Any other error is a database error
type OptimisticKVStoreReader struct {
	kvstore  kv.KVStoreReader
	baseline coreutil.StateBaseline
}

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

func (o *OptimisticKVStoreReader) Baseline() coreutil.StateBaseline {
	return o.baseline
}

// IsStateValid check the validity of the baseline
func (o *OptimisticKVStoreReader) IsStateValid() bool {
	return o.baseline.IsValid()
}

func (o *OptimisticKVStoreReader) Get(key kv.Key) ([]byte, error) {
	if !o.baseline.IsValid() {
		return nil, coreutil.ErrorStateInvalidated
	}
	ret, err := o.kvstore.Get(key)
	if err != nil {
		return nil, err
	}
	if !o.baseline.IsValid() {
		return nil, coreutil.ErrorStateInvalidated
	}
	return ret, nil
}

func (o *OptimisticKVStoreReader) Has(key kv.Key) (bool, error) {
	if !o.baseline.IsValid() {
		return false, coreutil.ErrorStateInvalidated
	}
	ret, err := o.kvstore.Has(key)
	if err != nil {
		return false, err
	}
	if !o.baseline.IsValid() {
		return false, coreutil.ErrorStateInvalidated
	}
	return ret, nil
}

func (o *OptimisticKVStoreReader) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	if !o.baseline.IsValid() {
		return coreutil.ErrorStateInvalidated
	}
	if err := o.kvstore.Iterate(prefix, f); err != nil {
		return err
	}
	if !o.baseline.IsValid() {
		return coreutil.ErrorStateInvalidated
	}
	return nil
}

func (o *OptimisticKVStoreReader) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	if !o.baseline.IsValid() {
		return coreutil.ErrorStateInvalidated
	}
	if err := o.kvstore.IterateKeys(prefix, f); err != nil {
		return err
	}
	if !o.baseline.IsValid() {
		return coreutil.ErrorStateInvalidated
	}
	return nil
}

func (o *OptimisticKVStoreReader) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	if !o.baseline.IsValid() {
		return coreutil.ErrorStateInvalidated
	}
	if err := o.kvstore.IterateSorted(prefix, f); err != nil {
		return err
	}
	if !o.baseline.IsValid() {
		return coreutil.ErrorStateInvalidated
	}
	return nil
}

func (o *OptimisticKVStoreReader) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	if !o.baseline.IsValid() {
		return coreutil.ErrorStateInvalidated
	}
	if err := o.kvstore.IterateKeysSorted(prefix, f); err != nil {
		return err
	}
	if !o.baseline.IsValid() {
		return coreutil.ErrorStateInvalidated
	}
	return nil
}

func (o *OptimisticKVStoreReader) MustGet(key kv.Key) []byte {
	o.baseline.MustValidate()
	defer o.baseline.MustValidate()

	return o.kvstore.MustGet(key)
}

func (o *OptimisticKVStoreReader) MustHas(key kv.Key) bool {
	o.baseline.MustValidate()
	defer o.baseline.MustValidate()

	return o.kvstore.MustHas(key)
}

func (o *OptimisticKVStoreReader) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	o.baseline.MustValidate()
	defer o.baseline.MustValidate()

	o.kvstore.MustIterate(prefix, f)
}

func (o *OptimisticKVStoreReader) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	o.baseline.MustValidate()
	defer o.baseline.MustValidate()

	o.kvstore.MustIterateKeys(prefix, f)
}

func (o *OptimisticKVStoreReader) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	o.baseline.MustValidate()
	defer o.baseline.MustValidate()

	o.kvstore.MustIterateSorted(prefix, f)
}

func (o *OptimisticKVStoreReader) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	o.baseline.MustValidate()
	defer o.baseline.MustValidate()

	o.kvstore.MustIterateKeysSorted(prefix, f)
}

const (
	defaultRetryDelay   = 300 * time.Millisecond
	defaultRetryTimeout = 5 * time.Second
)

// RetryOnStateInvalidated repeats function while it returns ErrorStateInvalidated
// Optional parameters:
//   - timeouts[0] - overall timeout
//   - timeouts[1] - repeat delay
func RetryOnStateInvalidated(fun func() error, timeouts ...time.Duration) error {
	timeout := defaultRetryTimeout
	if len(timeouts) >= 1 {
		timeout = timeouts[0]
	}
	timeoutAfter := time.Now().Add(timeout)
	retryDelay := defaultRetryDelay
	if len(timeouts) >= 2 {
		retryDelay = timeouts[1]
	}

	var err error
	for err = fun(); errors.Is(err, coreutil.ErrorStateInvalidated); err = fun() {
		time.Sleep(retryDelay)
		if time.Now().After(timeoutAfter) {
			return xerrors.Errorf("optimistic read retry timeout. Last error: %w", err)
		}
	}
	return err
}
