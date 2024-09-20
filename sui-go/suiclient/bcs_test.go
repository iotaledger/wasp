package suiclient_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalBCS(t *testing.T) {
	v := "hello"
	vEnc := bcs.MustMarshal(&v)

	var vDec string
	err := suiclient.UnmarshalBCS(vEnc, &vDec)
	require.NoError(t, err)
	require.Equal(t, v, vDec)

	vDec = ""
	vEncWithExcess := append(vEnc, 0x1)
	err = suiclient.UnmarshalBCS(vEncWithExcess, &vDec)
	require.Error(t, err)
}
