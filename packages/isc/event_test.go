package isc_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
)

func TestEventSerialize(t *testing.T) {
	eTest := isc.Event{
		ContractID: isc.Hname(1223),
		Payload:    []byte("rand message payload"),
		Topic:      "this is topic",
		Timestamp:  uint64(time.Now().Unix()),
	}

	data1 := eTest.Bytes()
	var e isc.Event
	err := e.Read(bytes.NewReader(data1))
	require.NoError(t, err)
	require.Equal(t, eTest.ContractID, e.ContractID)
	require.Equal(t, eTest.Payload, e.Payload)
	require.Equal(t, eTest.Topic, e.Topic)
	require.Equal(t, eTest.Timestamp, e.Timestamp)

	var data2 bytes.Buffer
	eTest.Write(&data2)
	require.Equal(t, data1, data2.Bytes())
}
