// logsc is a smart contract that takes requests to log a message and adds it to the log
package logsc

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"github.com/iotaledger/wasp/plugins/publisher"
)

const ProgramHash = "4YguJ8NyyN7RtRy56XXBABY79cYMoKup7sm3YxoNB755"

const (
	RequestCodeAddLog = sctransaction.RequestCode(uint16(0))
)

type logscEntryPoint func(ctx processor.Sandbox)

type logscProcessor map[sctransaction.RequestCode]logscEntryPoint

var Processor = logscProcessor{
	RequestCodeAddLog: handleAddLogRequest,
}

func New() processor.Processor {
	return Processor
}

func (p logscProcessor) GetEntryPoint(code sctransaction.RequestCode) (processor.EntryPoint, bool) {
	ep, ok := p[code]
	return ep, ok
}

func (ep logscEntryPoint) Run(ctx processor.Sandbox) {
	ep(ctx)
}

const logArrayKey = "log"

func handleAddLogRequest(ctx processor.Sandbox) {
	msg, ok := ctx.Request().GetString("message")
	if !ok {
		fmt.Printf("[logsc] invalid request: missing message argument")
		return
	}

	length, ok, err := ctx.State().GetInt64(logArrayKey)
	if err != nil {
		fmt.Printf("[logsc] %v", err)
		return
	}
	if !ok {
		length = 0
	}

	length += 1
	ctx.State().SetInt64(logArrayKey, length)
	ctx.State().SetString(fmt.Sprintf("%s:%d", logArrayKey, length-1), msg)

	publisher.Publish("logsc-addlog", fmt.Sprintf("length=%d", length), fmt.Sprintf("msg=[%s]", msg))
}
