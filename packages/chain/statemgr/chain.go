// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/coretypes"
)

//Chain is a subinterface of packages/chain.Chain
type Chain interface {
	ID() *coretypes.ChainID
	EventRequestProcessed() *events.Event
	ReceiveMessage(interface{})
	Dismiss()
}
