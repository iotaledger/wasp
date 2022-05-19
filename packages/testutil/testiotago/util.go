package testiotago

import (
	"math/big"
	"math/rand"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func RandNativeTokenID() (ret iotago.NativeTokenID) {
	copy(ret[:], tpkg.RandBytes(len(ret)))
	return
}

func NewNativeTokenAmount(id iotago.NativeTokenID, amount uint64) *iotago.NativeToken {
	return &iotago.NativeToken{
		ID:     id,
		Amount: new(big.Int).SetUint64(amount),
	}
}

func RandNativeTokenAmount(amount uint64) *iotago.NativeToken {
	return NewNativeTokenAmount(RandNativeTokenID(), amount)
}

func RandUTXOInput() (ret iotago.UTXOInput) {
	copy(ret.TransactionID[:], tpkg.RandBytes(len(ret.TransactionID)))
	ret.TransactionOutputIndex = uint16(rand.Intn(10))
	return
}
