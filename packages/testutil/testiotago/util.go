package testiotago

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func RandNativeTokenID() (ret iotago.NativeTokenID) {
	copy(ret[:], tpkg.RandBytes(len(ret)))
	return
}

func RandNativeTokenAmount(amount uint64) *iotago.NativeToken {
	return &iotago.NativeToken{
		ID:     RandNativeTokenID(),
		Amount: new(big.Int).SetUint64(amount),
	}
}
