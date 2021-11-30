// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Provides implementations for chain.ChainMetrics methods
package chainimpl

import (
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
)

func (c *chainObj) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics {
	return c.nodeConn.GetMetrics()
}
