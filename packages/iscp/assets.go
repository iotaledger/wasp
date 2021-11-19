package iscp

import (
	"math/big"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
)

// Assets is used as assets in the UTXO and as tokens in transfer
type Assets struct {
	Iotas  uint64
	Tokens map[iotago.NativeTokenID]*big.Int
}

func NewAssets(iotas uint64, tokens iotago.NativeTokens) *Assets {
	ret := &Assets{
		Iotas:  iotas,
		Tokens: make(map[iotago.NativeTokenID]*big.Int),
	}
	for _, token := range tokens {
		ret.Tokens[token.ID] = token.Amount
	}
	return ret
}

func (a *Assets) String() string {
	panic("not implemented")
}

func (a *Assets) Bytes() []byte {
	mu := marshalutil.New()
	a.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (a *Assets) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.WriteUint64(a.Iotas)
	for id, amount := range a.Tokens {
		mu.WriteBytes(id[:])
		amtBytes := amount.Bytes()
		mu.WriteUint8(uint8(len(amtBytes)))
		mu.Write(amount)
	}
}

// NewAssetsFromMarshalUtil assumes that the data present in mu is already trimmed (only contains the assets bytes), it will read until EOF
func NewAssetsFromMarshalUtil(mu *marshalutil.MarshalUtil) (*Assets, error) {
	ret := &Assets{
		Tokens: make(map[iotago.NativeTokenID]*big.Int),
	}
	var err error
	if ret.Iotas, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	for {
		nativeTokenIDBytes, err := mu.ReadBytes(iotago.NativeTokenIDLength)
		if err != nil {
			return nil, err
		}
		var nativeTokenID [iotago.NativeTokenIDLength]byte
		copy(nativeTokenID[:], nativeTokenIDBytes)

		tokenAmountByteLen, err := mu.ReadUint8()
		if err != nil {
			return nil, err
		}
		tokenAmountBytes, err := mu.ReadBytes(int(tokenAmountByteLen))
		if err != nil {
			return nil, err
		}
		ret.Tokens[nativeTokenID] = new(big.Int).SetBytes(tokenAmountBytes)

		isEOF, err := mu.DoneReading()
		if err != nil {
			return nil, err
		}
		if isEOF {
			return ret, nil
		}
	}
}
