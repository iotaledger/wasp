package codec

import (
	"github.com/ethereum/go-ethereum/common"
)

var EthereumAddress = NewCodecFromBCS[common.Address]()
