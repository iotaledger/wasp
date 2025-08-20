package gpa_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	TestMsgID1 gpa.MessageType = iota
	TestMsgWrapped
)

type TestMsg struct {
	A int32
	B string
}

func (m *TestMsg) MsgType() gpa.MessageType {
	return TestMsgID1
}

func (m *TestMsg) Recipient() gpa.NodeID {
	return gpa.NodeID{}
}

func (m *TestMsg) SetSender(gpa.NodeID) {
}

type WrappedMsg struct {
	C []bool
}

func (m *WrappedMsg) MsgType() gpa.MessageType {
	return TestMsgWrapped
}

func (m *WrappedMsg) Recipient() gpa.NodeID {
	return gpa.NodeID{}
}

func (m *WrappedMsg) SetSender(gpa.NodeID) {
}

func TestUnmarshalMessage(t *testing.T) {
	decodeWrapped := func(b []byte) (gpa.Message, error) {
		return bcs.Unmarshal[*WrappedMsg](b)
	}

	unmarshal := func(data []byte) (gpa.Message, error) {
		return gpa.UnmarshalMessage(data, gpa.Mapper{
			TestMsgID1: func() gpa.Message { return &TestMsg{} },
		}, gpa.Fallback{
			TestMsgWrapped: decodeWrapped,
		})
	}

	var encBuf bytes.Buffer
	enc := bcs.NewEncoder(&encBuf)
	enc.Encode(TestMsgID1)
	enc.Encode(TestMsg{A: 42, B: "hello"})
	require.NoError(t, enc.Err())

	msg, err := unmarshal(encBuf.Bytes())
	require.NoError(t, err)
	require.Equal(t, &TestMsg{A: 42, B: "hello"}, msg)

	encBuf.Reset()
	enc = bcs.NewEncoder(&encBuf)
	enc.Encode(TestMsgWrapped)
	enc.Encode(WrappedMsg{C: []bool{true, false}})
	require.NoError(t, enc.Err())

	msg, err = unmarshal(encBuf.Bytes())
	require.NoError(t, err)
	require.Equal(t, &WrappedMsg{C: []bool{true, false}}, msg)

	encBuf.Reset()
	enc = bcs.NewEncoder(&encBuf)
	enc.Encode(gpa.MessageType(123))
	enc.Encode(TestMsg{A: 42, B: "hello"})
	require.NoError(t, enc.Err())

	_, err = unmarshal(encBuf.Bytes())
	require.Error(t, err)
}
