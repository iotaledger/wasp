package optimism

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"golang.org/x/xerrors"
)

type OptimisticKVStoreReader struct {
	kvstore  kv.KVStoreReader
	baseline *coreutil.SolidStateBaseline
}

type ErrorStateInvalidated struct {
	error
}

var ErrStateHasBeenInvalidated = &ErrorStateInvalidated{xerrors.New("virtual state has been invalidated")}

// NewOptimisticKVStoreReader creates an instance of the optimistic reader on top of the KV store reader
// Each instance contain own read baseline. It is guaranteed that from the moment of calling SetBaseline
// to the last operation without errors returned the state wasn't modified
// If state is modified, the KV store operation will return ErrStateHasBeenInvalidated
func NewOptimisticKVStoreReader(store kv.KVStoreReader, baseline *coreutil.SolidStateBaseline) *OptimisticKVStoreReader {
	ret := &OptimisticKVStoreReader{
		kvstore:  store,
		baseline: baseline,
	}
	ret.SetBaseline()
	return ret
}

func (o *OptimisticKVStoreReader) SetBaseline() {
	o.baseline.SetBaseline()
}

func (o *OptimisticKVStoreReader) IsStateValid() bool {
	return o.baseline.IsValid()
}

func (o *OptimisticKVStoreReader) Get(key kv.Key) ([]byte, error) {
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

func RepeatIfUnlucky(f func() error) error {
	err := f()
	if err == ErrStateHasBeenInvalidated {
		err = f()
	}
	return err
}
