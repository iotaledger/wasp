package nodeconn

import "github.com/iotaledger/hive.go/events"

var EventNodeMessageReceived *events.Event

func init() {
	EventNodeMessageReceived = events.NewEvent(events.ByteSliceCaller)
}
