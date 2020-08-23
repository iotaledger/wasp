package wasmpoc

type ConfigObject struct {
	vm *wasmVMPocProcessor
}

func NewConfigObject(h *wasmVMPocProcessor) *ConfigObject {
	return &ConfigObject{vm: h}
}

func (o *ConfigObject) GetInt(keyId int32) int64 {
	panic("implement Config.GetInt")
}

func (o *ConfigObject) GetObjectId(keyId int32, typeId int32) int32 {
	panic("implement Config.GetObjectId")
}

func (o *ConfigObject) GetString(keyId int32) string {
	panic("implement Config.GetString")
}

func (o *ConfigObject) SetInt(keyId int32, value int64) {
	panic("implement Config.SetInt")
}

func (o *ConfigObject) SetString(keyId int32, value string) {
	panic("implement Config.SetString")
}
