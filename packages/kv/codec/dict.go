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

var Dict = NewCodecFromBCS[dict.Dict]()

func DictFromSlice(params []any) dict.Dict {
	if len(params)%2 != 0 {
		panic("DictFromSlice: len(params) % 2 != 0")
	}
	r := dict.Dict{}
	for i := 0; i < len(params)/2; i++ {
		key, ok := params[2*i].(string)
		if !ok {
			panic("DictFromSlice: string expected")
		}
		r[kv.Key(key)] = Encode(params[2*i+1])
	}
	return r
}
