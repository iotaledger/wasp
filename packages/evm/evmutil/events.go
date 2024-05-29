package evmutil

import "github.com/ethereum/go-ethereum/common"

func AddressToIndexedTopic(addr common.Address) (ret common.Hash) {
	copy(ret[len(ret)-len(addr):], addr[:])
	return
}
