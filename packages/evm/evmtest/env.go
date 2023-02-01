// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"testing"

	"github.com/ethereum/go-ethereum/log"
)

func InitGoEthLogger(t testing.TB) {
	log.Root().SetHandler(log.FuncHandler(func(r *log.Record) error {
		if r.Lvl <= log.LvlWarn {
			t.Logf("[%s] %s", r.Lvl.AlignedString(), r.Msg)
		}
		return nil
	}))
}
