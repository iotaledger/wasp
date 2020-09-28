package peering

import "time"

const (
	// equal and larger msg types are committee messages
	// those with smaller are reserved by the package for heartbeat and handshake messages
	FirstCommitteeMsgCode = byte(0x10)

	MsgTypeReserved  = byte(0)
	MsgTypeHandshake = byte(1)
	MsgTypeMsgChunk  = byte(2)

	restartAfter = 1 * time.Second
	dialTimeout  = 1 * time.Second
	dialRetries  = 10
	backoffDelay = 500 * time.Millisecond
)
