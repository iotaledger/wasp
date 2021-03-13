// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction_old"
)

type StateTransactionMsg struct {
	*sctransaction_old.TransactionEssence
}

type TransactionInclusionLevelMsg struct {
	TxId  *valuetransaction.ID
	Level byte
}

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
