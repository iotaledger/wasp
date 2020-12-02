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
