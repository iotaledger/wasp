// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcp

import "time"

const (
	msgTypeReserved  = byte(0)
	msgTypeHandshake = byte(1)
	msgTypeMsgChunk  = byte(2)

	restartAfter = 1 * time.Second
	dialTimeout  = 1 * time.Second
	dialRetries  = 10
	backoffDelay = 500 * time.Millisecond
)
