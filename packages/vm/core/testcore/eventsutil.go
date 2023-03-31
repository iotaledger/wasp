package testcore

import (
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func EventsViewResultToStringArray(result dict.Dict) []string {
	entries := collections.NewArray16ReadOnly(result, blocklog.ParamEvent)
	ret := make([]string, entries.Len())
	for i := range ret {
		data := entries.GetAt(uint16(i))
		ret[i] = string(data)
	}
	return ret
}
