// logsc is a smart contract that takes requests to log a message and adds it to the log
package logsc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/plugins/publisher"
)

const ProgramHash = "4YguJ8NyyN7RtRy56XXBABY79cYMoKup7sm3YxoNB755"

const (
	RequestCodeAddLog = coretypes.EntryPointCode(0)
)

type logscEntryPoint func(ctx vmtypes.Sandbox)

type logscProcessor map[coretypes.EntryPointCode]logscEntryPoint

var entryPoints = logscProcessor{
	RequestCodeAddLog: handleAddLogRequest,
}

func GetProcessor() vmtypes.Processor {
	return entryPoints
}

func (p logscProcessor) GetEntryPoint(code coretypes.EntryPointCode) (vmtypes.EntryPoint, bool) {
	ep, ok := p[code]
	return ep, ok
}

func (v logscProcessor) GetDescription() string {
	return "LogSc hard coded smart contract processor"
}

func (ep logscEntryPoint) Call(ctx vmtypes.Sandbox, params ...interface{}) interface{} {
	ep(ctx)
	return nil
}

func (v logscEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return v
}

const logArrayKey = kv.Key("log")

func handleAddLogRequest(ctx vmtypes.Sandbox) {
	msg, ok, _ := ctx.AccessRequest().Args().GetString("message")
	if !ok {
		fmt.Printf("[logsc] invalid request: missing message argument")
		return
	}

	// TODO: implement using tlog
	length, _ := ctx.AccessState().GetInt64(logArrayKey)
	length += 1
	ctx.AccessState().SetInt64(logArrayKey, length)
	ctx.AccessState().SetString(kv.Key(fmt.Sprintf("%s:%d", logArrayKey, length-1)), msg)

	publisher.Publish("logsc-addlog", fmt.Sprintf("length=%d", length), fmt.Sprintf("msg=[%s]", msg))
}
