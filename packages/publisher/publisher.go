package publisher

import (
	"github.com/iotaledger/hive.go/core/events"
)

var Event = events.NewEvent(func(handler interface{}, params ...interface{}) {
	callback := handler.(func(msgType string, parts []string))
	msgType := params[0].(string)
	parts := params[1].([]string)
	callback(msgType, parts)
})

func Publish(msgType string, parts ...string) {
	Event.Trigger(msgType, parts)
}
