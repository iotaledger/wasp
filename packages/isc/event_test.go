package isc_test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestEventSerialize(t *testing.T) {
	event := &isc.Event{
		ContractID: isc.Hname(1223),
		Topic:      "this is a topic",
		Timestamp:  uint64(time.Now().UnixNano()),
		Payload:    []byte("message payload"),
	}
	bcs.TestCodec(t, event)
	rwutil.BytesTest(t, event, func(data []byte) (*isc.Event, error) {
		return bcs.Unmarshal[*isc.Event](data)
	})
}
