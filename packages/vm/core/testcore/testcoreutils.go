package testcore

import (
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func EventsViewResultToStringArray(result dict.Dict) ([]string, error) {
	entries := collections.NewArray16ReadOnly(result, blocklog.ParamEvent)
	ret := make([]string, entries.MustLen())
	for i := range ret {
		data, err := entries.GetAt(uint16(i))
		if err != nil {
			return nil, err
		}
		ret[i] = string(data)
	}
	return ret, nil
}
