// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

type BalancesMsg struct {
	Balances map[valuetransaction.ID][]*balance.Balance
}

type RequestMsg struct {
	*sctransaction_old.TransactionEssence
	Index      uint16
	FreeTokens coretypes.ColoredBalancesOld
}

func (reqMsg *RequestMsg) RequestId() *coretypes.RequestID {
	ret := coretypes.NewRequestID(reqMsg.TransactionEssence.ID(), reqMsg.Index)
	return &ret
}

func (reqMsg *RequestMsg) RequestBlock() *sctransaction_old.RequestSection {
	return reqMsg.Requests()[reqMsg.Index]
}

func (reqMsg *RequestMsg) Timelock() uint32 {
	return reqMsg.RequestBlock().Timelock()
}
