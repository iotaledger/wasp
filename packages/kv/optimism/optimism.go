package optimism

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"golang.org/x/xerrors"
)

type OptimisticKVStoreReader struct {
	kvstore  kv.KVStoreReader
	baseline *coreutil.StateIndexBaseline
}

type ErrorStateInvalidated struct {
	error
}

var ErrStateHasBeenInvalidated = &ErrorStateInvalidated{xerrors.New("virtual state has been invalidated")}

func NewOptimisticKVStoreReader(store kv.KVStoreReader, baseline *coreutil.StateIndexBaseline) *OptimisticKVStoreReader {
	return &OptimisticKVStoreReader{
		kvstore:  store,
		baseline: baseline,
	}
}

func (o *OptimisticKVStoreReader) SetBaseline() {
	o.baseline.SetBaseline()
}

func (o *OptimisticKVStoreReader) Get(key kv.Key) ([]byte, error) {
	if !o.baseline.IsValid() {
		return nil, ErrStateHasBeenInvalidated
	}
	return o.kvstore.Get(key)
}

func (o *OptimisticKVStoreReader) Has(key kv.Key) (bool, error) {
	if !o.baseline.IsValid() {
		return false, ErrStateHasBeenInvalidated
	}
	return o.kvstore.Has(key)
}

func (o *OptimisticKVStoreReader) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	return o.kvstore.Iterate(prefix, f)
}

func (o *OptimisticKVStoreReader) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	return o.kvstore.IterateKeys(prefix, f)
}

func (o *OptimisticKVStoreReader) IterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	return o.kvstore.IterateSorted(prefix, f)
}

func (o *OptimisticKVStoreReader) IterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) error {
	if !o.baseline.IsValid() {
		return ErrStateHasBeenInvalidated
	}
	return o.kvstore.IterateKeysSorted(prefix, f)
}

func (o *OptimisticKVStoreReader) MustGet(key kv.Key) []byte {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	return o.kvstore.MustGet(key)
}

func (o *OptimisticKVStoreReader) MustHas(key kv.Key) bool {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	return o.kvstore.MustHas(key)
}

func (o *OptimisticKVStoreReader) MustIterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	o.kvstore.MustIterate(prefix, f)
}

func (o *OptimisticKVStoreReader) MustIterateKeys(prefix kv.Key, f func(key kv.Key) bool) {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	o.kvstore.MustIterateKeys(prefix, f)
}

func (o *OptimisticKVStoreReader) MustIterateSorted(prefix kv.Key, f func(key kv.Key, value []byte) bool) {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	o.kvstore.MustIterateSorted(prefix, f)
}

func (o *OptimisticKVStoreReader) MustIterateKeysSorted(prefix kv.Key, f func(key kv.Key) bool) {
	if !o.baseline.IsValid() {
		panic(ErrStateHasBeenInvalidated)
	}
	o.kvstore.MustIterateKeysSorted(prefix, f)
}
