// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package wasmpoc

import (
	"fmt"
	"github.com/iotaledger/wart/host"
	"github.com/iotaledger/wart/host/interfaces/level"
	"github.com/iotaledger/wart/host/interfaces/objtype"
	"github.com/iotaledger/wart/wasm/consts/desc"
	"github.com/iotaledger/wart/wasm/executors"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	ProgramHash = "BDREf2rz36AvboHYWfWXgEUG5K8iynLDZAZwKnPBmKM9"

	RequestNop           = sctransaction.RequestCode(uint16(1))
	RequestInc           = sctransaction.RequestCode(uint16(2))
	RequestIncRepeat1    = sctransaction.RequestCode(uint16(3))
	RequestIncRepeatMany = sctransaction.RequestCode(uint16(4))
	RequestPlaceBet      = sctransaction.RequestCode(uint16(5))
	RequestLockBets      = sctransaction.RequestCode(uint16(6))
	RequestPayWinners    = sctransaction.RequestCode(uint16(7))
	RequestPlayPeriod    = sctransaction.RequestCode(uint16(8) | sctransaction.RequestCodeProtected)
	RequestTokenMint     = sctransaction.RequestCode(uint16(9))
)

var entrypoints = map[sctransaction.RequestCode]string{
	RequestNop:           "no_op",
	RequestInc:           "increment",
	RequestIncRepeat1:    "incrementRepeat1",
	RequestIncRepeatMany: "incrementRepeatMany",
	RequestPlaceBet:      "placeBet",
	RequestLockBets:      "lockBets",
	RequestPayWinners:    "payWinners",
	RequestPlayPeriod:    "playPeriod",
	RequestTokenMint:     "tokenMint",
}

func init() {
	host.CreateRustAdapter()
}

type wasmVMPocProcessor struct {
	host.HostBase
	ctx        vmtypes.Sandbox
	entrypoint string
}

func GetProcessor() vmtypes.Processor {
	return &wasmVMPocProcessor{}
}

func (h *wasmVMPocProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	//TODO converts request code into Wasm code entry point name
	// needs to be changed to use entry point name instead of code
	// we don't want to burden the SC creator with extra work
	entrypoint, ok := entrypoints[code]
	if !ok {
		return nil, false
	}
	h.entrypoint = entrypoint
	return h, true
}

func (h *wasmVMPocProcessor) Run(ctx vmtypes.Sandbox) {
	reqId := ctx.AccessRequest().ID()
	ctx.Publish(fmt.Sprintf("run wasmVMPocProcessor: reqCode = %s reqId = %s timestamp = %d",
		ctx.AccessRequest().Code().String(), reqId.String(), ctx.GetTimestamp()))

	//TODO check what caching optimizations we can do to prevent
	// rebuilding entire admin from scratch on every request
	NewRootObject(h, ctx)

	//TODO for now load Wasm code from hardcoded location
	// in the future we will need to change things so
	// that we locate the code by hash and entrypoint
	// by name instead of request code number
	runner := executors.NewWasmRunner(h)
	err := runner.Load("D:\\Work\\Go\\src\\github.com\\iotaledger\\wasp\\tools\\cluster\\tests\\wasptest\\wasmtest_bg.wasm")
	if err != nil {
		ctx.Publish("error loading wasm: " + err.Error())
		return
	}

	module := runner.Module()
	for _, export := range module.Exports {
		if export.ImportName == h.entrypoint {
			if export.ExternalType != desc.FUNC {
				ctx.Publish("error running wasm: wrong export type")
				return
			}
			function := module.Functions[export.Index]
			err = runner.RunFunction(function)
			if err != nil {
				ctx.Publish("error running wasm: " + err.Error())
				return
			}
			break
		}
	}

	if h.HasError() {
		return
	}

	ctx.Publish("Processing transfers...")
	transfersId := h.GetObjectId(1, KeyTransfers, objtype.OBJTYPE_MAP_ARRAY)
	transfers := h.GetObject(transfersId).(*TransfersArray)
	for i := int32(0); i < transfers.GetLength(); i++ {
		transferId := h.GetObjectId(transfersId, i, objtype.OBJTYPE_MAP)
		transfer := h.GetObject(transferId).(*TransferObject)
		transfer.Send(h)
	}

	ctx.Publish("Processing requests...")
	requestsId := h.GetObjectId(1, KeyRequests, objtype.OBJTYPE_MAP_ARRAY)
	requests := h.GetObject(requestsId).(*RequestsArray)
	for i := int32(0); i < requests.GetLength(); i++ {
		requestId := h.GetObjectId(requestsId, i, objtype.OBJTYPE_MAP)
		request := h.GetObject(requestId).(*RequestObject)
		request.Send(h)
	}
}

func (h *wasmVMPocProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return h
}

func (h *wasmVMPocProcessor) Log(logLevel int, text string) {
	if logLevel >= level.TRACE {
		h.ctx.Publish(text)
	}
}
