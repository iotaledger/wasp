package codec

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

func TestTokenSchemeDeSeri(t *testing.T) {
	ts := &iotago.SimpleTokenScheme{}
	enc := EncodeTokenScheme(ts)
	tsBack, err := DecodeTokenScheme(enc)
	require.NoError(t, err)
	require.EqualValues(t, ts, tsBack)
}
