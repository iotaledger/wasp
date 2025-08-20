package isc_test

import (
	"testing"
	"time"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
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

	event = &isc.Event{
		ContractID: isc.Hname(1223),
		Topic:      "this is a topic",
		Timestamp:  uint64(123456789),
		Payload:    []byte("message payload"),
	}
	bcs.TestCodecAndHash(t, event, "ac816b79c1ca")
}
