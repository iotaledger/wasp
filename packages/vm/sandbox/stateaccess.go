package sandbox

import (
	"github.com/iotaledger/wasp/packages/kv"
)

func (s *sandbox) Has(name kv.Key) (bool, error) {
	return s.vmctx.Has(name)
}

func (s *sandbox) Iterate(prefix kv.Key, f func(key kv.Key, value []byte) bool) error {
	return s.vmctx.Iterate(prefix, f)
}

func (s *sandbox) IterateKeys(prefix kv.Key, f func(key kv.Key) bool) error {
	return s.vmctx.IterateKeys(prefix, f)
}

func (s *sandbox) Get(name kv.Key) ([]byte, error) {
	return s.vmctx.Get(name)
}

func (s *sandbox) Del(name kv.Key) {
	s.vmctx.Del(name)
}

func (s *sandbox) Set(name kv.Key, value []byte) {
	s.vmctx.Set(name, value)
}
