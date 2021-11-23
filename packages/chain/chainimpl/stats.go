// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Provides implementations for chain.ChainStats methods
package chainimpl

import (
	"github.com/iotaledger/wasp/packages/chain"
)

func (c *chainObj) GetNodeConnectionStats() chain.NodeConnectionMessagesStats {
	return c.nodeConn.GetStats()
}
