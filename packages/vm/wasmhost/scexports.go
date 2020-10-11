package wasmhost

type ScExports struct {
	ArrayObject
}

func (o *ScExports) SetString(keyId int32, value string) {
	_, ok := o.vm.codeToFunc[keyId]
	if ok {
		o.error("SetString: duplicate code")
	}
	_, ok = o.vm.funcToCode[value]
	if ok {
		o.error("SetString: duplicate function")
	}
	o.vm.funcToCode[value] = keyId
	o.vm.codeToFunc[keyId] = value
}
