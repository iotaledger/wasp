package peering

import "time"

const (
	// equal and larger msg types are committee messages
	// those with smaller are reserved by the package for heartbeat and handshake messages
	FirstCommitteeMsgCode = byte(0x10)

	MsgTypeHeartbeat = byte(0)
	MsgTypeHandshake = byte(1)

	restartAfter = 1 * time.Second
	dialTimeout  = 1 * time.Second
	dialRetries  = 10
	backoffDelay = 500 * time.Millisecond
	// heartbeat msg
	numHeartbeatsToKeep = 5               // number of heartbeats to save for average latencyRingBuf
	heartbeatEvery      = 5 * time.Second // heartBeat period
	isDeadAfterMissing  = 2               // is dead after 4 heartbeat periods missing
)
