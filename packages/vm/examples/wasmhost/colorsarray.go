package wasmhost

type ColorsArray struct {
	ArrayObject
	requestOnly bool
	colors      []string
}

func NewColorsArray(vm *wasmVMPocProcessor) HostObject {
	return &ColorsArray{ArrayObject: ArrayObject{vm: vm, name: "Colors"}, requestOnly: false}
}

func NewColorsArrayRequest(vm *wasmVMPocProcessor) HostObject {
	return &ColorsArray{ArrayObject: ArrayObject{vm: vm, name: "Colors"}, requestOnly: true}
}

func (a *ColorsArray) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(a.GetLength())
	}
	return a.ArrayObject.GetInt(keyId)
}

func (a *ColorsArray) GetLength() int32 {
	a.loadColors()
	return int32(len(a.colors))
}

func (a *ColorsArray) GetString(keyId int32) string {
	if keyId >= 0 && keyId < a.GetLength() {
		return a.colors[keyId]
	}
	return a.ArrayObject.GetString(keyId)
}

func (a *ColorsArray) loadColors() {
	if a.colors != nil {
		return
	}
	//TODO determine valid colors for account or request and add them base58-encoded to colors array
}
