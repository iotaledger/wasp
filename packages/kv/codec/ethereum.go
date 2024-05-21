package codec

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var EthereumAddress = NewCodecEx(decodeEthereumAddress)

func decodeEthereumAddress(b []byte) (common.Address, error) {
	if len(b) != common.AddressLength {
		return common.Address{}, errors.New("decodeEthereumAddress: invalid length")
	}
	return common.BytesToAddress(b), nil
}
