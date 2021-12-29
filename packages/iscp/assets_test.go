package iscp

import (
	"math/big"
	"testing"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

func TestMarshalling(t *testing.T) {
	maxVal, e := big.NewInt(2), big.NewInt(256)
	maxVal.Exp(maxVal, e, nil)
	maxVal.Sub(maxVal, big.NewInt(1))

	tokens := iotago.NativeTokens{
		&iotago.NativeToken{
			ID:     [iotago.NativeTokenIDLength]byte{1},
			Amount: big.NewInt(100),
		},
		&iotago.NativeToken{
			ID:     [iotago.NativeTokenIDLength]byte{2},
			Amount: big.NewInt(200),
		},
		&iotago.NativeToken{
			ID:     [iotago.NativeTokenIDLength]byte{3},
			Amount: maxVal,
		},
	}

	assets := NewAssets(1, tokens)
	bytes := assets.Bytes()
	assets2, err := AssetsFromMarshalUtil(marshalutil.New(bytes))
	require.NoError(t, err)
	require.Equal(t, assets.Iotas, assets2.Iotas)
	require.Equal(t, len(assets.Tokens), len(assets2.Tokens))
	for i := range tokens {
		require.Equal(t, assets.Tokens[i], assets2.Tokens[i])
	}
}
