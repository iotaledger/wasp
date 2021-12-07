package wasmproc

import (
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type ScSandboxObject struct {
	ScDict
}

func (o *ScSandboxObject) DelKey(keyID, typeID int32) {
	o.Panicf("DelKey: cannot delete predefined key")
}

func (o *ScSandboxObject) GetBytes(keyID, typeID int32) []byte {
	if keyID == wasmhost.KeyLength && (o.typeID&wasmhost.OBJTYPE_ARRAY) != 0 {
		return o.ScDict.GetBytes(keyID, typeID)
	}
	o.InvalidKey(keyID)
	return nil
}

func (o *ScSandboxObject) GetObjectID(keyID, typeID int32) int32 {
	o.InvalidKey(keyID)
	return 0
}

func (o *ScSandboxObject) SetBytes(keyID, typeID int32, bytes []byte) {
	o.InvalidKey(keyID)
}
