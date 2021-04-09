// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

//NodeConn is implemented by packages/txstream/client.Client in goshimmer
type NodeConn interface {
	RequestBacklog(addr ledgerstate.Address)
	RequestConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID)
}
