package codec

import "github.com/iotaledger/wasp/v2/packages/kv"

func StateGetOr[T any](s kv.KVReader, key kv.Key, def T) T {
	b := s.Get(key)
	if b == nil {
		return def
	}

	r, err := Decode(b, def)
	if err != nil {
		panic(err)
	}

	return r
}

func StateGet[T any](s kv.KVReader, key kv.Key) (r T) {
	return StateGetOr(s, key, r)
}

func StateSet[T any](s kv.KVWriter, key kv.Key, value T) {
	s.Set(key, Encode(value))
}
