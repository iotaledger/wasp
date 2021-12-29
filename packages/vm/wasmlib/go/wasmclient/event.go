package wasmclient

import (
	"strconv"
)

type Event struct {
	message   []string
	Timestamp uint32
}

func (e *Event) Init(message []string) {
	e.message = message
	e.Timestamp = e.NextUint32()
}

func (e *Event) next() string {
	next := e.message[0]
	e.message = e.message[1:]
	return next
}

func (e *Event) NextAddress() Address {
	return Address(e.next())
}

func (e *Event) NextAgentID() AgentID {
	return AgentID(e.next())
}

func (e *Event) NextBytes() []byte {
	return Base58Decode(e.next())
}

func (e *Event) NextBool() bool {
	return e.next() != "0"
}

func (e *Event) NextChainID() ChainID {
	return ChainID(e.next())
}

func (e *Event) NextColor() Color {
	return Color(e.next())
}

func (e *Event) NextHash() Hash {
	return Hash(e.next())
}

func (e *Event) NextHname() Hname {
	return Hname(e.nextUint(32))
}

func (e *Event) nextInt(bitSize int) int64 {
	val, err := strconv.ParseInt(e.next(), 10, bitSize)
	if err != nil {
		panic("int parse error")
	}
	return val
}

func (e *Event) NextInt8() int8 {
	return int8(e.nextInt(8))
}

func (e *Event) NextInt16() int16 {
	return int16(e.nextInt(16))
}

func (e *Event) NextInt32() int32 {
	return int32(e.nextInt(32))
}

func (e *Event) NextInt64() int64 {
	return e.nextInt(64)
}

func (e *Event) NextRequestID() RequestID {
	return RequestID(e.next())
}

func (e *Event) NextString() string {
	return e.next()
}

func (e *Event) nextUint(bitSize int) uint64 {
	val, err := strconv.ParseUint(e.next(), 10, bitSize)
	if err != nil {
		panic("uint parse error")
	}
	return val
}

func (e *Event) NextUint8() uint8 {
	return uint8(e.nextUint(8))
}

func (e *Event) NextUint16() uint16 {
	return uint16(e.nextUint(16))
}

func (e *Event) NextUint32() uint32 {
	return uint32(e.nextUint(32))
}

func (e *Event) NextUint64() uint64 {
	return e.nextUint(64)
}
