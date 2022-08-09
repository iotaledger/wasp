package codec

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func MakeDict(vars map[string]interface{}) dict.Dict {
	ret := dict.New()
	for k, v := range vars {
		ret.Set(kv.Key(k), Encode(v))
	}
	return ret
}

func EncodeDict(value dict.Dict) []byte {
	return value.Bytes()
}

func DecodeDict(b []byte) (dict.Dict, error) {
	return dict.FromBytes(b)
}
