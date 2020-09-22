package wasmpoc

import "github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"

type ColorsArray struct {
	ArrayObject
	requestOnly bool
	colors      []string
}

func NewColorsArray(h *wasmVMPocProcessor) interfaces.HostObject {
	return &ColorsArray{ArrayObject: ArrayObject{vm: h, name: "Colors"}, requestOnly: false}
}

func NewColorsArrayRequest(h *wasmVMPocProcessor) interfaces.HostObject {
	return &ColorsArray{ArrayObject: ArrayObject{vm: h, name: "Colors"}, requestOnly: true}
}

func (a *ColorsArray) GetInt(keyId int32) int64 {
	switch keyId {
	case interfaces.KeyLength:
		return int64(a.GetLength())
	}
	return a.GetInt(keyId)
}

func (a *ColorsArray) GetLength() int32 {
	a.loadColors()
	return int32(len(a.colors))
}

func (a *ColorsArray) GetString(keyId int32) string {
	if keyId >= 0 && keyId < a.GetLength() {
		return a.colors[keyId]
	}
	return a.GetString(keyId)
}

func (a *ColorsArray) loadColors() {
	if a.colors != nil {
		return
	}
	//TODO determine valid colors for account or request and add them to colors array
}
