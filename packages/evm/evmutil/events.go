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

func ERC721TokenIDToIndexedTopic(tokenID *big.Int) (ret common.Hash) {
	tokenIDPacked := PackUint256(tokenID)
	if len(tokenIDPacked) != len(ret) {
		panic("expected same length")
	}
	copy(ret[:], tokenIDPacked) // same len
	return
}

func PackUint256(uint256 *big.Int) []byte {
	return lo.Must((abi.Arguments{{Type: lo.Must(abi.NewType("uint256", "", nil))}}).Pack(uint256))
}
