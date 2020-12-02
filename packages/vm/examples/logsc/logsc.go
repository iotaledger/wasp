// logsc is a smart contract that takes requests to log a message and adds it to the log
package logsc

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/plugins/publisher"
)

const ProgramHash = "4YguJ8NyyN7RtRy56XXBABY79cYMoKup7sm3YxoNB755"

var (
	RequestCodeAddLog = coret.Hn("codeAddLog")
)

type logscEntryPoint func(ctx vmtypes.Sandbox)

type logscProcessor map[coret.Hname]logscEntryPoint

var entryPoints = logscProcessor{
	RequestCodeAddLog: handleAddLogRequest,
}

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (p logscProcessor) GetEntryPoint(code coret.Hname) (vmtypes.EntryPoint, bool) {
	ep, ok := p[code]
	return ep, ok
}

func (v logscProcessor) GetDescription() string {
	return "LogSc hard coded smart contract processor"
}

func (ep logscEntryPoint) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ep(ctx)
	return nil, nil
}

// TODO
func (ep logscEntryPoint) IsView() bool {
	return false
}

// TODO
func (ep logscEntryPoint) CallView(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	panic("implement me")
}

func (v logscEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return v
}

const logArrayKey = kv.Key("log")

func handleAddLogRequest(ctx vmtypes.Sandbox) {
	params := ctx.Params()
	msg, ok, _ := params.GetString("message")
	if !ok {
		fmt.Printf("[logsc] invalid request: missing message argument")
		return
	}

	length, _ := ctx.State().GetInt64(logArrayKey)
	length += 1
	ctx.State().SetInt64(logArrayKey, length)
	ctx.State().SetString(kv.Key(fmt.Sprintf("%s:%d", logArrayKey, length-1)), msg)

	publisher.Publish("logsc-addlog", fmt.Sprintf("length=%d", length), fmt.Sprintf("msg=[%s]", msg))
}
