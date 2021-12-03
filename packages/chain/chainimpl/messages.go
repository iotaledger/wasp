// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

// DismissChainMsg sent by component to the chain core in case of major setback
type DismissChainMsg struct {
	Reason string
}

type OffLedgerRequestMsg struct {
	Req         *iscp.OffLedgerRequestData
	SenderNetID string
}

type RequestAckMsg struct {
	ReqID       *iscp.RequestID
	SenderNetID string
}
