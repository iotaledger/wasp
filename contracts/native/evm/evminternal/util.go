// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evminternal

import (
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func Result(value []byte) dict.Dict {
	if value == nil {
		return nil
	}
	return dict.Dict{evm.FieldResult: value}
}
