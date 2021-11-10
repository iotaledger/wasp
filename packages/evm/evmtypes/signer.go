// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

func Signer(chainID *big.Int) types.Signer {
	return types.NewEIP155Signer(chainID)
}
