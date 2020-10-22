package wasmhost

type ScExports struct {
	ArrayObject
}

func (o *ScExports) SetString(keyId int32, value string) {
	o.vm.SetExport(keyId, value)
}
