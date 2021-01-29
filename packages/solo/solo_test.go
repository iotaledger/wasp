package solo

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPutBlobData(t *testing.T) {
	env := New(t, false, false)
	data := []byte("data-datadatadatadatadatadatadatadata")
	h := env.PutBlobDataIntoRegistry(data)
	require.EqualValues(t, h, hashing.HashData(data))

	p := sctransaction.NewRequestArgs()
	h1 := p.AddAsBlobHash("dataName", data)
	require.EqualValues(env.T, h, h1)

	sargs, ok, err := p.SolidifyRequestArguments(env.registry)
	require.NoError(env.T, err)
	require.True(env.T, ok)
	require.Len(env.T, sargs, 1)
	require.EqualValues(env.T, data, sargs.MustGet("dataName"))
}
