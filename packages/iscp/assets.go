package iscp

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v3"
)

// Assets is used as assets in the UTXO and as tokens in Transfer
type Assets struct {
	Iotas  uint64
	Tokens iotago.NativeTokens
}

func NewAssets(iotas uint64, tokens iotago.NativeTokens) *Assets {
	return &Assets{
		Iotas:  iotas,
		Tokens: tokens,
	}
}

func (a *Assets) String() string {
	panic("not implemented")
}

func (a *Assets) Bytes() []byte {
	mu := marshalutil.New()
	a.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

var NativeAssetsSerializationArrayRules = iotago.NativeTokenArrayRules()

func (a *Assets) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.WriteUint64(a.Iotas)
	tokenBytes, err := serializer.NewSerializer().WriteSliceOfObjects(&a.Tokens, serializer.DeSeriModePerformLexicalOrdering, nil, serializer.SeriLengthPrefixTypeAsUint16, &NativeAssetsSerializationArrayRules, func(err error) error {
		return fmt.Errorf("unable to serialize alias output native tokens: %w", err)
	}).Serialize()
	if err != nil {
		panic(fmt.Errorf("unexpected error serializing native tokens: %w", err))
	}
	mu.WriteUint16(uint16(len(tokenBytes)))
	mu.WriteBytes(tokenBytes)
}

// ToMap creates respective map by summing up repetitive token IDs
func FindNativeTokenBalance(nts iotago.NativeTokens, id *iotago.NativeTokenID) *big.Int {
	for _, nt := range nts {
		if nt.ID == *id {
			return nt.Amount
		}
	}
	return nil
}

func NewAssetsFromMarshalUtil(mu *marshalutil.MarshalUtil) (*Assets, error) {
	ret := &Assets{
		Tokens: make(iotago.NativeTokens, 0),
	}
	var err error
	if ret.Iotas, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	tokenBytesLength, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	tokenBytes, err := mu.ReadBytes(int(tokenBytesLength))
	if err != nil {
		return nil, err
	}
	_, err = serializer.NewDeserializer(tokenBytes).
		ReadSliceOfObjects(&ret.Tokens, serializer.DeSeriModePerformLexicalOrdering, nil, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, &NativeAssetsSerializationArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for alias output: %w", err)
		}).Done()
	if err != nil {
		return nil, err
	}
	return ret, nil
}
