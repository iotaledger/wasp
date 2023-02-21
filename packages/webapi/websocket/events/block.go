package events

import (
	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/wasp/packages/publisher"
)

func (e *EventManager) BlockEventHandler() *event.Closure[*publisher.BlockApplied] {
	return event.NewClosure(func(event *publisher.BlockApplied) {

	})
}
