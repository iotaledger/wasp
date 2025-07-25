// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

type inputL1Info struct {
	gasCoins []*coin.CoinWithRef
	l1params *parameters.L1Params
}

func NewInputL1Info(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.Input {
	return &inputL1Info{gasCoins: gasCoins, l1params: l1params}
}

func (inp *inputL1Info) String() string {
	return fmt.Sprintf("{cons.inputL1Info: coins: %s, l1params=%s}", inp.gasCoins, inp.l1params)
}
