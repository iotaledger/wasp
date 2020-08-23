package wasmpoc

import (
	"github.com/iotaledger/wart/host/interfaces"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type RequestObject struct {
	vm         *wasmVMPocProcessor
	address    string
	code       int64
	delay      int64
	args       kv.MustCodec
}

func NewRequestObject(h *wasmVMPocProcessor) *RequestObject {
	return &RequestObject{vm: h}
}

func (o *RequestObject) GetInt(keyId int32) int64 {
	switch keyId {
	//case interfaces.KeyReqCode:
	//	return o.code
	//case interfaces.KeyReqDelay:
	//	return o.delay
	default:
		o.vm.SetError("Invalid key")
	}
	return 0
}

func (o *RequestObject) GetObjectId(keyId int32, typeId int32) int32 {
	panic("implement Request.GetObjectId")
}

func (o *RequestObject) GetString(keyId int32) string {
	switch keyId {
	//case interfaces.KeyReqAddress:
	//	return o.address
	default:
		o.vm.SetError("Invalid key")
	}
	return ""
}

func (o *RequestObject) Send(ctx interfaces.HostInterface) {
	o.vm.Logf("REQ SEND c%d d%d a'%s'", o.code, o.delay, o.address)
	if o.address == "" {
		o.vm.ctx.SendRequestToSelfWithDelay(sctransaction.RequestCode(uint16(o.code)), nil, uint32(o.delay))
	}
}

func (o *RequestObject) SetInt(keyId int32, value int64) {
	switch keyId {
	case interfaces.KeyLength:
		// clear request, tracker will still know about it
		// so maybe move it to an allocation pool for reuse
		o.address = ""
		o.code = 0
		o.delay = 0
	case KeyReqCode:
		o.code = value
	case KeyReqDelay:
		o.delay = value
	default:
		o.args.SetInt64(kv.Key(o.vm.GetKey(keyId)), value)
	}
}

func (o *RequestObject) SetString(keyId int32, value string) {
	switch keyId {
	case KeyReqAddress:
		o.address = value
	default:
		o.args.SetString(kv.Key(o.vm.GetKey(keyId)), value)
	}
}
