package codec

import (
	"math/big"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

func TestTokenSchemeDeSeri(t *testing.T) {
	ts := &iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}
	enc := EncodeTokenScheme(ts)
	tsBack, err := DecodeTokenScheme(enc)
	require.NoError(t, err)
	require.EqualValues(t, ts, tsBack)
}
