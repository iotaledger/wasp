package solo

import (
	"testing"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/stretchr/testify/require"
)

func TestPutBlobData(t *testing.T) {
	env := New(t, false, false)
	data := []byte("data-datadatadatadatadatadatadatadata")
	h := env.PutBlobDataIntoRegistry(data)
	require.EqualValues(t, h, hashing.HashData(data))

	p := requestargs.New(nil)
	h1 := p.AddAsBlobRef("dataName", data)
	require.EqualValues(env.T, h, h1)

	sargs, ok, err := p.SolidifyRequestArguments(env.blobCache)
	require.NoError(env.T, err)
	require.True(env.T, ok)
	require.Len(env.T, sargs, 1)
	require.EqualValues(env.T, data, sargs.MustGet("dataName"))
}
