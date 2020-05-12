package dispatcher

import "github.com/iotaledger/hive.go/events"

var EventSCDataLoaded *events.Event

func init() {
	EventSCDataLoaded = events.NewEvent(events.CallbackCaller)
}
