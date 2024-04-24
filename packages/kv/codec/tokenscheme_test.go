package codec

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
)

func TestTokenSchemeSerialization(t *testing.T) {
	ts := &iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}
	enc := TokenScheme.Encode(ts)
	tsBack, err := TokenScheme.Decode(enc)
	require.NoError(t, err)
	require.EqualValues(t, ts, tsBack)
}
