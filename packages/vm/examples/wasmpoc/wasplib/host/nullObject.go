package host

import "github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"

type nullObject struct {
	ctx interfaces.HostInterface
}

func NewNullObject(h interfaces.HostInterface) interfaces.HostObject {
	return &nullObject{ctx: h}
}

func (n *nullObject) error(text string) {
	n.ctx.SetString(1, interfaces.KeyError, text)
}

func (n *nullObject) GetInt(keyId int32) int64 {
	n.error("Null.GetInt")
	return 0
}

func (n *nullObject) GetObjectId(keyId int32, typeId int32) int32 {
	n.error("Null.GetObjectId")
	return 0
}

func (n *nullObject) GetString(keyId int32) string {
	n.error("Null.GetString")
	return ""
}

func (n *nullObject) SetInt(keyId int32, value int64) {
	n.error("Null.SetInt")
}

func (n *nullObject) SetString(keyId int32, value string) {
	n.error("Null.SetString")
}
