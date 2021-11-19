package iscp

import (
	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
)

// Assets is used as assets in the UTXO and as tokens in transfer
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

func (a *Assets) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.WriteUint64(a.Iotas)
	// tokenBytes, err := serializer.NewSerializer().WriteSliceOfObjects(&a.Tokens, serializer.DeSeriModePerformLexicalOrdering, nil, serializer.SeriLengthPrefixTypeAsUint16, nativeTokensArrayRules, func(err error) error {
	// 	return fmt.Errorf("unable to serialize alias output native tokens: %w", err)
	// }).Serialize()
	// TODO this isn't complete, we're missing some stuff from iotago
	panic("not implemented")
}

func NewAssetsFromMarshalUtil(mu *marshalutil.MarshalUtil) (*Assets, error) {
	ret := &Assets{}
	var err error
	if ret.Iotas, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	// TODO this isn't complete, we're missing some stuff from iotago
	panic("not implemented")
}
