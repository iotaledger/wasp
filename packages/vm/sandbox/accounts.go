// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (s *sandbox) IncomingTransfer() coretypes.ColoredBalances {
	return s.vmctx.GetIncoming()
}

func (s *sandbox) Balance(col balance.Color) int64 {
	return s.vmctx.GetBalance(col)
}

func (s *sandbox) Balances() coretypes.ColoredBalances {
	return s.vmctx.GetMyBalances()
}
