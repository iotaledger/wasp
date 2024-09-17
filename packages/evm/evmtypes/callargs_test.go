package evmtypes_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestCallMsgCodec(t *testing.T) {
	bcs.TestCodec(t, ethereum.CallMsg{
		From:  common.Address{},
		To:    &common.Address{},
		Gas:   100,
		Data:  []byte{1, 2, 3, 4},
		Value: big.NewInt(100),
	})

	bcs.TestCodec(t, ethereum.CallMsg{
		From:  common.Address{},
		Gas:   100,
		Data:  []byte{1, 2, 3, 4},
		Value: big.NewInt(100),
	})
}
