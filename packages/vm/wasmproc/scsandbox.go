package wasmproc

import (
	"encoding/binary"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type ScSandboxObject struct {
	ScDict
	vm *wasmProcessor
}

func (o *ScSandboxObject) invalidKey(keyId int32) {
	o.Panic("Invalid key: %d", keyId)
}

func (o *ScSandboxObject) GetBytes(keyId int32, typeId int32) []byte {
	if (o.typeId&wasmhost.OBJTYPE_ARRAY) != 0 && keyId == wasmhost.KeyLength {
		bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes, uint64(o.length))
		return bytes
	}
	o.invalidKey(keyId)
	return nil
}

func (o *ScSandboxObject) GetObjectId(keyId int32, typeId int32) int32 {
	o.invalidKey(keyId)
	return 0
}

func (o *ScSandboxObject) GetString(keyId int32) string {
	o.invalidKey(keyId)
	return ""
}

func (o *ScSandboxObject) SetBytes(keyId int32, typeId int32, value []byte) {
	o.invalidKey(keyId)
}

func (o *ScSandboxObject) SetInt(keyId int32, value int64) {
	o.invalidKey(keyId)
}

func (o *ScSandboxObject) SetString(keyId int32, value string) {
	o.invalidKey(keyId)
}
