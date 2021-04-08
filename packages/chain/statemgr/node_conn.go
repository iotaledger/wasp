// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

type NodeConn interface {
	RequestBacklog(addr ledgerstate.Address)
}
