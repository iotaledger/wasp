package isc_test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestEventSerialize(t *testing.T) {
	event := &isc.Event{
		ContractID: isc.Hname(1223),
		Payload:    []byte("message payload"),
		Topic:      "this is a topic",
		Timestamp:  uint64(time.Now().UnixNano()),
	}
	bcs.TestCodec(t, event)
}
