// Package evmutil provides utility functions for EVM operations and interactions.
package evmutil

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
)

func AddressToIndexedTopic(addr common.Address) (ret common.Hash) {
	copy(ret[len(ret)-len(addr):], addr[:])
	return
}

func PackUint256(uint256 *big.Int) []byte {
	return lo.Must((abi.Arguments{{Type: lo.Must(abi.NewType("uint256", "", nil))}}).Pack(uint256))
}
