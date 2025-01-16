// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputGasInfo struct {
	gasCoins []*coin.CoinWithRef
	gasPrice uint64
}

func NewInputGasInfo(gasCoins []*coin.CoinWithRef, gasPrice uint64) gpa.Input {
	return &inputGasInfo{gasCoins: gasCoins, gasPrice: gasPrice}
}

func (inp *inputGasInfo) String() string {
	return fmt.Sprintf("{cons.inputGasInfo: coins: %v, price=%d}", inp.gasCoins, inp.gasPrice)
}
