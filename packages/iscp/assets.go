package iscp

import (
	"fmt"
	"math/big"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// Assets is used as assets in the UTXO and as tokens in transfer
type Assets struct {
	Iotas  uint64
	Tokens iotago.NativeTokens
}

var IOTA_TOKEN_ID = []byte{0}

func NewEmptyAssets() *Assets {
	return &Assets{
		Tokens: make([]*iotago.NativeToken, 0),
	}
}

func NewAssets(iotas uint64, tokens iotago.NativeTokens) *Assets {
	return &Assets{
		Iotas:  iotas,
		Tokens: tokens,
	}
}

func NewAssetsFromDict(d dict.Dict) (*Assets, error) {
	ret := NewEmptyAssets()
	var err error
	ret.Iotas, err = codec.DecodeUint64(d.MustGet(kv.Key(IOTA_TOKEN_ID)))
	if err != nil {
		return ret, err
	}
	d.Del(kv.Key(IOTA_TOKEN_ID))
	for key, val := range d {
		var tokenID [iotago.NativeTokenIDLength]byte
		copy(tokenID[:], []byte(key))
		token := &iotago.NativeToken{
			ID:     tokenID,
			Amount: new(big.Int).SetBytes(val),
		}
		ret.Tokens = append(ret.Tokens, token)
	}
	return ret, nil
}

func (a *Assets) String() string {
	panic("not implemented")
}

func (a *Assets) Bytes() []byte {
	mu := marshalutil.New()
	a.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (a *Assets) Equals(b *Assets) bool {
	if a.Iotas != b.Iotas {
		return false
	}
	if len(a.Tokens) != len(b.Tokens) {
		return false
	}
	bTokensSet := b.Tokens.MustSet()
	for _, token := range a.Tokens {
		if token.Amount.Cmp(bTokensSet[token.ID].Amount) != 0 {
			return false
		}
	}
	return true
}

func (a *Assets) Add(b *Assets) {
	a.Iotas += b.Iotas
	resultTokens := a.Tokens.MustSet()
	for _, token := range b.Tokens {
		if resultTokens[token.ID] != nil {
			resultTokens[token.ID].Amount.Add(
				resultTokens[token.ID].Amount,
				token.Amount,
			)
			continue
		}
		resultTokens[token.ID] = token
	}
	a.Tokens = nativeTokensFromSet(resultTokens)
}

func (a *Assets) ToDict() dict.Dict {
	ret := dict.New()
	ret.Set(kv.Key(IOTA_TOKEN_ID), codec.EncodeUint64(a.Iotas))
	for _, token := range a.Tokens {
		ret.Set(kv.Key(token.ID[:]), token.Amount.Bytes())
	}
	return ret
}

func nativeTokensFromSet(set iotago.NativeTokensSet) iotago.NativeTokens {
	ret := make(iotago.NativeTokens, len(set))
	i := 0
	for _, token := range set {
		ret[i] = token
		i++
	}
	return ret
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
