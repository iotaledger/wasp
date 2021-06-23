package wasmproc

import (
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type ScSandboxObject struct {
	ScDict
	vm *WasmProcessor //nolint:structcheck
}

func (o *ScSandboxObject) invalidKey(keyID int32) {
	o.Panic("invalid key: %d", keyID)
}

func (o *ScSandboxObject) GetBytes(keyID, typeID int32) []byte {
	if keyID == wasmhost.KeyLength && (o.typeID&wasmhost.OBJTYPE_ARRAY) != 0 {
		return o.Int64Bytes(int64(o.length))
	}
	o.invalidKey(keyID)
	return nil
}

func (o *ScSandboxObject) GetObjectID(keyID, typeID int32) int32 {
	o.invalidKey(keyID)
	return 0
}

func (o *ScSandboxObject) SetBytes(keyID, typeID int32, bytes []byte) {
	o.invalidKey(keyID)
}
