package evmtypes_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	bcs "github.com/iotaledger/bcs-go"
)

func TestCallMsgCodec(t *testing.T) {
	bcs.TestCodec(t, ethereum.CallMsg{
		From:  common.Address{1, 2, 3},
		To:    &common.Address{4, 5, 6},
		Gas:   100,
		Data:  []byte{1, 2, 3, 4},
		Value: big.NewInt(100),
	})

	bcs.TestCodec(t, ethereum.CallMsg{
		From:  common.Address{1, 2, 3},
		Gas:   100,
		Data:  []byte{1, 2, 3, 4},
		Value: big.NewInt(100),
	})
}
